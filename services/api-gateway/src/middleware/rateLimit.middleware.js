const rateLimit = require('express-rate-limit');
const config = require('../config');
const logger = require('../utils/logger');


const basicLimiter = rateLimit({
  windowMs: config.rateLimit.windowMs,
  max: config.rateLimit.max,
  standardHeaders: true, // Return rate limit info in the `RateLimit-*` headers
  legacyHeaders: false, // Disable the `X-RateLimit-*` headers
  handler: (req, res, next, options) => {
    logger.warn(`Rate limit exceeded for IP: ${req.ip}`);
    res.status(options.statusCode).json({
      error: 'RateLimitError',
      message: options.message,
      statusCode: options.statusCode,
    });
  }
});


const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 20, // limit each IP to 20 requests per windowMs
  standardHeaders: true,
  legacyHeaders: false,
  handler: (req, res, next, options) => {
    logger.warn(`Auth rate limit exceeded for IP: ${req.ip}`);
    res.status(options.statusCode).json({
      error: 'RateLimitError',
      message: 'Too many authentication attempts, please try again later',
      statusCode: options.statusCode,
    });
  }
});

module.exports = {
  basicLimiter,
  authLimiter,
};