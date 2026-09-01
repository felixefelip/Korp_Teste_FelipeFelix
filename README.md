# Projeto técnico: Sistema de emissão de Notas Fiscais

Dois microsserviços em Go, um em Laravel, um frontend Angular e RabbitMQ entre
eles:

- **inventory** (`:8000`) — catálogo de produtos e razão de movimentações de estoque.
- **billing** (`:8001`) — notas fiscais, seus itens e a DANFE.
- **finance** (`:8002`) — Laravel + Livewire, ainda no esqueleto.
- **frontend** (`:4200`) — Angular, com proxy para as duas APIs.

Cada serviço tem seu próprio Postgres e nenhum acessa o banco do outro. O
fechamento de uma nota baixa o estoque por evento, numa saga coreografada com
compensação.

- [docs/arquitetura.md](docs/arquitetura.md) — o desenho da integração e as
  decisões por trás dele.

## Detalhamento técnico

O checklist do desafio e as respostas às perguntas de detalhamento estão em
[docs/detalhamento-tecnico.md](docs/detalhamento-tecnico.md).

## Como rodar

```bash
cp .env.example .env      # opcional, só para o preenchimento por IA
docker compose up -d
```

Esse comando sobe os bancos, o RabbitMQ e o frontend. Os demais serviços sobem
como container ocioso — o processo é iniciado à mão, cada um no seu terminal:

```bash
make inventory-run
make billing-run
make finance-run
```

O `finance` guarda a `APP_KEY` no próprio `.env`, que não vai para o
repositório. Em clone novo, antes do primeiro `make finance-run`:

```bash
make finance-setup
```

Isso cria o `.env` a partir do exemplo, gera a chave e compila os assets. Para
hot reload do Vite durante o desenvolvimento, `make finance-vite` num terminal
à parte.

A mensageria do finance roda em dois processos próprios, cada um no seu
terminal — o relay publica o outbox, o consumidor lê a fila:

```bash
make finance-relay
make finance-consume
```

A aplicação fica em http://localhost:4200 e o painel do RabbitMQ em
http://localhost:15672 (`guest` / `guest`).

Os bancos são criados e migrados na subida de cada serviço, então não há passo
de migração separado.

## Testes

```bash
make inventory-test
make billing-test
make finance-test
make front-test
```

Nenhum serviço mocka a camada de persistência: os testes rodam contra um
Postgres de verdade, no banco de teste do próprio serviço (`go_api_test` nos
serviços Go, `finance_test` no finance).

## Preenchimento por IA

Na tela de criar nota fiscal o usuário descreve o pedido em português ("vender
3 notebooks Dell e 2 monitores LG") e os itens chegam preenchidos no
formulário. O rascunho não é gravado: quem salva é o usuário, pelo caminho de
sempre.

Precisa de uma chave da Anthropic no `.env`:

```
ANTHROPIC_API_KEY=sk-ant-...
```

Sem a chave o billing sobe igual e só esse botão fica indisponível.

## Portas

| Serviço | Porta |
|---|---|
| frontend | 4200 |
| inventory | 8000 |
| billing | 8001 |
| finance | 8002 |
| inventory_db | 5433 |
| billing_db | 5434 |
| finance_db | 5435 |
| finance vite | 5173 |
| rabbitmq | 5672 / 15672 |
