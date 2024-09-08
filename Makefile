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
#          Docker Commands
# ================================

up:
	docker-compose down
	docker-compose build
	docker-compose up -d

down:
	docker-compose down