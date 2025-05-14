# BlogBish - Microservices Blog Platform

A modern, cloud-native blogging platform built with Go microservices architecture.

## Project Structure

```
blogbish/
├── auth-service/           # Authentication and user management
├── post-service/          # Blog post and category management
├── comment-service/       # Comment management (coming soon)
├── media-service/        # Media handling (coming soon)
├── analytics-service/    # Analytics and metrics (coming soon)
├── docker-compose.yml    # Docker composition for all services
├── prometheus.yml        # Prometheus monitoring configuration
├── Makefile             # Build and management commands
├── go.work              # Go workspace configuration
└── README.md
```

## Services

### Currently Implemented:

1. **Auth Service** (Port: 8080)

   - User registration and authentication
   - JWT token management
   - Role-based access control

2. **Post Service** (Port: 8081)
   - Blog post CRUD operations
   - Category management
   - Tag management
   - Post search and filtering

### Coming Soon:

3. **Comment Service** (Port: 8082)
4. **Media Service** (Port: 8083)
5. **Analytics Service** (Port: 8084)

## Tech Stack

- **Backend**: Go (Golang)
- **Framework**: Chi (lightweight HTTP router)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: RabbitMQ (coming soon)
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker
- **Monitoring**: Prometheus & Grafana
- **Logging**: ELK Stack (coming soon)

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Make (for using Makefile commands)
- PostgreSQL 15
- Redis 7

## Getting Started

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/blogbish.git
   cd blogbish
   ```

2. Start all services:

   ```bash
   make docker-up
   ```

3. Run database migrations:

   ```bash
   make migrate-up
   ```

4. Access the services:
   - Auth Service: http://localhost:8080
   - Post Service: http://localhost:8081
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000

## Development

### Available Make Commands

- `make build` - Build all services
- `make test` - Run tests for all services
- `make clean` - Clean build artifacts
- `make docker-up` - Start all services
- `make docker-down` - Stop all services
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback database migrations
- `make run-auth-service` - Run auth service locally
- `make run-post-service` - Run post service locally

### Adding a New Service

1. Create a new directory for your service
2. Copy the basic service structure from existing services
3. Add the service to:
   - docker-compose.yml
   - go.work
   - Makefile (SERVICES variable)
   - prometheus.yml (if metrics are needed)

## API Documentation

### Auth Service Endpoints

- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login user
- `GET /auth/me` - Get current user info (Protected)

### Post Service Endpoints

- `POST /posts` - Create a new post (Protected)
- `GET /posts` - List posts with filtering
- `GET /posts/{slug}` - Get post by slug
- `PUT /posts/{slug}` - Update post (Protected)
- `DELETE /posts/{slug}` - Delete post (Protected)
- `GET /categories` - List categories
- `POST /categories` - Create category (Protected)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
