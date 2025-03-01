const { expressjwt: jwt } = require('express-jwt');
const { StatusCodes } = require('http-status-codes');
const config = require('../config');
const logger = require('../utils/logger');

/**
 * JWT authentication middleware
 */

const authenticate = jwt({
  secret: config.jwt.secret,
  algorithms: ['HS256'],
  credentialsRequired: true,
});

/**
 * Optional JWT authentication middleware
 * Does not require authentication but will validate the token if present
 */

const optionalAuthenticate = jwt({
  secret: config.jwt.secret,
  algorithms: ['HS256'],
  credentialsRequired: false,
});

/**
 * Middleware to check if the user has the required role
 * @param {string|string[]} roles - Required role(s)
 */

const authorize = (roles) => {
  return (req, res, next) => {
    // If no user is authenticated (should not happen due to authenticate middleware)
    if (!req.auth) {
      logger.warn('Unauthorized access attempt without authentication');
      return res.status(StatusCodes.UNAUTHORIZED).json({
        error: 'AuthenticationError',
        message: 'Authentication required',
      });
    }

    const userRoles = req.auth.roles || [];
    
    // Convert roles parameter to array if it's a string
    const requiredRoles = Array.isArray(roles) ? roles : [roles];
    
    // Check if user has any of the required roles
    const hasRequiredRole = requiredRoles.some(role => userRoles.includes(role));
    
    if (!hasRequiredRole) {
      logger.warn(`Forbidden access attempt by user ${req.auth.sub} to ${req.originalUrl}`);
      return res.status(StatusCodes.FORBIDDEN).json({
        error: 'ForbiddenError',
        message: 'You do not have permission to access this resource',
      });
    }
    
    next();
  };
};

module.exports = {
  authenticate,
  optionalAuthenticate,
  authorize,
};