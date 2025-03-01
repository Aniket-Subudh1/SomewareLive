const pinoHttp = require('pino-http');
const logger = require('../utils/logger');

/**
 * HTTP request logging middleware using pino-http
 */
const requestLogger = pinoHttp({
  logger,
  customLogLevel: (req, res, err) => {
    if (res.statusCode >= 500 || err) {
      return 'error';
    }
    if (res.statusCode >= 400) {
      return 'warn';
    }
    return 'info';
  },
  customSuccessMessage: (req, res) => {
    return `${req.method} ${req.url} completed with status ${res.statusCode}`;
  },
  customErrorMessage: (req, res, err) => {
    return `${req.method} ${req.url} failed with error: ${err.message}`;
  },
  customProps: (req, res) => {
    return {
      responseTime: res.responseTime,
      userAgent: req.headers['user-agent'],
      remoteAddress: req.ip,
      correlationId: req.headers['x-correlation-id'] || req.id,
    };
  },
  // Don't log requests for health checks to avoid cluttering logs
  autoLogging: {
    ignore: (req) => req.url === '/health',
  },
});

module.exports = {
  requestLogger,
};