# go-plate

Opinionated Go REST Backend Boilerplate. Using `go 1.22.0`.

As I've developed my Glassdoor clone [Compared](https://joselico.com/work/compared) and a few other smaller apps,
I've tinkered around with Go a bit. I've tried quite a few different project structures and techniques, experimenting to find what works best for me.

After all that trial and error, I decided to put together this Go backend boilerplate project.
It is supposed to serve as a distillation of all the lessons I've learned and a starting point that incorporates the patterns and practices I've found to be most effective.
Something that allows me (and maybe you) to spin up a new project fast and be productive.

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
- [ ] JWT generation and validation middleware
- [x] Rate Limiting (Using Token Bucket & Leaky Bucket, implemented with in-memory storage for local rate limiting, and Redis for distributed systems across multiple server instances)
- [x] CI/CD pipeline for GCP Cloud Run service
- [ ] CI/CD pipeline for AWS services
- [ ] OAuth (Google)
- [x] Example endpoints to showcase functionality and use
- [ ] Documentation generation with [Swagger](https://swagger.io/)

## How to use

For local development, copy `.env.example` to `.env` with your variables or otherwise inject them.
For docker, a sample `docker-compose.yml` is at the root with these variables.

At the root of the project is a Makefile with some util commands to help development.

### Go commands

`make dev` -> go run with live reload thru [air](https://github.com/air-verse/air)

`make run` -> go run

`make build` -> go build

### Docker commands

`make up` -> Spins up docker container(s) with docker-compose

`make down` -> Spins down docker container(s) with docker-compose

`make down -v` -> Spins down docker container(s) and volume(s) with docker-compose

### Migrations

**Local**

`make migration name=<migration_name>` -> Create a new migration

`make migrate-up` -> Runs up migrations

`make migrate-down` -> Runs down migrations

**Docker**

`migrate-up-docker` -> Runs up migrations

`migrate-down-docker` -> Runs down migrations
