# ================================
#          Go Commands
# ================================

dev:
	ENV=LOCAL air

run:
	ENV=LOCAL go run ./examples/main

build:
	go build -o bin/main ./examples/main

test:
	go test -v ./...

# ================================
#         Docker Commands
# ================================

up:
	docker-compose down
	docker-compose build
	docker-compose up -d

down:
	docker-compose down

down-v:
	docker-compose down -v

# ================================
#           Migrations
# ================================

migration:
	migrate create -ext sql -dir examples/migrate/migrations $(name)

migrate-up:
	ENV=LOCAL go run examples/migrate/main.go up examples/migrate/migrations

migrate-down:
	ENV=LOCAL go run examples/migrate/main.go down examples/migrate/migrations

# ================================
#         Docker Migrations  
# ================================

migrate-up-docker:
	docker-compose build migrate
	docker-compose run --rm migrate up .

migrate-down-docker:
	docker-compose build migrate
	docker-compose run --rm migrate down .