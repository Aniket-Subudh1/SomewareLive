const { Sequelize } = require('sequelize');
const config = require('../config');
const logger = require('../utils/logger');


const sequelize = new Sequelize(
  config.database.database,
  config.database.username,
  config.database.password,
  {
    host: config.database.host,
    port: config.database.port,
    dialect: config.database.dialect,
    schema: config.database.schema,
    pool: {
      max: config.database.pool.max,
      min: config.database.pool.min,
      acquire: config.database.pool.acquire,
      idle: config.database.pool.idle,
    },
    logging: config.database.logging 
      ? (msg) => logger.debug(msg)
      : false,
  }
);


async function testConnection() {
  try {
    await sequelize.authenticate();
    logger.info('Database connection has been established successfully.');
    
    // Force sync all models
    logger.info('Syncing database models...');
    const syncResult = await sequelize.sync({ 
      force: true, 
      logging: console.log 
    });
    logger.info('Database models synced successfully!');
    
    return true;
  } catch (error) {
    logger.error('Unable to connect to the database:', error);
    console.error('Detailed error:', error);
    return false;
  }
}


function initModels() {
  // Make sure these paths are correct
  const User = require('../models/user.model')(sequelize);
  const Token = require('../models/token.model')(sequelize);

  // Define associations
  User.hasMany(Token, { foreignKey: 'userId', as: 'tokens' });
  Token.belongsTo(User, { foreignKey: 'userId', as: 'user' });

  return {
    User,
    Token,
    sequelize,
  };
}

module.exports = {
  sequelize,
  testConnection,
  initModels,
};