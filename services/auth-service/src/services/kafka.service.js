const { Kafka } = require('kafkajs');
const config = require('../config');
const logger = require('../utils/logger');

const kafka = new Kafka({
  clientId: config.kafka.clientId,
  brokers: config.kafka.brokers,
  retry: {
    initialRetryTime: 100,
    retries: 8
  }
});

const producer = kafka.producer({
  allowAutoTopicCreation: true,
  transactionTimeout: 30000
});

let isProducerConnected = false;


const connectProducer = async () => {
  if (!isProducerConnected) {
    try {
      await producer.connect();
      isProducerConnected = true;
      logger.info('Successfully connected to Kafka producer');
    } catch (error) {
      logger.error(`Failed to connect to Kafka producer: ${error.message}`);
      throw error;
    }
  }
};

const disconnectProducer = async () => {
  if (isProducerConnected) {
    try {
      await producer.disconnect();
      isProducerConnected = false;
      logger.info('Disconnected from Kafka producer');
    } catch (error) {
      logger.error(`Failed to disconnect from Kafka producer: ${error.message}`);
      throw error;
    }
  }
};


const sendMessage = async (topic, message, key = null) => {
  if (!isProducerConnected) {
    await connectProducer();
  }

  try {
    const messageObj = {
      key: key || message.id || String(Date.now()),
      value: JSON.stringify(message),
      headers: {
        'source': config.kafka.clientId,
        'timestamp': String(Date.now()),
        'correlationId': message.correlationId || `msg-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
      }
    };

    await producer.send({
      topic,
      messages: [messageObj]
    });

    logger.debug(`Message sent to topic ${topic}:`, { key: messageObj.key });
  } catch (error) {
    logger.error(`Error sending message to topic ${topic}: ${error.message}`);
    throw error;
  }
};


const sendAuthEvent = async (eventType, data) => {
  const message = {
    type: eventType,
    timestamp: new Date().toISOString(),
    data
  };

  return sendMessage(config.kafka.topics.authEvents, message, data.userId || data.id);
};


const sendUserEvent = async (eventType, data) => {
  const message = {
    type: eventType,
    timestamp: new Date().toISOString(),
    data
  };

  return sendMessage(config.kafka.topics.userEvents, message, data.userId || data.id);
};

module.exports = {
  connectProducer,
  disconnectProducer,
  sendMessage,
  sendAuthEvent,
  sendUserEvent,
  kafka, 
};