export default {
  New  : function(){
    return {
      _id   : '',
      start : 0,
      end   : 0,
      period: 0,
      
      name: '',
      src : '',
      
      created_at: 0,
      last_end  : 0,
      next      : 0,
      meta      : {
        binanceAddressList: [
          'a1', 'a2'
        ]
      },
      
      check: [
        (pool, tx) => {
          if (tx.to.isBinance) {
            pool.setFlag(tx.from, 'binance_l2')
          }
        },
        (pool, tx) => {
          [tx.from, tx.to].forEach(w => {
            if (pool.get(w).amount > 100) {
              pool.checkFlag(w, 'a100')
            }
          })
        }
      ],
      
      // filter() {},
      tickBlock(block) {
        //Calc
        block.tx.forEach(rec => {
          
        })
        
        //pre calc is binance wallet?
        
        //check is tx to in binance group?
        //math flow
        //check amount
        
        //save Snap groups.
      },
      
      walletFilter(w) {
        return w.amount > 100
      },
      flowFilter(from, to) {
      
      },
      
      walletTick(from, to) {
        // from.flagCheck(w => w.amount > 100);
        // from.flagCheck(w => to === '');
        
        // if (from.amount > 100){
        //   from.g100 = true
        // }
        
      },
      
      writer(db, time, value) {
        db.collection('chart-' + 'test').updateOne({time: time}, {
          '$set': {
            'time' : time,
            'value': value
          }
        }, {upsert: true})
        
      }
      
    }
  },
  check: function (block) {
    if (block.time === undefined) {
      return 'time is empty'
    }
    if (block.tx === undefined) {
      return 'tx is empty'
    }
    if (!Array.isArray(block.tx)) {
      return 'tx must be array'
    }
  },
  save(db, name, list) {
    // let col = db.collection('chain-' + name);
    //
    // let batch = col.initializeOrderedBulkOp();
    //
    // list.forEach(block => {
    //   batch.find({time: block.time}).upsert().updateOne({'$set': block});
    // });
    //
    // return batch.execute();
  },
  sync(db, name) {
    // let col = db.collection('chain-' + name);
    //
    // Promise.all(
    //   [
    //     col.createIndex({'time': 1}),
    //     col.count(),
    //     col.find().sort({time: 1}).limit(1).toArray(),
    //     col.find().sort({time: -1}).limit(1).toArray()
    //   ]
    // )
    //   .then(([total, first, last]) => {
    //     db.collection('info').updateOne({name: name}, {
    //       '$set': {
    //         'total'     : total,
    //         'min'       : first[0].time,
    //         'max'       : last[0].time,
    //         'updated_at': new Date().getTime() / 1000 | 0
    //       }
    //     }, {upsert: true})
    //   });
  },
  getInfo(db) {
    // return db.collection('info').find().toArray()
  }
}