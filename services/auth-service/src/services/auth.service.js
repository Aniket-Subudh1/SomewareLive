const { StatusCodes } = require('http-status-codes');
const { ApiError } = require('../utils/error-handler');
const { generateAccessToken, generateRefreshToken, verifyToken } = require('../utils/jwt.util');
const { comparePassword } = require('../utils/password.util');
const logger = require('../utils/logger');
const kafkaService = require('./kafka.service');
const { sequelize } = require('../db');
const db = require('../db').initModels();


const register = async (userData) => {
  const { email, password, firstName, lastName, role = 'user' } = userData;

  try {
    // Check if email already exists
    const existingUser = await db.User.findOne({ where: { email } });
    if (existingUser) {
      throw new ApiError(
        StatusCodes.CONFLICT,
        'Email already registered'
      );
    }

    // Create the user without transaction initially for simplicity
    const user = await db.User.create({
      email,
      password, // Will be hashed by model hooks
      firstName,
      lastName,
      role,
    });
    
    // Remove password from response
    const userResponse = user.toJSON();
    delete userResponse.password;

    // Try to send Kafka event but don't fail if it doesn't work
    try {
      await kafkaService.sendUserEvent('user.created', userResponse);
    } catch (kafkaError) {
      logger.warn('Failed to send user.created event to Kafka:', kafkaError.message);
      // Continue anyway
    }

    return userResponse;
  } catch (error) {
    logger.error('Error registering user:', error);
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(
      StatusCodes.INTERNAL_SERVER_ERROR,
      'Error registering user',
      false,
      error.stack
    );
  }
};


const login = async (email, password) => {
  try {
    // Find user with password
    const user = await db.User.scope('withPassword').findOne({ where: { email } });
    if (!user) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Invalid email or password'
      );
    }

    // Check if user is active
    if (!user.active) {
      throw new ApiError(
        StatusCodes.FORBIDDEN,
        'Account is disabled'
      );
    }

    // Verify password
    const isPasswordValid = await user.verifyPassword(password);
    if (!isPasswordValid) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Invalid email or password'
      );
    }

    // Generate tokens
    const accessToken = generateAccessToken(user);
    const refreshTokenObj = generateRefreshToken(user);

    // Update last login timestamp
    user.lastLogin = new Date();
    await user.save();

    // Create refresh token record directly
    await db.Token.create({
      userId: user.id,
      token: refreshTokenObj.token,
      type: 'refresh',
      expiresAt: refreshTokenObj.expiresAt,
      blacklisted: false
    });

    // Try to send Kafka event but don't fail if it doesn't work
    try {
      await kafkaService.sendAuthEvent('user.logged_in', {
        userId: user.id,
        email: user.email,
        timestamp: new Date().toISOString()
      });
    } catch (kafkaError) {
      logger.warn('Failed to send user.logged_in event to Kafka:', kafkaError.message);
      // Continue anyway
    }

    // Return tokens and user info (excluding password)
    const userResponse = user.toJSON();
    delete userResponse.password;

    return {
      accessToken,
      refreshToken: refreshTokenObj.token,
      user: userResponse
    };
  } catch (error) {
    logger.error('Error during login:', error);
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(
      StatusCodes.INTERNAL_SERVER_ERROR,
      'Error during login process',
      false,
      error.stack
    );
  }
};


const logout = async (refreshToken) => {
  if (!refreshToken) {
    return true; // No token to invalidate
  }

  try {
    // Decode token to get the user ID
    const decoded = verifyToken(refreshToken, 'refresh');
    if (!decoded) {
      return true; // Token already invalid
    }

    // Find token in database
    const tokenRecord = await db.Token.findOne({
      where: {
        token: refreshToken,
        type: 'refresh',
        blacklisted: false
      }
    });
    
    if (!tokenRecord) {
      return true; // Token not found or already blacklisted
    }

    // Blacklist token
    tokenRecord.blacklisted = true;
    await tokenRecord.save();

    // Try to send Kafka event
    try {
      await kafkaService.sendAuthEvent('user.logged_out', {
        userId: decoded.sub,
        timestamp: new Date().toISOString()
      });
    } catch (kafkaError) {
      logger.warn('Failed to send user.logged_out event to Kafka:', kafkaError.message);
      // Continue anyway
    }

    return true;
  } catch (error) {
    logger.error('Error during logout:', error);
    return false; // Return false on error but don't throw
  }
};


const refreshToken = async (refreshToken) => {
  try {
    // Verify refresh token
    const decoded = verifyToken(refreshToken, 'refresh');
    if (!decoded) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Invalid refresh token'
      );
    }

    // Find token in database
    const tokenRecord = await db.Token.findOne({
      where: {
        token: refreshToken,
        type: 'refresh',
        blacklisted: false,
        expiresAt: {
          [sequelize.Sequelize.Op.gt]: new Date()
        }
      }
    });
    
    if (!tokenRecord) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Refresh token revoked or expired'
      );
    }

    // Find user
    const user = await db.User.findByPk(decoded.sub);
    if (!user || !user.active) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'User not found or inactive'
      );
    }

    // Generate new access token
    const accessToken = generateAccessToken(user);

    // Try to send Kafka event
    try {
      await kafkaService.sendAuthEvent('user.token_refreshed', {
        userId: user.id,
        timestamp: new Date().toISOString()
      });
    } catch (kafkaError) {
      logger.warn('Failed to send user.token_refreshed event to Kafka:', kafkaError.message);
      // Continue anyway
    }

    return {
      accessToken
    };
  } catch (error) {
    logger.error('Error refreshing token:', error);
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(
      StatusCodes.INTERNAL_SERVER_ERROR,
      'Error refreshing token',
      false,
      error.stack
    );
  }
};


const verifyAuth = async (accessToken) => {
  try {
    // Verify access token
    const decoded = verifyToken(accessToken, 'access');
    if (!decoded) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'Invalid access token'
      );
    }

    // Find user
    const user = await db.User.findByPk(decoded.sub);
    if (!user || !user.active) {
      throw new ApiError(
        StatusCodes.UNAUTHORIZED,
        'User not found or inactive'
      );
    }

    return {
      id: user.id,
      email: user.email,
      role: user.role,
      firstName: user.firstName,
      lastName: user.lastName
    };
  } catch (error) {
    logger.error('Error verifying authentication:', error);
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(
      StatusCodes.UNAUTHORIZED,
      'Error verifying authentication',
      false,
      error.stack
    );
  }
};

module.exports = {
  register,
  login,
  logout,
  refreshToken,
  verifyAuth
};