up:
	docker compose up -d

down:
	docker compose down

run:
	docker compose exec -e GO_ENV=development inventory go run cmd/main.go

test:
	docker compose exec -e GO_ENV=test inventory go test -count=1 -p 1 ./...

front-test:
	docker compose exec frontend npm test -- --watch=false
