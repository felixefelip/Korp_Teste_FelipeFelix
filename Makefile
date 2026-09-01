up:
	docker compose up -d

down:
	docker compose down

inventory-run:
	docker compose exec -e GO_ENV=development inventory go run cmd/main.go

inventory-test:
	docker compose exec -e GO_ENV=test inventory go test -count=1 -p 1 ./...

billing-run:
	docker compose exec -e GO_ENV=development billing go run cmd/main.go

billing-test:
	docker compose exec -e GO_ENV=test billing go test -count=1 -p 1 ./...

finance-setup:
	docker compose exec finance sh -c "cp -n .env.example .env; php artisan key:generate && npm run build"

finance-run:
	docker compose exec finance sh -c "php artisan migrate --force && php artisan serve --host=0.0.0.0 --port=8002"

finance-relay:
	docker compose exec finance php artisan finance:relay

finance-consume:
	docker compose exec finance php artisan finance:consume

finance-vite:
	docker compose exec finance npm run dev -- --host 0.0.0.0

finance-test:
	docker compose exec -e DB_DATABASE=finance_test finance php artisan test

front-test:
	docker compose exec frontend npm test -- --watch=false
