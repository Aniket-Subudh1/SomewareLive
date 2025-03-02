# User Service

This service manages user profiles, organizations, and teams for the Slido Clone platform.

## Features

- User profile management
- Team management
- Organization management
- Integration with Auth Service
- Kafka event streaming
- MongoDB data storage

## Technology Stack

- Go (Golang)
- Gin web framework
- MongoDB
- Kafka for event streaming
- JWT for token validation

## Getting Started

### Prerequisites

- Go 1.21 or higher
- MongoDB
- Kafka

### Installation

1. Clone the repository
2. Install dependencies:

```bash
go mod download
```

3. Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

4. Edit the `.env` file with your configuration settings.

### Development

To run the service in development mode:

```bash
go run main.go
```

### Building

To build the service:

```bash
go build -o user-service .
```

### Testing

To run tests:

```bash
go test ./...
```

## API Documentation

### Base URL

All API routes are accessible through:

```
/api
```

### Health Check

- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed health check with dependency status

### User Endpoints

- `GET /api/me` - Get current user
- `PUT /api/me` - Update current user
- `GET /api/users` - List users
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create a new user
- `PUT /api/users/:id` - Update a user
- `DELETE /api/users/:id` - Delete a user
- `POST /api/users/:id/activate` - Activate a user
- `POST /api/users/:id/deactivate` - Deactivate a user

### Team Endpoints

- `GET /api/teams` - List teams
- `GET /api/teams/:id` - Get team by ID
- `POST /api/teams` - Create a new team
- `PUT /api/teams/:id` - Update a team
- `DELETE /api/teams/:id` - Delete a team
- `GET /api/teams/:id/members` - List team members
- `POST /api/teams/:id/members` - Add a member to a team
- `PUT /api/teams/:id/members/:userId` - Update a team member
- `DELETE /api/teams/:id/members/:userId` - Remove a member from a team

### Organization Endpoints

- `GET /api/organizations` - List organizations
- `GET /api/organizations/:id` - Get organization by ID
- `POST /api/organizations` - Create a new organization
- `PUT /api/organizations/:id` - Update an organization
- `DELETE /api/organizations/:id` - Delete an organization
- `GET /api/organizations/:id/members` - List organization members
- `POST /api/organizations/:id/members` - Add a member to an organization
- `PUT /api/organizations/:id/members/:userId` - Update an organization member
- `DELETE /api/organizations/:id/members/:userId` - Remove a member from an organization

## Event Schema

### Published Events

- `user.created` - When a new user is created
- `user.updated` - When a user is updated
- `user.deleted` - When a user is deleted
- `user.activated` - When a user is activated
- `user.deactivated` - When a user is deactivated
- `team.created` - When a new team is created
- `team.updated` - When a team is updated
- `team.deleted` - When a team is deleted
- `team.member.added` - When a member is added to a team
- `team.member.updated` - When a team member is updated
- `team.member.removed` - When a member is removed from a team

### Consumed Events

- `auth.user.created` - When a user is created in the Auth Service

## Container Support

Build the Docker image:

```bash
docker build -t slido-clone/user-service .
```

Run the container:

```bash
docker run -p 8001:8001 --env-file .env slido-clone/user-service
```

## Configuration

The service is configured through environment variables. See `.env.example` for all available options.