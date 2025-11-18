const Database = require('better-sqlite3');

console.log('Node version:', process.version);
console.log('Testing better-sqlite3 (native module)...');

// Create an in-memory database
const db = new Database(':memory:');

// Create a simple table
db.exec('CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)');

// Insert some data
const insert = db.prepare('INSERT INTO test (name) VALUES (?)');
insert.run('npm');
insert.run('node-gyp');

// Query the data
const rows = db.prepare('SELECT * FROM test').all();

console.log('Database test successful!');
console.log('Rows:', rows);

db.close();

console.log('✓ better-sqlite3 native module working correctly with npm');
