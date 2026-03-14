const Database = require('better-sqlite3');
const bcrypt = require('bcrypt');

console.log('Node version:', process.version);
console.log('Testing better-sqlite3 and bcrypt (native modules)...');

// Create an in-memory database
const db = new Database(':memory:');

// Create a simple table
db.exec('CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)');

// Insert some data
const insert = db.prepare('INSERT INTO test (name) VALUES (?)');
insert.run('pnpm');
insert.run('node-gyp');

// Query the data
const rows = db.prepare('SELECT * FROM test').all();

console.log('Database test successful!');
console.log('Rows:', rows);

db.close();

console.log('Testing bcrypt (native module)...');
const saltRounds = 10;
const hash = bcrypt.hashSync('pnpm-native-deps', saltRounds);
const match = bcrypt.compareSync('pnpm-native-deps', hash);
console.log('Bcrypt hash generated:', hash.substring(0, 20) + '...');
console.log('Bcrypt compare result:', match);

console.log('✓ better-sqlite3 and bcrypt native modules working correctly with pnpm');
