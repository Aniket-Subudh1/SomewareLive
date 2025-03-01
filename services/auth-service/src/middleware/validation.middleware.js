const Joi = require('joi');
const { StatusCodes } = require('http-status-codes');
const { validatePasswordStrength } = require('../utils/password.util');


const validate = (schema, property = 'body') => {
  return (req, res, next) => {
    const { error, value } = schema.validate(req[property], {
      abortEarly: false,
      stripUnknown: true,
    });

    if (error) {
      const errors = error.details.map((detail) => ({
        field: detail.path.join('.'),
        message: detail.message,
      }));

      return res.status(StatusCodes.BAD_REQUEST).json({
        status: 'error',
        message: 'Validation error',
        errors,
      });
    }

    // Replace request data with validated data
    req[property] = value;
    return next();
  };
};

// Custom password validator function
const passwordValidator = (value, helpers) => {
  const result = validatePasswordStrength(value);
  if (!result.isValid) {
    return helpers.message(result.message);
  }
  return value;
};

// Registration validation schema
const registerSchema = Joi.object({
  email: Joi.string()
    .email({ tlds: { allow: false } })
    .required()
    .trim()
    .lowercase()
    .messages({
      'string.email': 'Please enter a valid email address',
      'string.empty': 'Email is required',
      'any.required': 'Email is required',
    }),
  password: Joi.string()
    .min(8)
    .required()
    .custom(passwordValidator)
    .messages({
      'string.min': 'Password must be at least 8 characters long',
      'string.empty': 'Password is required',
      'any.required': 'Password is required',
    }),
  firstName: Joi.string()
    .required()
    .trim()
    .messages({
      'string.empty': 'First name is required',
      'any.required': 'First name is required',
    }),
  lastName: Joi.string()
    .required()
    .trim()
    .messages({
      'string.empty': 'Last name is required',
      'any.required': 'Last name is required',
    }),
  role: Joi.string()
    .valid('user', 'presenter', 'admin')
    .default('user'),
});

// Login validation schema
const loginSchema = Joi.object({
  email: Joi.string()
    .email({ tlds: { allow: false } })
    .required()
    .trim()
    .lowercase()
    .messages({
      'string.email': 'Please enter a valid email address',
      'string.empty': 'Email is required',
      'any.required': 'Email is required',
    }),
  password: Joi.string()
    .required()
    .messages({
      'string.empty': 'Password is required',
      'any.required': 'Password is required',
    }),
});

// Refresh token validation schema
const refreshTokenSchema = Joi.object({
  refreshToken: Joi.string()
    .required()
    .messages({
      'string.empty': 'Refresh token is required',
      'any.required': 'Refresh token is required',
    }),
});

// Change password validation schema
const changePasswordSchema = Joi.object({
  currentPassword: Joi.string()
    .required()
    .messages({
      'string.empty': 'Current password is required',
      'any.required': 'Current password is required',
    }),
  newPassword: Joi.string()
    .min(8)
    .required()
    .custom(passwordValidator)
    .messages({
      'string.min': 'New password must be at least 8 characters long',
      'string.empty': 'New password is required',
      'any.required': 'New password is required',
    }),
  confirmPassword: Joi.string()
    .required()
    .valid(Joi.ref('newPassword'))
    .messages({
      'string.empty': 'Confirm password is required',
      'any.required': 'Confirm password is required',
      'any.only': 'Passwords do not match',
    }),
});

// Update user validation schema
const updateUserSchema = Joi.object({
  firstName: Joi.string().trim(),
  lastName: Joi.string().trim(),
  role: Joi.string().valid('user', 'presenter', 'admin'),
});

// ID parameter validation schema
const idParamSchema = Joi.object({
  id: Joi.string()
    .guid({ version: 'uuidv4' })
    .required()
    .messages({
      'string.guid': 'Invalid ID format',
      'string.empty': 'ID is required',
      'any.required': 'ID is required',
    }),
});

module.exports = {
  validateRegister: validate(registerSchema),
  validateLogin: validate(loginSchema),
  validateRefreshToken: validate(refreshTokenSchema),
  validateChangePassword: validate(changePasswordSchema),
  validateUpdateUser: validate(updateUserSchema),
  validateIdParam: validate(idParamSchema, 'params'),
};