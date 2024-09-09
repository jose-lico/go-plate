# go-plate

Opinionated Go REST Backend Boilerplate. Using `go 1.22.0`.

As I've developed my Glassdoor [Compared](https://joselico.com/work/compared) clone and a few other smaller apps, 
I've tinkered around with Go a bit. I've tried quite a few different project structures and techniques, experimenting to find what works best for me.

After all that trial and error, I decided to put together this Go backend boilerplate project. 
It is supposed to serve as a distillation of all the lessons I've learned and a starting point that incorporates the patterns and practices I've found to be most effective.
Something that allows me (and maybe you) to spin up a new project fast and be productive. 

## Overview & Features

- [x] Using [chi](https://github.com/go-chi/chi) router (Lightweight, 100% compatible with net/http)
- [x] Loading local .env variables with [godotenv](https://github.com/joho/godotenv)
- [x] CORS management with [cors](https://github.com/rs/cors)
- [x] Using Docker for containerization
- [x] Endpoint versioning with URL or Headers (`/api/v1/user` or `X-API-Version: v1`)
- [ ] Input validation with [validator](https://github.com/go-playground/validator)
- [x] Connection to Redis cache using [go-redis](https://github.com/redis/go-redis)
- [ ] ORM and connection to PostgreSQL using [gorm](https://github.com/go-gorm/gorm) 
- [ ] Auth functions & middleware to generate and compare hashes
- [ ] JWT functions & middleware to generate and validate JWTs
- [ ] Rate Limiting
- [ ] CI/CD pipeline for GCP services
- [ ] CI/CD pipeline for AWS services
- [ ] OAuth (Google)
- [ ] Some example endpoints and tests
- [ ] Documentation generation with [Swagger](https://swagger.io/)

## How to use

For local development, copy `.env.example` to `.env` with your variables or otherwise inject them.
For docker, a sample `docker-compose.yml` is at the root with these variables.

At the root of the project is a Makefile with some util commands to help development.

**Go commands**

`make dev` -> go run with live reload thru [air](https://github.com/air-verse/air)

`make run` -> go run

`make build` -> go build

**Docker commands**

`make up` -> Spins up docker container(s) with docker-compose

`make down` -> Spins down docker container(s) with docker-compose

`make down -v` -> Spins down docker container(s) and volume(s) with docker-compose
