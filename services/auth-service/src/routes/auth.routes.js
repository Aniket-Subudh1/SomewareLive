const express = require('express');
const router = express.Router();
const authController = require('../controllers/auth.controller');
const userController = require('../controllers/user.controller');
const { authenticate, authorize } = require('../middleware/auth.middleware');
const {
  validateRegister,
  validateLogin,
  validateRefreshToken,
  validateChangePassword,
  validateUpdateUser,
  validateIdParam,
} = require('../middleware/validation.middleware');

// Public routes
router.post('/register', validateRegister, authController.register);
router.post('/login', validateLogin, authController.login);
router.post('/logout', authController.logout);
router.post('/refresh-token', validateRefreshToken, authController.refreshAccessToken);
router.get('/verify', authController.verifyToken);

// Protected routes
router.get('/me', authenticate, authController.getCurrentUser);
router.put('/users/profile', authenticate, validateUpdateUser, userController.updateProfile);
router.post('/users/change-password', authenticate, validateChangePassword, userController.changePassword);
router.delete('/users/profile', authenticate, userController.deactivateOwnAccount);

// Admin routes
router.get('/users', authenticate, authorize('admin'), userController.getUsers);
router.post('/users', authenticate, authorize('admin'), validateRegister, userController.createUser);
router.get('/users/:id', authenticate, authorize('admin'), validateIdParam, userController.getUserById);
router.put('/users/:id', authenticate, authorize('admin'), validateIdParam, validateUpdateUser, userController.updateUser);
router.delete('/users/:id', authenticate, authorize('admin'), validateIdParam, userController.deactivateUser);

module.exports = router;