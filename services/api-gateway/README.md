# API Gateway Service

This service functions as the entry point for all API requests to the SomewareLive platform. It handles request routing, authentication, rate limiting, and load balancing.

## Features

- Request routing to appropriate microservices
- JWT authentication and authorization
- Rate limiting to prevent abuse
- Request logging and correlation IDs for tracing
- Health checks for all services
- CORS configuration

## Technology Stack

- Node.js
- Express.js
- express-http-proxy for request proxying
- express-jwt for JWT authentication
- pino for logging

## Getting Started

### Prerequisites

- Node.js (v16+)
- npm or yarn

### Installation

1. Install dependencies:
```bash
npm install
```

2. Create a `.env` file based on `.env.example`:
```bash
cp .env.example .env
```

3. Edit the `.env` file with your configuration settings.

### Development

To run the service in development mode with auto-reload:

```bash
npm run dev
```

### Production

To run the service in production mode:

```bash
npm start
```

## API Documentation

### Base URL

All API routes are accessible through the gateway at:

```
/api/{service-name}/...
```

### Available Services

- `/api/auth` - Authentication service
- `/api/users` - User management
- `/api/events` - Event management
- `/api/questions` - Q&A functionality
- `/api/polls` - Live polling
- `/api/quizzes` - Interactive quizzes
- `/api/wordcloud` - Word cloud generation
- `/api/feedback` - User feedback collection
- `/api/presentation` - Presentation software integration
- `/api/export` - Data export functionality
- `/api/notification` - Notifications and alerts

### Health Checks

- `GET /health` - Basic health check for the API Gateway
- `GET /health/services` - Health check for all downstream services

## Container Support

Build the Docker image:

```bash
docker build -t SomewareLive/api-gateway .
```

Run the container:

```bash
docker run -p 3000:3000 --env-file .env SomewareLive/api-gateway
```

## Configuration

The service is configured through environment variables. See `.env.example` for all available options.

## Metrics and Monitoring

The service logs all requests and errors using pino. In production, you should configure log forwarding to a centralized logging system.