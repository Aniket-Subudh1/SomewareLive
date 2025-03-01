const bcrypt = require('bcrypt');
const config = require('../config');

/**
 * Hash a password
 * @param {string} password - The plain text password
 * @returns {Promise<string>} - The hashed password
 */
const hashPassword = async (password) => {
  return bcrypt.hash(password, config.password.saltRounds);
};

/**
 * Compare a plain text password with a hash
 * @param {string} password - The plain text password
 * @param {string} hash - The hashed password
 * @returns {Promise<boolean>} - True if the password matches the hash
 */
const comparePassword = async (password, hash) => {
  return bcrypt.compare(password, hash);
};

/**
 * Generate a random password
 * @param {number} length - The length of the password
 * @returns {string} - The generated password
 */
const generateRandomPassword = (length = 12) => {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_-+=';
  let password = '';
  
  // Ensure at least one character from each category
  password += chars.slice(0, 26).charAt(Math.floor(Math.random() * 26)); // lowercase
  password += chars.slice(26, 52).charAt(Math.floor(Math.random() * 26)); // uppercase
  password += chars.slice(52, 62).charAt(Math.floor(Math.random() * 10)); // number
  password += chars.slice(62).charAt(Math.floor(Math.random() * (chars.length - 62))); // special
  
  // Fill the rest of the password
  for (let i = 4; i < length; i++) {
    password += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  
  // Shuffle the password
  password = password.split('').sort(() => 0.5 - Math.random()).join('');
  
  return password;
};

/**
 * Validate password strength
 * @param {string} password - The password to validate
 * @returns {Object} - Validation result with isValid and message properties
 */
const validatePasswordStrength = (password) => {
  if (!password || password.length < 8) {
    return {
      isValid: false,
      message: 'Password must be at least 8 characters long'
    };
  }
  
  // Check for at least one uppercase letter
  if (!/[A-Z]/.test(password)) {
    return {
      isValid: false,
      message: 'Password must contain at least one uppercase letter'
    };
  }
  
  // Check for at least one lowercase letter
  if (!/[a-z]/.test(password)) {
    return {
      isValid: false,
      message: 'Password must contain at least one lowercase letter'
    };
  }
  
  // Check for at least one number
  if (!/\d/.test(password)) {
    return {
      isValid: false,
      message: 'Password must contain at least one number'
    };
  }
  
  // Check for at least one special character
  if (!/[!@#$%^&*()_\-+=<>?]/.test(password)) {
    return {
      isValid: false,
      message: 'Password must contain at least one special character'
    };
  }
  
  return {
    isValid: true,
    message: 'Password meets strength requirements'
  };
};

module.exports = {
  hashPassword,
  comparePassword,
  generateRandomPassword,
  validatePasswordStrength,
};