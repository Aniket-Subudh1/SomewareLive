const express = require('express');
const router = express.Router();
const { testConnection } = require('../db');
const { connectProducer } = require('../services/kafka.service');
const logger = require('../utils/logger');

router.get('/', async (req, res) => {
  const status = {
    service: 'auth-service',
    status: 'UP',
    timestamp: new Date().toISOString(),
    checks: {},
  };

  // Check database connection
  try {
    const isDbConnected = await testConnection();
    status.checks.database = {
      status: isDbConnected ? 'UP' : 'DOWN',
    };
  } catch (error) {
    logger.error('Health check database error:', error);
    status.checks.database = {
      status: 'DOWN',
      error: error.message,
    };
  }

  // Check Kafka connection
  try {
    await connectProducer();
    status.checks.kafka = {
      status: 'UP',
    };
  } catch (error) {
    logger.error('Health check Kafka error:', error);
    status.checks.kafka = {
      status: 'DOWN',
      error: error.message,
    };
  }

  // Determine overall status
  const isAllUp = Object.values(status.checks).every(
    check => check.status === 'UP'
  );
  
  status.status = isAllUp ? 'UP' : 'DEGRADED';
  
  // Return an appropriate status code
  const statusCode = isAllUp ? 200 : 503;
  res.status(statusCode).json(status);
});

module.exports = router;