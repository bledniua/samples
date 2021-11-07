import init from './controller';

export default (router, db)=>{
  let controllers = init(db);
  
  router.all('/add-block', controllers.addBlock);
  router.all('/block-info', controllers.getBlockInfo);
  
  return router
};