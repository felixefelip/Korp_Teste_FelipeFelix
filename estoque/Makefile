# Requer os containers no ar: docker compose up -d
# O ambiente vem de GO_ENV, e cada um usa seu proprio banco.

# Sobe a API em development (banco go_api_development).
run:
	docker compose exec -e GO_ENV=development go-app go run cmd/main.go

# Roda os testes contra o banco de teste (go_api_test).
# -count=1 desliga o cache: os testes dependem do estado do Postgres, que o Go
# nao enxerga, entao um resultado cacheado passaria ate com o banco fora do ar.
# -p 1 roda um pacote por vez: todos compartilham o mesmo banco, e o TRUNCATE
# de um pacote apagaria os dados de outro rodando em paralelo.
test:
	docker compose exec -e GO_ENV=test go-app go test -count=1 -p 1 ./...
