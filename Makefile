.PHONY: run test clean build

include .env

export

run:
	go run main.go

test:
	go test ./...

build:
	go build -o totp-server main.go

clean:
	rm -f totp-server

totp-generate:
	curl -X POST http://localhost:$(SERVER_PORT)/totp/generate \
		-H "Content-Type: application/json" \
		-d '{"account_name": "user@example.com"}'

totp-verify:
	curl -X POST http://localhost:$(SERVER_PORT)/totp/verify \
		-H "Content-Type: application/json" \
		-d '{"account_name": "user@example.com", "code": "123456"}'

health:
	curl http://localhost:$(SERVER_PORT)/health
