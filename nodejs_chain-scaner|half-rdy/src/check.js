// let x = eval('({check: () => {return 1}})');
// console.log(x);
// console.log(x.check());

let HashMap = require('hashmap');

let m = new HashMap();
console.log(m);
m.set('x', {data: 'text'});
console.log(m);
let x = m.get('x');
x.test = false;
console.log(m);
console.log(m.get('x'));


function f(pool) {
  pool.get('x').t = false
}
f(m);

console.log(m);
console.log(m.get('x'));