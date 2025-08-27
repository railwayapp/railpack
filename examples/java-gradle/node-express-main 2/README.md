# ExpressJS Starter Template!

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/Vp8hse?referralCode=vimo)

This is boilerplate/starter project for quickly building RESTful APIs using Node.js and [Express](https://expressjs.com/), written in JavaScript.
It will help you get started with a simple to follow format with some examples for routes, logging and middleware.

## Includes

### Server utilites:

- [morgan](https://www.npmjs.com/package/morgan)- HTTP request logger middleware for Node.js
- [dotenv](https://www.npmjs.com/package/dotenv)- Loads environment variables from `.env` file into `process.env`
- [cors](https://www.npmjs.com/package/cors)- CORS provides a Connect/Express middleware that can be used to enable CORS with various options.

### Development utilites

- [nodemon](https://www.npmjs.com/package/nodemon)- Helps develop node.js based application by automatically restarting the node server when it detects application file changes.

## Commands

### Setup

```
yarn
```

### Development

```
yarn dev
```

## About

The server runs a simple Express API server

`/` returns `status: ok`

`/hello` returns `message: Hello World!`

Unknown endpoints are handled in a middleware file.

The `hello` route is defined in the `helloRoute`

## Project Structure

```
src\
 |--routes\         # Routes
 |--utils\          # Utility files
 |--app.js          # Express app
 |--index.js        # App entry point
```
