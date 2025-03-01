const pino = require('pino');
const config = require('../config');

const logger = pino({
  level: config.logging.level,
  transport: {
    target: 'pino-pretty',
    options: {
      colorize: true,
      translateTime: 'SYS:standard',
    },
  },
  base: {
    service: 'api-gateway',
  },
});

module.exports = logger;