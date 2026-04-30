.PHONY: seed run build

seed:
	go run cmd/seed/main.go

run:
	go run main.go

build:
	go build -o bin/app main.go
