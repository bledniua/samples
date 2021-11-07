import block from './block'

export default (db) => {
  return {
    addBlock: function (req, resp) {
      if (req.body.name === undefined) {
        resp.end(JSON.stringify({error: 'name is empty'}));
        return
      }
      
      if (typeof req.body.name !== 'string') {
        resp.end(JSON.stringify({error: 'name is not string'}));
        return
      }
      
      const name = req.body.name;
      
      let list = [];
      if (req.body.item !== undefined) {
        let check = block.check(req.body.item);
        if (check !== undefined) {
          resp.end(JSON.stringify({item: req.body.item, error: check}));
          return;
        }
        list.push(req.body.item)
      }
      
      if (req.body.list !== undefined) {
        list = req.body.list;
        for (let i = 0; i < list.length; i++) {
          let error = block.check(list[i]);
          if (error !== undefined) {
            resp.end(JSON.stringify({item: req.body.item, error: error}));
            return;
          }
        }
      }
      
      block.save(db, name, list).then(({error, result}) => {
        if (error === undefined) {
          resp.end(JSON.stringify(result));
          block.sync(db, name)
        } else {
          resp.end(JSON.stringify({error: error}));
        }
      });
    },
    getBlockInfo: function (req, resp) {
      block.getInfo(db).then(data => {
        resp.end(JSON.stringify(data));
      })
    }
  }
}