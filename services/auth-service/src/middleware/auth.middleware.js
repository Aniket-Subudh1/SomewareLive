const { StatusCodes } = require('http-status-codes');
const { verifyToken, extractTokenFromHeader } = require('../utils/jwt.util');
const { ApiError } = require('../utils/error-handler');
const db = require('../db').initModels();
const logger = require('../utils/logger');


const authenticate = async (req, res, next) => {
  try {
    // Get authorization header
    const authHeader = req.headers.authorization;
    if (!authHeader) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Authentication required'
      );
    }

    // Extract token
    const token = extractTokenFromHeader(authHeader);
    if (!token) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Invalid token format'
      );
    }

    // Verify token
    const decoded = verifyToken(token, 'access');
    if (!decoded) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Invalid or expired token'
      );
    }

    // Check if user exists and is active
    const user = await db.User.findByPk(decoded.sub);
    if (!user || !user.active) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'User not found or inactive'
      );
    }

    // Attach user to request
    req.user = {
      id: user.id,
      email: user.email,
      role: user.role,
      firstName: user.firstName,
      lastName: user.lastName,
    };

    next();
  } catch (error) {
    next(error);
  }
};


const authorize = (roles) => {
  return (req, res, next) => {
    try {
      if (!req.user) {
        throw new ApiError(
          StatusCodes.UNAUTHORIZED,
          'Authentication required'
        );
      }

      // Convert roles to array if string
      const requiredRoles = Array.isArray(roles) ? roles : [roles];
      
      // Check if user has required role
      if (!requiredRoles.includes(req.user.role)) {
        logger.warn(`User ${req.user.id} with role ${req.user.role} attempted to access resource requiring ${requiredRoles.join(', ')}`);
        throw new ApiError(
          StatusCodes.FORBIDDEN,
          'Insufficient permissions'
        );
      }

      next();
    } catch (error) {
      next(error);
    }
  };
};

module.exports = {
  authenticate,
  authorize,
};