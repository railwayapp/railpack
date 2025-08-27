import http from "http";
import { config } from "dotenv";
import app from "./app.js";
import * as logger from "./utils/logger.js";
import process from 'node:process';

if (process.env.NODE_ENV !== "production") {
	config();
}
const server = http.createServer(app);

const PORT = process.env.PORT || 3003;

server.listen(PORT, () => {
	logger.info(`Server listening at http://localhost:${PORT}`);
	logger.info(`Access the root route at http://localhost:${PORT}/hello`);
});

process.on('SIGTERM', () => {
  console.log('Shutting down gracefully...');
  console.log('Cleanup complete, exiting.');
  process.exit(0);
});
