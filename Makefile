.PHONY: test

test:
	docker compose up -d
	echo "Waiting for database to start"
	sleep 15
	-go run internal/main.go
	docker compose down