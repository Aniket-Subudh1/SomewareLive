# Auth Service

This service handles user authentication and authorization for the SomewareLive platform. It manages user accounts, authentication tokens, and access control.

## Features

- User registration and login
- JWT-based authentication
- Role-based access control
- Password management
- User profile management
- Integration with Kafka for auth events

## Technology Stack

- Node.js
- Express.js
- PostgreSQL with Sequelize ORM
- JWT for tokens
- bcrypt for password hashing
- Kafka for event streaming

## Getting Started

### Prerequisites

- Node.js (v16+)
- PostgreSQL
- Kafka

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

4. Create the database:
```bash
createdb somewarelive_auth
```

5. Run database migrations:
```bash
npm run db:migrate
```

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

### Authentication Endpoints

- `POST /register` - Register a new user
- `POST /login` - Login a user
- `POST /logout` - Logout a user
- `POST /refresh-token` - Refresh access token
- `GET /verify` - Verify an access token
- `GET /me` - Get current user details

### User Management Endpoints

- `GET /users` - Get all users (Admin only)
- `GET /users/:id` - Get user by ID (Admin only)
- `POST /users` - Create a user (Admin only)
- `PUT /users/:id` - Update a user (Admin only)
- `DELETE /users/:id` - Deactivate a user (Admin only)
- `PUT /users/profile` - Update current user's profile
- `POST /users/change-password` - Change current user's password
- `DELETE /users/profile` - Deactivate current user's account

### Health Check Endpoint

- `GET /health` - Service health check

## Security

This service implements several security best practices:

1. Password hashing using bcrypt
2. JWT for token-based authentication
3. HTTP-only cookies for refresh tokens
4. CORS protection
5. Helmet for security headers
6. Rate limiting to prevent abuse
7. Input validation
8. Token blacklisting

## Event Publishing

The service publishes the following Kafka events:

- `user.created` - When a new user is created
- `user.updated` - When a user is updated
- `user.deactivated` - When a user is deactivated
- `user.logged_in` - When a user logs in
- `user.logged_out` - When a user logs out
- `user.token_refreshed` - When a token is refreshed
- `user.password_changed` - When a password is changed

## Container Support

Build the Docker image:

```bash
docker build -t somewarelive/auth-service .
```

Run the container:

```bash
docker run -p 3001:3001 --env-file .env somewarelive/auth-service
```

## Database Migrations

Run database migrations:

```bash
npm run db:migrate
```

This will:
1. Create the necessary tables
2. Create the admin user if it doesn't exist