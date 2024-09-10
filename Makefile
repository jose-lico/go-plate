# ================================
#          Go Commands
# ================================

dev:
	ENV=LOCAL air

run:
	ENV=LOCAL go run ./cmd/main

build:
	go build -o bin/main ./cmd/main

# ================================
#          Migrations
# ================================

migration:
	migrate create -ext sql -dir cmd/migrate/migrations $(name)

migrate-up:
	ENV=LOCAL go run cmd/migrate/main.go up

migrate-down:
	ENV=LOCAL go run cmd/migrate/main.go down

# ================================
#          Docker Commands
# ================================

up:
	docker-compose down
	docker-compose build
	docker-compose up -d

down:
	docker-compose down

down-v:
	docker-compose down -v