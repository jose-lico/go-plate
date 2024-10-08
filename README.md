# go-plate

Opinionated Go REST Backend Boilerplate. Using `go 1.22.0`.

As I've developed my Glassdoor clone [Compared](https://joselico.com/work/compared) and a few other smaller apps,
I've tinkered around with Go a bit. I've tried quite a few different project structures and techniques, experimenting to find what works best for me.

After all that trial and error, I decided to put together this Go backend boilerplate project.
It is supposed to serve as a distillation of all the lessons I've learned and a starting point that incorporates the patterns and practices I've found to be most effective.
Something that allows me (and maybe you) to spin up a new project fast and be productive.

It also includes some examples of how to use it and how I like to structure my backends.

## Install

`go get github.com/jose-lico/go-plate`

## Overview & Features

- [x] Routing with [chi](https://github.com/go-chi/chi) (Lightweight, 100% compatible with net/http)
- [x] Environment variable management using [godotenv](https://github.com/joho/godotenv)
- [x] CORS handling with [cors](https://github.com/rs/cors)
- [x] Containerization with Docker
- [x] API versioning via URL paths (`/api/v2/posts`) and custom headers (`X-API-Version: v1`)
- [x] Request payload validation using [validator](https://github.com/go-playground/validator)
- [x] SQL (PostgreSQL) integration with [gorm](https://github.com/go-gorm/gorm) ORM
- [x] Database schema management with [migrate](https://github.com/golang-migrate/migrate) for version-controlled and reproducible migrations
- [x] Redis caching implementation with [go-redis](https://github.com/redis/go-redis)
- [x] Secure password hashing and verification
- [x] Authentication middleware with session management (Redis-backed)
- [x] Rate Limiting (Implemented Token Bucket & Sliding Window, with in-memory storage for local rate limiting, and Redis for distributed systems across multiple server instances)
- [x] CI/CD pipeline for GCP Cloud Run service
- [x] CI/CD pipeline for AWS App Runner service
- [x] Example endpoints to showcase functionality and use
- [x] Documentation generation with [Swagger](https://swagger.io/)

## How to use

To get started, create a new `APIServer` object:

```go
import "github.com/jose-lico/go-plate/api"

func main() {
	cfg := config.NewAPIConfig()
	api := api.NewAPIServer(cfg)
	api.UseDefaultMiddleware()

	api.Run()
}
```

Register some endpoints:

```go
import "github.com/jose-lico/go-plate/api"

func main() {
	...

	api.Router.Get("/", hello)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
```

Connect to Redis and PostgreSQL:

```go
import "github.com/jose-lico/go-plate/database"

func main() {
	redisCFG := config.NewRedisConfig()
	redis, err := database.NewRedis(redisCFG)
	if err != nil {
		...
	}

	sqlCFG := config.NewSQLConfig()
	sql, err := database.NewSQLGormDB(sqlCFG)
	if err != nil {
		...
	}
}
```

Create some server-wide and endpoint specfic middleware:

```go
import (
	"github.com/jose-lico/go-plate/api"
	"github.com/jose-lico/go-plate/middleware"
	"github.com/jose-lico/go-plate/ratelimiting"
)

func main() {
	...

	api.Router.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(ratelimiting.NewInMemoryTokenBucket(0.05, 3, 10*time.Minute)))

		r.Get("/rate-limited", hello)
	})
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
```

## Structure

```
.
├── api
│   ├── api.go				// Server
├── auth
│   ├── password.go			// Hash and compare password
│   ├── token.go			// Generate random 32 byte token
├── config
│   ├── api_config.go			// API configuration
│   ├── redis_config.go			// Redis configuration
│   └── sql_config.go			// SQL configuration
├── database
│   ├── redis.go			// Redis interface, implemented with go-redis
│   └── sql_gorm.go			// SQL interface, using gorm
├── middleware
│   ├── rate_limit.go			// Rate litiming with algorithm of choice
│   ├── session.go			// Session with redis
│   └── versioning.go			// API versioning
├── ratelimiting
│   ├── mem_sliding_window.go		// In-memory Sliding Window
│   ├── mem_token_bucket.go		// In-memory Token Bucket
│   ├── rate_limiter.go			// Rate Limiter interface
│   └── redis_token_bucket.go		// Redis Token Bucket
├── utils
│   └── utils.go			// Utils functions
```

## Examples

To run the examples in `/examples`, copy `.env.example` to `.env` with your variables or otherwise inject them.
For docker, a sample `docker-compose.yml` is at the root with these variables.

At the root of the project there is a Makefile with some util commands to run these examples.

### Go commands

`make dev` -> go run with live reload thru [air](https://github.com/air-verse/air)

`make run` -> go run

`make build` -> go build

`make gen-docs` -> generates API docs

### Docker commands

`make up` -> Spins up docker containers with docker-compose

`make down` -> Spins down docker containers with docker-compose

`make down -v` -> Spins down docker containers and volumes with docker-compose

### Migrations

**Local**

`make migration name=<migration_name>` -> Create a new migration

`make migrate-up` -> Runs up migrations

`make migrate-down` -> Runs down migrations

**Docker**

`migrate-up-docker` -> Runs up migrations

`migrate-down-docker` -> Runs down migrations
