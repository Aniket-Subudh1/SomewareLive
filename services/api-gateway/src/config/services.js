module.exports = {
    auth: {
      url: process.env.AUTH_SERVICE_URL || 'http://localhost:3001',
      requiresAuth: false,
    },
    users: {
      url: process.env.USER_SERVICE_URL || 'http://localhost:8001',
      requiresAuth: true,
    },
    events: {
      url: process.env.EVENT_SERVICE_URL || 'http://localhost:3002',
      requiresAuth: true,
    },
    questions: {
      url: process.env.QUESTION_SERVICE_URL || 'http://localhost:8002',
      requiresAuth: true,
    },
    polls: {
      url: process.env.POLL_SERVICE_URL || 'http://localhost:3003',
      requiresAuth: true,
    },
    quizzes: {
      url: process.env.QUIZ_SERVICE_URL || 'http://localhost:8003',
      requiresAuth: true,
    },
    wordcloud: {
      url: process.env.WORDCLOUD_SERVICE_URL || 'http://localhost:3004',
      requiresAuth: true,
    },
    feedback: {
      url: process.env.FEEDBACK_SERVICE_URL || 'http://localhost:8004',
      requiresAuth: true,
    },
    presentation: {
      url: process.env.PRESENTATION_SERVICE_URL || 'http://localhost:3005',
      requiresAuth: true,
    },
    export: {
      url: process.env.EXPORT_SERVICE_URL || 'http://localhost:8005',
      requiresAuth: true,
    },
    notification: {
      url: process.env.NOTIFICATION_SERVICE_URL || 'http://localhost:3006',
      requiresAuth: true,
    },
  };