const { StatusCodes } = require('http-status-codes');
const logger = require('./logger');

/**
 * Custom API Error class
 */
class ApiError extends Error {
  constructor(statusCode, message, isOperational = true, stack = '') {
    super(message);
    this.statusCode = statusCode;
    this.isOperational = isOperational;
    
    if (stack) {
      this.stack = stack;
    } else {
      Error.captureStackTrace(this, this.constructor);
    }
  }
}

/**
 * Handle 404 errors
 */
const notFoundHandler = (req, res, next) => {
  const error = new ApiError(
    StatusCodes.NOT_FOUND,
    `Route not found: ${req.method} ${req.url}`
  );
  next(error);
};

/**
 * Global error handler
 */
const errorHandler = (err, req, res, next) => {
  let { statusCode, message } = err;
  
  // Default status code and message if not provided
  statusCode = statusCode || StatusCodes.INTERNAL_SERVER_ERROR;
  message = message || 'Internal Server Error';
  
  // Handle Sequelize errors
  if (err.name === 'SequelizeValidationError') {
    statusCode = StatusCodes.BAD_REQUEST;
    message = err.errors.map(e => e.message).join(', ');
  } else if (err.name === 'SequelizeUniqueConstraintError') {
    statusCode = StatusCodes.CONFLICT;
    message = 'Resource already exists';
  } else if (err.name === 'SequelizeForeignKeyConstraintError') {
    statusCode = StatusCodes.BAD_REQUEST;
    message = 'Invalid related resource';
  } else if (err.name === 'SequelizeDatabaseError') {
    statusCode = StatusCodes.BAD_REQUEST;
    message = 'Database error';
  }
  
  // Handle validation errors
  if (err.name === 'ValidationError') {
    statusCode = StatusCodes.BAD_REQUEST;
  }
  
  // Handle JWT errors
  if (err.name === 'JsonWebTokenError') {
    statusCode = StatusCodes.UNAUTHORIZED;
    message = 'Invalid token';
  } else if (err.name === 'TokenExpiredError') {
    statusCode = StatusCodes.UNAUTHORIZED;
    message = 'Token expired';
  }
  
  // Log error
  const logLevel = statusCode >= 500 ? 'error' : 'warn';
  logger[logLevel]({
    err: {
      message: err.message,
      stack: err.stack,
      name: err.name,
    },
    request: {
      method: req.method,
      url: req.url,
      headers: req.headers,
      params: req.params,
      query: req.query,
      body: req.body,
    },
  }, `${statusCode} - ${message}`);
  
  // Send response
  res.status(statusCode).json({
    status: 'error',
    statusCode,
    message,
    ...(process.env.NODE_ENV === 'development' && { stack: err.stack }),
  });
};

/**
 * Create a dedicated error for a specific case
 */
const createError = (statusCode, message, isOperational = true) => {
  return new ApiError(statusCode, message, isOperational);
};

module.exports = {
  ApiError,
  notFoundHandler,
  errorHandler,
  createError,
};