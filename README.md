# SomewareLive

A microservices-based clone of Slido with real-time audience interaction features.

## Features

- Anonymous Q&A with upvoting
- Live polls and quizzes
- Word clouds
- Real-time feedback
- Presentation software integration

## Architecture

This project uses a microservices architecture with:
- Node.js/Express for most services
- Go for high-performance services
- Kafka for event streaming
- MongoDB and PostgreSQL for data storage
- Redis for caching and pub/sub
- Socket.IO for real-time communication

## Getting Started

1. Clone the repository
2. Run `docker-compose up` to start development infrastructure
3. Navigate to each service directory and follow the README instructions

## Services

- API Gateway: Central entry point for API requests
- WebSocket Gateway: Handles real-time connections
- Auth Service: Authentication and authorization
- User Service: User management
- Event Service: Event and session management
- Question Service: Q&A functionality
- Poll Service: Live polling
- Quiz Service: Interactive quizzes
- WordCloud Service: Word cloud generation
- Feedback Service: User feedback collection
- Presentation Service: Presentation software integration
- Export Service: Data export functionality
- Notification Service: Email and push notifications
