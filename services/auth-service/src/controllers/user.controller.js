const { StatusCodes } = require('http-status-codes');
const asyncHandler = require('express-async-handler');
const userService = require('../services/user.service');
const logger = require('../utils/logger');


const getUsers = asyncHandler(async (req, res) => {
  const { page, limit, search, role } = req.query;
  const options = {
    page: parseInt(page) || 1,
    limit: parseInt(limit) || 10,
    search,
    role,
  };
  
  const result = await userService.getUsers(options);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    data: result,
  });
});


const getUserById = asyncHandler(async (req, res) => {
  const userId = req.params.id;
  const user = await userService.getUserById(userId);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    data: user,
  });
});


const createUser = asyncHandler(async (req, res) => {
  const user = await userService.createUser(req.body);
  
  res.status(StatusCodes.CREATED).json({
    status: 'success',
    message: 'User created successfully',
    data: user,
  });
});


const updateUser = asyncHandler(async (req, res) => {
  const userId = req.params.id;
  const updatedUser = await userService.updateUser(userId, req.body);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'User updated successfully',
    data: updatedUser,
  });
});


const changePassword = asyncHandler(async (req, res) => {
  const { currentPassword, newPassword } = req.body;
  const userId = req.user.id;
  
  await userService.changePassword(userId, currentPassword, newPassword);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Password changed successfully',
  });
});


const updateProfile = asyncHandler(async (req, res) => {
  const userId = req.user.id;
  // Only allow updating first name and last name for self-service
  const { firstName, lastName } = req.body;
  
  const updatedUser = await userService.updateUser(userId, { firstName, lastName });
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Profile updated successfully',
    data: updatedUser,
  });
});

const deactivateUser = asyncHandler(async (req, res) => {
  const userId = req.params.id;
  
  // Prevent deactivating own account
  if (userId === req.user.id) {
    return res.status(StatusCodes.BAD_REQUEST).json({
      status: 'error',
      message: 'You cannot deactivate your own account',
    });
  }
  
  await userService.deactivateUser(userId);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'User deactivated successfully',
  });
});


const deactivateOwnAccount = asyncHandler(async (req, res) => {
  const userId = req.user.id;
  
  await userService.deactivateUser(userId);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Your account has been deactivated successfully',
  });
});

module.exports = {
  getUsers,
  getUserById,
  createUser,
  updateUser,
  changePassword,
  updateProfile,
  deactivateUser,
  deactivateOwnAccount,
};