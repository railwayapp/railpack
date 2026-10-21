const assert = require('node:assert/strict');
const Database = require('better-sqlite3');

assert.equal(process.version, 'v20.20.2');
console.log('Node version:', process.version);

const db = new Database(':memory:');
assert.equal(db.prepare('SELECT 42 AS answer').get().answer, 42);
db.close();

console.log('better-sqlite3 source build works on Node 20 with pnpm');
