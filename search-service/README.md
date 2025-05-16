# Search Service

The Search Service is a microservice component of the BlogBish platform that provides full-text search and suggestion capabilities for blog posts and comments.

## Features

- Full-text search across blog posts and comments
- Search suggestions for posts, comments, tags, and categories
- Filtering by status, tags, and categories
- Sorting by relevance or date
- Redis caching for improved performance
- Elasticsearch for powerful text search capabilities

## API Endpoints

### Search

```
POST /search
{
    "query": "search term",
    "type": "post|comment|all",
    "from": 0,
    "size": 10,
    "status": "published",
    "tags": ["tag1", "tag2"],
    "category": "category1",
    "sort_by": "relevance|date",
    "sort_order": "asc|desc"
}
```

### Suggestions

```
POST /suggest
{
    "query": "search term",
    "type": "post|comment|tag|category",
    "limit": 5,
    "status": "published"
}
```

### Index Post

```
POST /index/post
{
    "id": "post-id",
    "title": "Post Title",
    "content": "Post content...",
    "excerpt": "Post excerpt...",
    "author_id": 123,
    "author_name": "John Doe",
    "categories": ["category1", "category2"],
    "tags": ["tag1", "tag2"],
    "status": "published",
    "created_at": "2024-03-15T12:00:00Z",
    "updated_at": "2024-03-15T12:00:00Z",
    "published_at": "2024-03-15T12:00:00Z"
}
```

### Index Comment

```
POST /index/comment
{
    "id": "comment-id",
    "post_id": "post-id",
    "content": "Comment content...",
    "user_id": 123,
    "user_name": "John Doe",
    "status": "approved",
    "created_at": "2024-03-15T12:00:00Z",
    "updated_at": "2024-03-15T12:00:00Z"
}
```

## Environment Variables

- `PORT`: Server port (default: 8084)
- `ELASTICSEARCH_URL`: Elasticsearch server URL
- `REDIS_HOST`: Redis server host
- `REDIS_PORT`: Redis server port

## Dependencies

- Go 1.22 or later
- Elasticsearch 8.12.1
- Redis 7.0 or later

## Running the Service

1. Set up environment variables
2. Run with Docker:

   ```bash
   docker build -t blogbish/search-service .
   docker run -p 8084:8084 blogbish/search-service
   ```

3. Or run with Docker Compose:
   ```bash
   docker-compose up search-service
   ```

## Architecture

The service follows a clean architecture pattern with the following layers:

- `handler`: HTTP request handlers
- `service`: Business logic layer
- `repository`: Data access layer
- `models`: Data models and DTOs

## Caching

The service uses Redis for caching search and suggestion results with a 15-minute TTL. This helps reduce load on Elasticsearch and improves response times for frequently requested searches.
