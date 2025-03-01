const express = require('express');
const router = express.Router();
const axios = require('axios');
const services = require('../config/services');
const logger = require('../utils/logger');

/**
 * @route GET /health
 * @desc Health check endpoint for API Gateway
 * @access Public
 */

router.get('/', async (req, res) => {
  // Basic health check for the API gateway itself
  res.status(200).json({
    status: 'UP',
    service: 'api-gateway',
    timestamp: new Date().toISOString(),
  });
});

/**
 * @route GET /health/services
 * @desc Check health of all services
 * @access Public
 */

router.get('/services', async (req, res) => {
  const serviceHealth = {};
  const servicePromises = [];

  // Check each service's health in parallel
  Object.entries(services).forEach(([name, { url }]) => {
    const serviceUrl = `${url}/health`;
    const promise = axios.get(serviceUrl, { timeout: 3000 })
      .then(() => {
        serviceHealth[name] = { status: 'UP' };
      })
      .catch(error => {
        logger.warn(`Health check failed for ${name} service: ${error.message}`);
        serviceHealth[name] = { 
          status: 'DOWN',
          error: error.code || error.message
        };
      });
    
    servicePromises.push(promise);
  });

  // Wait for all services to be checked
  await Promise.all(servicePromises);

  // Determine overall status - UP only if all services are UP
  const overallStatus = Object.values(serviceHealth).every(
    service => service.status === 'UP'
  ) ? 'UP' : 'DEGRADED';

  res.status(200).json({
    status: overallStatus,
    services: serviceHealth,
    timestamp: new Date().toISOString(),
  });
});

module.exports = router;