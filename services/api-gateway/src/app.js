const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const routes = require('./routes');
const config = require('./config');
const { requestLogger } = require('./middleware/logging.middleware');
const { basicLimiter } = require('./middleware/rateLimit.middleware');
const { errorHandler, notFoundHandler } = require('./utils/error-handler');
const logger = require('./utils/logger');

const app = express();

// Apply middlewares
app.use(helmet()); // Security headers
app.use(cors({
  origin: config.cors.origin,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Correlation-Id'],
  exposedHeaders: ['X-Correlation-Id'],
}));
app.use(express.json()); // Parse JSON bodies
app.use(express.urlencoded({ extended: true })); // Parse URL-encoded bodies
app.use(requestLogger); // Request logging
app.use(basicLimiter); // Rate limiting

// Add routes
app.use('/', routes);

// Error handling
app.use(notFoundHandler); // Handle 404 errors
app.use(errorHandler); // Global error handler

// Start the server
const PORT = config.server.port;
app.listen(PORT, () => {
  logger.info(`API Gateway started on port ${PORT} in ${config.server.env} mode`);
  logger.info(`Health check available at: http://localhost:${PORT}/health`);
});

// Handle unhandled rejections and exceptions
process.on('unhandledRejection', (reason, promise) => {
  logger.error(`Unhandled Rejection at: ${promise}, reason: ${reason}`);
});

process.on('uncaughtException', (error) => {
  logger.fatal(`Uncaught Exception: ${error.message}`, error);
  process.exit(1);
});

module.exports = app;