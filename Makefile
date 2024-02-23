.PHONY: test

test:
	docker compose up -d
	echo "Waiting for database to start"
	sleep 15
	echo "======= Test without lenient ======="
	-go run internal/main.go
	echo "======= Test with lenient ======="
	-go run internal/main.go -l true
	docker compose down