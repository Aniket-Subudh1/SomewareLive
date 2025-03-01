const { DataTypes } = require('sequelize');

module.exports = (sequelize) => {
  const Token = sequelize.define('Token', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true,
    },
    userId: {
      type: DataTypes.UUID,
      allowNull: false,
      references: {
        model: 'users',
        key: 'id',
      },
    },
    token: {
      type: DataTypes.TEXT, 
      allowNull: false,
    },
    type: {
      type: DataTypes.STRING(20),
      allowNull: false,
      validate: {
        isIn: [['refresh', 'reset', 'verification']]
      }
    },
    expiresAt: {
      type: DataTypes.DATE,
      allowNull: false,
    },
    blacklisted: {
      type: DataTypes.BOOLEAN,
      defaultValue: false,
    },
  }, {
    tableName: 'tokens',
    timestamps: true,
    indexes: [
      {
        unique: true,
        fields: ['token'],
      },
      {
        fields: ['userId'],
      },
      {
        fields: ['type'],
      },
      {
        fields: ['expiresAt'],
      },
    ],
  });

  // Class methods
  Token.createRefreshToken = async function(userId, token, expiresAt) {
    try {
      return await Token.create({
        userId,
        token,
        type: 'refresh',
        expiresAt,
        blacklisted: false
      });
    } catch (error) {
      console.error('Error creating refresh token:', error);
      throw error;
    }
  };

  Token.createResetToken = async function(userId, token, expiresAt) {
    try {
      return await Token.create({
        userId,
        token,
        type: 'reset',
        expiresAt,
        blacklisted: false
      });
    } catch (error) {
      console.error('Error creating reset token:', error);
      throw error;
    }
  };

  Token.createVerificationToken = async function(userId, token, expiresAt) {
    try {
      return await Token.create({
        userId,
        token,
        type: 'verification',
        expiresAt,
        blacklisted: false
      });
    } catch (error) {
      console.error('Error creating verification token:', error);
      throw error;
    }
  };

  Token.findValidToken = async function(token, type) {
    try {
      const now = new Date();
      return Token.findOne({
        where: {
          token,
          type,
          expiresAt: {
            [sequelize.Sequelize.Op.gt]: now,
          },
          blacklisted: false,
        },
      });
    } catch (error) {
      console.error('Error finding valid token:', error);
      return null;
    }
  };

  // Instance methods
  Token.prototype.blacklist = async function() {
    this.blacklisted = true;
    return this.save();
  };

  return Token;
};