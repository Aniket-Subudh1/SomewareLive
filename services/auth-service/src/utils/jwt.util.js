const jwt = require('jsonwebtoken');
const { v4: uuidv4 } = require('uuid');
const config = require('../config');
const logger = require('./logger');

/**
 * Generate a JWT access token
 * @param {Object} user - User object
 * @returns {string} - JWT token
 */
const generateAccessToken = (user) => {
  const payload = {
    sub: user.id,
    email: user.email,
    role: user.role,
    type: 'access',
    jti: uuidv4(),
  };

  const options = {
    expiresIn: config.jwt.accessExpiresIn,
    issuer: config.jwt.issuer,
  };

  return jwt.sign(payload, config.jwt.secret, options);
};

/**
 * Generate a JWT refresh token
 * @param {Object} user - User object
 * @returns {Object} - Token object with value and expiration
 */
const generateRefreshToken = (user) => {
  const jti = uuidv4();
  
  const payload = {
    sub: user.id,
    type: 'refresh',
    jti,
  };

  const options = {
    expiresIn: config.jwt.refreshExpiresIn,
    issuer: config.jwt.issuer,
  };

  const token = jwt.sign(payload, config.jwt.secret, options);
  
  // Calculate expiration time
  const decoded = jwt.decode(token);
  const expiresAt = new Date(decoded.exp * 1000);
  
  return {
    token,
    jti,
    expiresAt,
  };
};

/**
 * Verify a JWT token
 * @param {string} token - JWT token to verify
 * @param {string} type - Token type ('access' or 'refresh')
 * @returns {Object} - Decoded token payload or null if invalid
 */
const verifyToken = (token, type) => {
  try {
    const decoded = jwt.verify(token, config.jwt.secret);
    
    // Check if token type matches
    if (decoded.type !== type) {
      logger.warn(`Token type mismatch: expected ${type}, got ${decoded.type}`);
      return null;
    }
    
    return decoded;
  } catch (error) {
    logger.warn(`Token verification failed: ${error.message}`);
    return null;
  }
};

/**
 * Decode a JWT token without verification
 * @param {string} token - JWT token to decode
 * @returns {Object} - Decoded token payload
 */
const decodeToken = (token) => {
  return jwt.decode(token);
};

/**
 * Extract token from authorization header
 * @param {string} authHeader - Authorization header value
 * @returns {string|null} - Token or null if not found
 */
const extractTokenFromHeader = (authHeader) => {
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return null;
  }
  
  return authHeader.substring(7);
};

module.exports = {
  generateAccessToken,
  generateRefreshToken,
  verifyToken,
  decodeToken,
  extractTokenFromHeader,
};
