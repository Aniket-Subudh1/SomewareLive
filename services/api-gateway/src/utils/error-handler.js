const { StatusCodes } = require('http-status-codes');
const logger = require('./logger');

// Global error handler middleware
const errorHandler = (err, req, res, next) => {
  let statusCode = err.statusCode || StatusCodes.INTERNAL_SERVER_ERROR;
  let message = err.message || 'Internal Server Error';
  let error = err.name || 'Error';

  // Handle JWT errors
  if (err.name === 'UnauthorizedError') {
    statusCode = StatusCodes.UNAUTHORIZED;
    message = 'Invalid token';
    error = 'AuthenticationError';
  }

  // Log the error
  logger.error({
    err,
    request: {
      method: req.method,
      url: req.url,
      params: req.params,
      query: req.query,
    },
  }, `Error: ${message}`);

  // Send error response
  res.status(statusCode).json({
    error,
    message,
    statusCode,
    timestamp: new Date().toISOString(),
    path: req.url,
  });
};

// Not found middleware
const notFoundHandler = (req, res) => {
  logger.warn(`Route not found: ${req.method} ${req.url}`);
  
  res.status(StatusCodes.NOT_FOUND).json({
    error: 'NotFoundError',
    message: `Route not found: ${req.method} ${req.url}`,
    statusCode: StatusCodes.NOT_FOUND,
    timestamp: new Date().toISOString(),
    path: req.url,
  });
};

module.exports = {
  errorHandler,
  notFoundHandler,
};