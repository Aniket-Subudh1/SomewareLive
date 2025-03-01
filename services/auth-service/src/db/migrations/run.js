const { sequelize } = require('../index');
const db = require('../index').initModels();
const { hashPassword } = require('../../utils/password.util');
const logger = require('../../utils/logger');

async function runMigrations() {
  try {
    // Sync database models
    await sequelize.sync({ alter: true });
    logger.info('Database synchronized successfully');

    // Check if admin user exists
    const adminExists = await db.User.findOne({
      where: { email: 'admin@slidoclone.com' }
    });

    // Create admin user if not exists
    if (!adminExists) {
      const adminUser = await db.User.create({
        email: 'admin@slidoclone.com',
        password: await hashPassword('Admin@123'),
        firstName: 'Admin',
        lastName: 'User',
        role: 'admin',
        verified: true,
        active: true,
      });
      logger.info('Admin user created successfully');
    }

    logger.info('Migrations completed successfully');
    process.exit(0);
  } catch (error) {
    logger.error('Migration error:', error);
    process.exit(1);
  }
}


runMigrations();