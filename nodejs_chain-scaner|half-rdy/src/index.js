'use strict';

import timeout from 'connect-timeout'
import 'core-js/stable';

import express from 'express'
import { HashMap } from 'hashmap'
import { MongoClient } from 'mongodb';
import 'regenerator-runtime/runtime';
import api from './api'

const app = express();

app.use(
  express.urlencoded({
                       extended: true
                     })
);

app.use(express.json());
app.use(timeout('5s'));

const mongoClient = new MongoClient('mongodb://localhost:27017', {useUnifiedTopology: true});
let col           = undefined;
let dbGlobal      = undefined;

function DelayWork() {
  setTimeout(check_work, 10000)
}

function check(pool, block) {

}

function save(prev, walletCursor, txCursor) {
  if (prev === undefined) {
    prev = {}
  }
  let next = prev;
  
  let w = undefined;
  do {
    w = walletCursor.next();
    if (w !== undefined) {
      if (w.amount > 100) {
      
      }
      
    }
  } while (w !== undefined);
  
  let tx = undefined;
  do {
    tx = txCursor.next();
  } while (tx !== undefined);
  
  return next
}

async function check_work() {
  let workList = await col.find({$and: [{stop: {$ne: true}}, {$where: 'this.end > this.last_end'}]}).limit(1).toArray();
  
  console.log('workList', workList);
  
  if (workList.length) {
    let work = workList[0];
    let c    = dbGlobal.collection('chain-' + work.name);
    if (work.last_end < work.start) {
      work.last_end = Math.max(work.start - 1, 0);
    }
    let nextSpan = work.last_end + work.chart_time_frame;
    
    nextSpan += work.chart_time_frame - nextSpan % work.chart_time_frame;
    let end       = work.end < nextSpan ? work.end : nextSpan;
    let startSpan = end - work.chart_time_frame;
    
    let blocks = await c.find({time: {$gt: work.last_end, $lte: end}}).sort({time: 1}).toArray();
    
    if (blocks.length === 0) {
      await col.updateOne({_id: work._id}, {$set: {last_end: work.end}});
      return DelayWork();
    }
    //ToDo
    
    let addrList = new HashMap();
    //getAllAddressList
    blocks.forEach(b => {
      b.tx.forEach(tx => {
        addrList.set(tx.from, {amount: 0});
        addrList.set(tx.to, {amount: 0});
      })
    });
    
    //FetchAllAddressToPool
    let wa      = dbGlobal.collection('w-pool-' + work._id);
    let wallets = await wa.find({address: {$in: addrList.keys()}}).toArray();
    wallets.forEach(w => {
      addrList.set(w.address, w)
    });
    
    //CalcPool
    blocks.forEach(b => {
      b.tx.forEach(tx => {
        let from = addrList.get(tx.from);
        let to   = addrList.get(tx.to);
        from.amount -= tx.amount;
        to.amount += tx.amount;
        
      })
    });
    
    //makeCheck
    // потому что кошелек может получить статус 100 бтц держателя и хорошо бы точно знать в каком блоке это произошло
    // потому что размер снапа может быть неделя и больше
    
    blocks.forEach(b => check(addrList, b));
    
    //sync
    let batch = wa.initializeOrderedBulkOp();
    
    addrList.values().forEach((value) => {
      if (value._id === undefined) {
        batch.insert(value)
      } else {
        batch.find({_id: value._id}).update({$set: value})
      }
      // batch.find({time: block.time}).upsert().updateOne({'$set': block});
    });
    
    console.log(await batch.execute());
    
    //save
    
    let prev = await dbGlobal.collection('chart-' + work._id)
      .find({idx: {$lt: startSpan}})
      .sort({idx: -1})
      .toArray();
    if (prev.length) {
      prev = prev[0]
    } else {
      prev = undefined
    }
    
    let walletCursor = {
      data   : [],
      limit  : 1000,
      last_id: undefined,
      async next() {
        if (data.length === 0) {
          let find = undefined;
          if (this.last_id) {
            find = {_id: {$gt: this.last_id}}
          }
          this.data = await dbGlobal.collection('w-pool-' + work._id).find(find).limit(this.limit).toArray();
          if (this.data.length) {
            this.last_id = this.data[this.data.length - 1]._id
          }
        }
        return data.length ? this.data.splice(0, 1) : undefined
      }
    };
    
    let blockCursor = {
      data: blocks,
      async next() {
        return data.length ? this.data.splice(0, 1) : undefined
      }
    };
    
    let next = save(prev, walletCursor, blockCursor);
    
    next.idx = nextSpan;
    
    await dbGlobal.collection('chart-' + work._id).find({idx: next.idx}).upsert().updateOne({'$set': next});
    
    let lastBlock = blocks[blocks.length - 1];
    await col.updateOne({_id: work._id}, {$set: {last_end: lastBlock.time}});
    // console.log();
    setTimeout(check_work, 1000);
  } else {
    DelayWork()
  }
}

mongoClient.connect().then(() => {
  console.info('mongodb connected');
  
  const db = mongoClient.db('chain');
  dbGlobal = db;
  col      = db.collection('work');
  check_work();
  
  const router = express.Router();
  app.use('/api', api(router, db));
  
  console.info('start listen server localhost:8080');
  app.listen(8080);
});