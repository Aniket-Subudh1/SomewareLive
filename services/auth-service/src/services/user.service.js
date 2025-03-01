const { StatusCodes } = require('http-status-codes');
const { ApiError } = require('../utils/error-handler');
const { hashPassword } = require('../utils/password.util');
const logger = require('../utils/logger');
const kafkaService = require('./kafka.service');
const { sequelize } = require('../db');
const db = require('../db').initModels();


const getUsers = async (options = {}) => {
  const { page = 1, limit = 10, search, role } = options;
  const offset = (page - 1) * limit;
  
  // Build query filters
  const where = {};
  
  if (search) {
    where[sequelize.Sequelize.Op.or] = [
      { email: { [sequelize.Sequelize.Op.iLike]: `%${search}%` } },
      { firstName: { [sequelize.Sequelize.Op.iLike]: `%${search}%` } },
      { lastName: { [sequelize.Sequelize.Op.iLike]: `%${search}%` } }
    ];
  }
  
  if (role) {
    where.role = role;
  }

  // Get users
  const { count, rows } = await db.User.findAndCountAll({
    where,
    limit,
    offset,
    order: [['createdAt', 'DESC']],
  });

  return {
    totalItems: count,
    items: rows,
    totalPages: Math.ceil(count / limit),
    currentPage: page,
  };
};


const getUserById = async (userId) => {
  const user = await db.User.findByPk(userId);
  if (!user) {
    throw new ApiError(
      StatusCodes.NOT_FOUND,
      'User not found'
    );
  }
  return user;
};


const updateUser = async (userId, userData) => {
  const { firstName, lastName, role } = userData;
  
  // Find user
  const user = await db.User.findByPk(userId);
  if (!user) {
    throw new ApiError(
      StatusCodes.NOT_FOUND,
      'User not found'
    );
  }

  // Update user fields
  if (firstName !== undefined) user.firstName = firstName;
  if (lastName !== undefined) user.lastName = lastName;
  if (role !== undefined) user.role = role;

  // Save changes
  await user.save();

  // Send Kafka event
  await kafkaService.sendUserEvent('user.updated', user.toJSON());

  return user;
};


const changePassword = async (userId, currentPassword, newPassword) => {
  // Find user with password
  const user = await db.User.scope('withPassword').findByPk(userId);
  if (!user) {
    throw new ApiError(
      StatusCodes.NOT_FOUND,
      'User not found'
    );
  }

  // Verify current password
  const isPasswordValid = await user.verifyPassword(currentPassword);
  if (!isPasswordValid) {
    throw new ApiError(
      StatusCodes.UNAUTHORIZED,
      'Current password is incorrect'
    );
  }

  // Update password
  user.password = newPassword; // Model hooks will hash the password
  await user.save();

  // Invalidate all refresh tokens for this user
  await db.Token.update(
    { blacklisted: true },
    { where: { userId, type: 'refresh' } }
  );

  // Send Kafka event
  await kafkaService.sendUserEvent('user.password_changed', {
    userId: user.id,
    timestamp: new Date().toISOString()
  });

  return true;
};


const deactivateUser = async (userId) => {
  // Find user
  const user = await db.User.findByPk(userId);
  if (!user) {
    throw new ApiError(
      StatusCodes.NOT_FOUND,
      'User not found'
    );
  }

  // Deactivate user
  user.active = false;
  await user.save();

  // Invalidate all refresh tokens for this user
  await db.Token.update(
    { blacklisted: true },
    { where: { userId, type: 'refresh' } }
  );

  // Send Kafka event
  await kafkaService.sendUserEvent('user.deactivated', {
    userId: user.id,
    timestamp: new Date().toISOString()
  });

  return true;
};


const createUser = async (userData) => {
  const { email, password, firstName, lastName, role = 'user' } = userData;

  // Check if email already exists
  const existingUser = await db.User.findOne({ where: { email } });
  if (existingUser) {
    throw new ApiError(
      StatusCodes.CONFLICT,
      'Email already registered'
    );
  }

  // Create user with transaction
  const transaction = await sequelize.transaction();
  try {
    // Create the user
    const user = await db.User.create({
      email,
      password,
      firstName,
      lastName,
      role,
      verified: true, // Admin-created users are verified by default
    }, { transaction });

    // Commit transaction
    await transaction.commit();
    
    // Send Kafka event
    await kafkaService.sendUserEvent('user.created', user.toJSON());

    return user;
  } catch (error) {
    // Rollback transaction
    await transaction.rollback();
    logger.error('Error creating user:', error);
    throw error;
  }
};

module.exports = {
  getUsers,
  getUserById,
  updateUser,
  changePassword,
  deactivateUser,
  createUser,
};