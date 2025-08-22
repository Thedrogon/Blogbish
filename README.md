# BlogBish - Microservices Blog Platform

A modern, cloud-native blogging platform built with Go microservices architecture.

## Project Structure

```
blogbish/
├── auth-service/           # Authentication and user management
├── post-service/           # Blog post and category management
├── comment-service/        # Comment management
├── media-service/          # Media handling
├── search-service/         # Full-text search and suggestions
├── docker-compose.yml      # Docker composition for all services
├── prometheus.yml          # Prometheus monitoring configuration
├── Makefile                # Build and management commands
├── go.work                 # Go workspace configuration
└── README.md
```

## Services

### Currently Implemented

1. **Auth Service** (Port: 8080)

   - User registration and authentication
   - JWT token management
   - Role-based access control

2. **Post Service** (Port: 8081)
   - Blog post CRUD operations
   - Category management
   - Filtering & pagination

3. **Media Service** (Port: 8082)
   - Upload/download files to MinIO (S3-compatible)
   - Metadata management, Redis caching

4. **Comment Service** (Port: 8083)
   - Comment CRUD, like, report, moderate
   - WebSocket hub for live comments

5. **Search Service** (Port: 8084)
   - Full-text search and suggestions (Elasticsearch)
   - Index posts and comments

### Planned

- Analytics Service

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
   - Media Service: http://localhost:8082
   - Comment Service: http://localhost:8083
   - Search Service: http://localhost:8084

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

### Auth Service Endpoints (Port 8080)

- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login user
- `GET /auth/me` - Get current user info (Protected)

### Post Service Endpoints (Port 8081)

- `POST /posts` - Create a new post (Protected)
- `GET /posts` - List posts with filtering and pagination
- `GET /posts/{id}` - Get post by ID
- `PUT /posts/{id}` - Update post by ID (Protected)
- `DELETE /posts/{id}` - Delete post by ID (Protected)
- `GET /categories` - List categories
- `POST /categories` - Create category (Protected)
- `GET /categories/{slug}` - Get category by slug
- `PUT /categories/{slug}` - Update category by slug (Protected)
- `DELETE /categories/{slug}` - Delete category by slug (Protected)

### Media Service Endpoints (Port 8082)

- `POST /api/v1/media/upload` - Upload a file (Protected)
- `GET /api/v1/media/{id}` - Get media metadata
- `GET /api/v1/media/{id}/download` - Download file
- `PUT /api/v1/media/{id}/metadata` - Update metadata (Protected)
- `DELETE /api/v1/media/{id}` - Delete file (Protected)

### Comment Service Endpoints (Port 8083)

- `POST /comments` - Create a comment (Protected)
- `GET /comments` - List comments (query: `post_id`, `user_id`, `status`, `parent_id`, `page`, `page_size`)
- `GET /comments/{id}` - Get comment by ID
- `PUT /comments/{id}` - Update comment (Protected, author only)
- `DELETE /comments/{id}` - Delete comment (Protected, author only)
- `POST /comments/{id}/like` - Like a comment
- `POST /comments/{id}/report` - Report a comment
- `PUT /comments/{id}/moderate?status={approved|rejected|pending}` - Moderate comment
- `GET /ws?post_id={postId}` - WebSocket for live comments (Protected)

### Search Service Endpoints (Port 8084)

- `POST /search` - Search posts/comments
- `POST /suggest` - Get search suggestions
- `POST /index/post` - Index a post
- `POST /index/comment` - Index a comment

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
