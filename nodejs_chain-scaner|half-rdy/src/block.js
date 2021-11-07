// import collection from './collection'

export default {
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
    let col = db.collection('chain-' + name);
    
    let batch = col.initializeOrderedBulkOp();
    
    list.forEach(block => {
      batch.find({time: block.time}).upsert().updateOne({'$set': block});
    });
    
    return batch.execute();
  },
  sync(db, name) {
    let col = db.collection('chain-' + name);
    
    Promise.all(
      [
        col.createIndex({'time': 1}),
        col.count(),
        col.find().sort({time: 1}).limit(1).toArray(),
        col.find().sort({time: -1}).limit(1).toArray()
      ]
    )
      .then(([total, first, last]) => {
        db.collection('info').updateOne({name: name}, {
          '$set': {
            'total'     : total,
            'min'       : first[0].time,
            'max'       : last[0].time,
            'updated_at': new Date().getTime() / 1000 | 0
          }
        },{upsert: true})
      });
  },
  getInfo(db){
    return db.collection('info').find().toArray()
  }
}