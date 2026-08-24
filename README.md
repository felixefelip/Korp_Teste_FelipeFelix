# Korp — notas fiscais e estoque

Dois microsserviços em Go, um frontend Angular e RabbitMQ entre eles:

- **inventory** (`:8000`) — catálogo de produtos e razão de movimentações de estoque.
- **billing** (`:8001`) — notas fiscais, seus itens e a DANFE.
- **frontend** (`:4200`) — Angular, com proxy para as duas APIs.

Cada serviço tem seu próprio Postgres e nenhum acessa o banco do outro. O
fechamento de uma nota baixa o estoque por evento, numa saga coreografada com
compensação. O desenho e as decisões estão em
[docs/arquitetura.md](docs/arquitetura.md).

## Como rodar

```bash
cp .env.example .env      # opcional, só para o preenchimento por IA
make up
```

`make up` sobe os bancos, o RabbitMQ e o frontend. Os dois serviços Go sobem
como container ocioso — o processo é iniciado à mão, cada um no seu terminal:

```bash
make inventory-run
make billing-run
```

A aplicação fica em http://localhost:4200 e o painel do RabbitMQ em
http://localhost:15672 (`guest` / `guest`).

Os bancos são criados e migrados na subida de cada serviço, então não há passo
de migração separado.

## Testes

```bash
make inventory-test
make billing-test
make front-test
```

Os testes de Go rodam contra um Postgres de verdade, no banco de teste do
próprio serviço — nada é mockado na camada de repositório.

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
| inventory_db | 5433 |
| billing_db | 5434 |
| rabbitmq | 5672 / 15672 |
