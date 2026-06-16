.PHONY: seed run build es-reset es-populate

seed:
	go run cmd/seed/main.go

run:
	go run main.go

build:
	go build -o bin/app main.go

es-reset:
	go run cmd/seed/es_reset/main.go

es-populate:
	go run cmd/seed/es_populate/main.go
