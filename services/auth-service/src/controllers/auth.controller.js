const { StatusCodes } = require('http-status-codes');
const asyncHandler = require('express-async-handler');
const authService = require('../services/auth.service');
const { extractTokenFromHeader } = require('../utils/jwt.util');
const logger = require('../utils/logger');


const register = asyncHandler(async (req, res) => {
  const user = await authService.register(req.body);
  
  res.status(StatusCodes.CREATED).json({
    status: 'success',
    message: 'User registered successfully',
    data: user,
  });
});


const login = asyncHandler(async (req, res) => {
  const { email, password } = req.body;
  const result = await authService.login(email, password);
  
  // Set refresh token in an HTTP-only cookie
  res.cookie('refreshToken', result.refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
  });
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Login successful',
    data: {
      accessToken: result.accessToken,
      user: result.user,
    },
  });
});


const logout = asyncHandler(async (req, res) => {
  // Get refresh token from cookie or request body
  const refreshToken = req.cookies?.refreshToken || req.body?.refreshToken;
  
  await authService.logout(refreshToken);
  
  // Clear the refresh token cookie
  res.clearCookie('refreshToken');
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Logout successful',
  });
});


const refreshAccessToken = asyncHandler(async (req, res) => {
  // Get refresh token from cookie or request body
  const refreshToken = req.cookies?.refreshToken || req.body?.refreshToken;
  
  if (!refreshToken) {
    return res.status(StatusCodes.BAD_REQUEST).json({
      status: 'error',
      message: 'Refresh token is required',
    });
  }
  
  const result = await authService.refreshToken(refreshToken);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Token refreshed successfully',
    data: result,
  });
});


const verifyToken = asyncHandler(async (req, res) => {
  // Get token from authorization header
  const authHeader = req.headers.authorization;
  if (!authHeader) {
    return res.status(StatusCodes.BAD_REQUEST).json({
      status: 'error',
      message: 'Authorization header is required',
    });
  }
  
  const token = extractTokenFromHeader(authHeader);
  if (!token) {
    return res.status(StatusCodes.BAD_REQUEST).json({
      status: 'error',
      message: 'Invalid token format',
    });
  }
  
  const user = await authService.verifyAuth(token);
  
  res.status(StatusCodes.OK).json({
    status: 'success',
    message: 'Token is valid',
    data: { user },
  });
});


const getCurrentUser = asyncHandler(async (req, res) => {
  res.status(StatusCodes.OK).json({
    status: 'success',
    data: req.user,
  });
});

module.exports = {
  register,
  login,
  logout,
  refreshAccessToken,
  verifyToken,
  getCurrentUser,
};