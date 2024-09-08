# go-plate

Opinionated Go REST Backend Example

Using `go 1.22.0`

## How to use

Copy `.env.example` to `.env` with your variables or otherwise inject them.

`make dev` -> live reload thru [air](https://github.com/air-verse/air)

`make run` -> go run

`make build` -> go build

**Docker commands**

`make up` -> Spins up docker container(s) with docker-compose

`make down` -> Spins down docker container(s) with docker-compose

## Features

- Uses chi router
- Using docker for containerization
- Version endpoints with URL or Headers
- Connection to redis cache
