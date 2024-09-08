dev:
	ENV=LOCAL air

run:
	ENV=LOCAL go run ./cmd/main

build:
	go build -o bin/main ./cmd/main