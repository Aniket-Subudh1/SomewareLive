const express = require('express');
const router = express.Router();
const proxy = require('express-http-proxy');
const { authenticate, optionalAuthenticate } = require('../middleware/auth.middleware');
const { authLimiter } = require('../middleware/rateLimit.middleware');
const services = require('../config/services');
const healthRoutes = require('./health.route');
const logger = require('../utils/logger');

// Health routes
router.use('/health', healthRoutes);

// Setup routes for each service
Object.entries(services).forEach(([name, service]) => {
  const basePath = `/api/${name}`;
  const proxyMiddleware = proxy(service.url, {
    proxyReqPathResolver: req => {
      // Remove the service base path from the URL
      const origUrl = req.originalUrl;
      const path = origUrl.replace(new RegExp(`^${basePath}`), '');
      logger.debug(`Proxying request: ${req.method} ${origUrl} -> ${service.url}${path}`);
      return path || '/';
    },
    proxyErrorHandler: (err, res, next) => {
      logger.error(`Proxy error for ${name} service: ${err.message}`);
      if (err.code === 'ECONNREFUSED' || err.code === 'ECONNRESET') {
        return res.status(503).json({
          error: 'ServiceUnavailableError',
          message: `The ${name} service is currently unavailable`,
          statusCode: 503
        });
      }
      next(err);
    },
    proxyReqOptDecorator: (proxyReqOpts, srcReq) => {
      // Pass the user info to the downstream service if authenticated
      if (srcReq.auth) {
        proxyReqOpts.headers['X-User-Id'] = srcReq.auth.sub;
        proxyReqOpts.headers['X-User-Roles'] = srcReq.auth.roles ? srcReq.auth.roles.join(',') : '';
      }
      
      // Generate and pass correlation ID for request tracing
      const correlationId = srcReq.headers['x-correlation-id'] || `req-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      proxyReqOpts.headers['X-Correlation-Id'] = correlationId;
      
      return proxyReqOpts;
    }
  });

  // Apply rate limiting and authentication based on service config
  if (name === 'auth') {
    // Auth routes have special rate limiting
    router.use(basePath, authLimiter, proxyMiddleware);
  } else if (service.requiresAuth) {
    // Protected routes
    router.use(basePath, authenticate, proxyMiddleware);
  } else {
    // Public routes with optional authentication
    router.use(basePath, optionalAuthenticate, proxyMiddleware);
  }
});

module.exports = router;