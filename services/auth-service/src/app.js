const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const cookieParser = require('cookie-parser');
const pinoHttp = require('pino-http');
const { testConnection } = require('./db');
const { connectProducer, disconnectProducer } = require('./services/kafka.service');
const authRoutes = require('./routes/auth.routes');
const healthRoutes = require('./routes/health.routes');
const { errorHandler, notFoundHandler } = require('./utils/error-handler');
const config = require('./config');
const logger = require('./utils/logger');

// Initialize Express app
const app = express();

// Initialize database models
require('./db').initModels();

// Set up middlewares
app.use(helmet()); // Security headers
app.use(cors({
  origin: config.cors.origin,
  credentials: true,
  exposedHeaders: ['Content-Length', 'Authorization'],
}));
app.use(express.json()); // Parse JSON bodies
app.use(express.urlencoded({ extended: true })); // Parse URL-encoded bodies
app.use(cookieParser()); // Parse cookies

// Request logging
app.use(pinoHttp({
  logger,
  customLogLevel: (req, res, err) => {
    if (res.statusCode >= 500 || err) return 'error';
    if (res.statusCode >= 400) return 'warn';
    return 'info';
  },
  autoLogging: {
    ignore: (req) => req.url === '/health',
  },
}));

// Set up routes
app.use('/health', healthRoutes);
app.use('/', authRoutes);

// Error handling
app.use(notFoundHandler);
app.use(errorHandler);

// Start the server
const PORT = config.server.port;

async function startServer() {
  try {
    // Test database connection
    const isDbConnected = await testConnection();
    if (!isDbConnected) {
      logger.error('Unable to connect to the database');
      process.exit(1);
    }

    // Connect to Kafka
    await connectProducer();

    // Start the server
    app.listen(PORT, () => {
      logger.info(`Auth Service started on port ${PORT} in ${config.server.env} mode`);
      logger.info(`Health check available at: http://localhost:${PORT}/health`);
    });
  } catch (error) {
    logger.error('Failed to start server:', error);
    process.exit(1);
  }
}

// Handle graceful shutdown
process.on('SIGTERM', async () => {
  logger.info('SIGTERM received. Shutting down gracefully...');
  await disconnectProducer();
  process.exit(0);
});

process.on('SIGINT', async () => {
  logger.info('SIGINT received. Shutting down gracefully...');
  await disconnectProducer();
  process.exit(0);
});

// Handle unhandled rejections and exceptions
process.on('unhandledRejection', (reason, promise) => {
  logger.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

process.on('uncaughtException', (error) => {
  logger.fatal('Uncaught Exception:', error);
  process.exit(1);
});

// Start the server
startServer();

module.exports = app;