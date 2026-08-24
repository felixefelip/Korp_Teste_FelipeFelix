# Projeto técnico: Sistema de emissão de Notas Fiscais

Dois microsserviços em Go, um frontend Angular e RabbitMQ entre eles:

- **inventory** (`:8000`) — catálogo de produtos e razão de movimentações de estoque.
- **billing** (`:8001`) — notas fiscais, seus itens e a DANFE.
- **frontend** (`:4200`) — Angular, com proxy para as duas APIs.

Cada serviço tem seu próprio Postgres e nenhum acessa o banco do outro. O
fechamento de uma nota baixa o estoque por evento, numa saga coreografada com
compensação.

- [docs/arquitetura.md](docs/arquitetura.md) — o desenho da integração e as
  decisões por trás dele.
- [docs/detalhamento-tecnico.md](docs/detalhamento-tecnico.md) — o detalhamento
  técnico pedido no desafio: ciclos de vida do Angular, RxJS, bibliotecas,
  dependências no Go e tratamento de erros.

## Checklist do desafio

### Funcionalidades

- [x] **Cadastro de produtos** — código, descrição e saldo, em *Estoque › Produtos*.
      O saldo é informado como **estoque inicial** no cadastro; a partir daí ele é
      resultado das movimentações e não volta a ser campo editável, para que o
      razão e o saldo nunca discordem.
- [x] **Cadastro de notas fiscais** — numeração sequencial sugerida pelo sistema
      (série + número, `MAX(number) + 1` dentro da série), status inicial
      **Aberta** e inclusão de vários produtos com suas quantidades.
- [x] **Impressão de notas fiscais** — ação **Imprimir** no menu da nota:
  - [x] só aparece para notas **Abertas** — a nota fechada oferece *Ver DANFE* e
        *Reabrir*, e a API recusa um segundo fechamento com `409`;
  - [x] durante o processamento a nota exibe o status **Processando**;
  - [x] ao concluir, o status vira **Fechada**;
  - [x] o saldo dos produtos é baixado conforme as quantidades da nota.

### Requisitos obrigatórios

- [x] **Arquitetura de microsserviços** — `inventory` (produtos e saldos) e
      `billing` (notas fiscais), cada um com seu Postgres, conversando por
      RabbitMQ.
- [x] **Tratamento de falhas** — o sistema se recupera sozinho de inventory,
      RabbitMQ ou banco fora do ar, e o que não se resolve sozinho vira aviso na
      tela com a ação **Tentar novamente**. Os cenários estão em
      [Cenários de indisponibilidade](docs/arquitetura.md#cenários-de-indisponibilidade),
      todos reproduzíveis com um `docker compose stop`.
- [x] **Conexão real com banco de dados** — dois Postgres, um por serviço.

### Requisitos opcionais

- [x] **Tratamento de concorrência** — o fechamento trava os produtos com
      `SELECT ... ORDER BY id FOR UPDATE` e valida a **soma por produto** dentro
      da mesma transação que grava os movimentos. Duas notas disputando o último
      saldo são serializadas pelo lock: uma fecha, a outra volta para *Aberta*
      com o motivo.
- [x] **Uso de inteligência artificial** — o usuário descreve o pedido em
      português e os itens chegam preenchidos no formulário. Detalhes em
      [Preenchimento por IA](#preenchimento-por-ia).
- [x] **Idempotência** — mensagem repetida não duplica efeito: no inventory a
      verificação é por existência dos movimentos da nota, no billing é por
      estado (`UPDATE ... WHERE status = 'CLOSING'`).

## Como rodar

```bash
cp .env.example .env      # opcional, só para o preenchimento por IA
docker compose up -d
```

Esse comando sobe os bancos, o RabbitMQ e o frontend. Os dois serviços Go sobem
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
