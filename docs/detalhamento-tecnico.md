# Detalhamento técnico

Este documento tem o checklist do desafio e as respostas às perguntas de
detalhamento. O desenho da integração entre os dois serviços — saga, outbox,
filas, idempotência — está em [arquitetura.md](arquitetura.md).

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
      [Cenários de indisponibilidade](arquitetura.md#cenários-de-indisponibilidade),
      todos reproduzíveis com um `docker compose stop`.
- [x] **Conexão real com banco de dados** — dois Postgres, um por serviço.

### Requisitos opcionais

- [x] **Tratamento de concorrência** — o fechamento trava os produtos com
      `SELECT ... ORDER BY id FOR UPDATE` e valida a **soma por produto** dentro
      da mesma transação que grava os movimentos. Duas notas disputando o último
      saldo são serializadas pelo lock: uma fecha, a outra volta para *Aberta*
      com o motivo.
  - [`ApplyInvoice`](../inventory/internal/infra/db/stock_movement_repository.go#L78-L110)
    — a transação que trava, decide e grava
  - [`lockProducts`](../inventory/internal/infra/db/stock_movement_repository.go#L201-L224)
    — `FOR UPDATE` com `ORDER BY id`, para duas notas com produtos em comum
    pegarem os locks na mesma sequência
  - [`QuantityRequiredByProduct`](../inventory/internal/model/invoice_stock.go#L28-L36)
    — agrega por produto antes de comparar, senão dois itens do mesmo produto
    passam separadamente contra o mesmo saldo
- [x] **Uso de inteligência artificial** — o usuário descreve o pedido em
      português e os itens chegam preenchidos no formulário. Detalhes em
      [Preenchimento por IA](../README.md#preenchimento-por-ia).
- [x] **Idempotência** — mensagem repetida não duplica efeito: no inventory a
      verificação é por existência dos movimentos da nota, no billing é por
      estado (`UPDATE ... WHERE status = 'CLOSING'`).
  - [`alreadyApplied`](../inventory/internal/infra/db/stock_movement_repository.go#L226-L235)
    — a nota já baixada é reconhecida sob o mesmo lock
  - [`ResolveInvoiceStock`](../inventory/internal/model/invoice_stock_decision.go#L26-L44)
    — a checagem de replay vem **antes** da de saldo: na reentrega o estoque já
    foi baixado, e comparar de novo recusaria uma nota que deu certo
  - [`StockMovement.BillingInvoiceItemID`](../inventory/internal/model/stock_movement.go#L26)
    — índice único, a garantia no banco
  - [`moveFrom`](../billing/internal/infra/db/invoice_repository.go#L239-L270)
    — a transição só se aplica a partir do estado esperado; `RowsAffected == 0`
    significa "alguém já moveu", que é confirmado e ignorado

## Ciclos de vida do Angular

**Nenhum hook clássico foi utilizado.** Não há uma única ocorrência de
`ngOnInit`, `ngOnDestroy`, `ngOnChanges`, `ngAfterViewInit`,
`ngAfterContentInit` ou `ngDoCheck` em `frontend/src` — nem nos componentes,
nem nos testes.

O projeto usa a API de signals do Angular, em que cada hook tem um substituto
direto:

| Hook clássico | O que foi usado no lugar | Onde |
|---|---|---|
| `ngOnInit` | o próprio construtor | `invoice-list.ts:66`, `invoice-edit.ts:40`, `invoice-new`, `product-edit` |
| `ngOnChanges` | `effect()` lendo `input()` | `invoice-form.ts:90/101/119/133`, `product-form.ts:62`, `movement-form.ts:58` |
| `ngOnDestroy` | `onCleanup` dentro do `effect()` | `invoice-list.ts:69/79`, `menu-button.ts:58` |
| `ngAfterViewInit` | `effect()` sobre `viewChildren()` | `menu-button.ts:82` |

**Por que o construtor substitui o `ngOnInit`.** O `ngOnInit` existe porque um
`@Input()` decorado só recebe valor depois da construção do componente. Com
`input()` como signal o valor é lido reativamente, quando for lido — então o
construtor volta a ser o lugar certo para disparar a carga inicial. O `id` da
rota também é lido antes disso, no field initializer:

```ts
private readonly invoiceId = Number(this.route.snapshot.paramMap.get('id'));

constructor() {
  this.load();
  this.loadProducts();
}
```

**Por que o `onCleanup` substitui o `ngOnDestroy`.** Os dois usos são de
recursos que precisam ser desfeitos: os `setInterval` do polling da listagem de
notas e os listeners de `click`/`scroll`/`resize` do menu de ações. O
`onCleanup` roda tanto quando o efeito reexecuta quanto quando o componente é
destruído, e fica escrito no mesmo bloco que registrou o recurso:

```ts
effect((onCleanup) => {
  if (!this.processing()) {
    return;
  }

  const timer = setInterval(() => this.refresh(), this.pollInterval());

  onCleanup(() => clearInterval(timer));
});
```

Com `ngOnDestroy` o handle do timer teria que virar campo da classe, e a
limpeza ficaria em outro lugar do arquivo — separada da linha que a exige, e
sem cobrir a reexecução (o intervalo muda de 1,5s para 10s conforme o estágio
das notas).

O que é só derivação de estado não vira efeito nenhum: usa `computed()`, como o
filtro da listagem e o `pollInterval`.

## RxJS

Sim, mas restrito à camada de serviço. **No código de aplicação são dois nomes**:
`Observable` e `tap`, importados nos quatro serviços (`product.service.ts`,
`movement.service.ts`, `invoice.service.ts`, `catalog.service.ts`). Nenhum outro
operador.

`Observable` aparece como tipo de retorno, porque é o que o `HttpClient`
devolve. `tap` faz uma coisa só — espelhar a resposta HTTP num signal que o
serviço expõe:

```ts
list(): Observable<Product[]> {
  return this.http
    .get<Product[]>(RESOURCE)
    .pipe(tap((products) => this._products.set(products)));
}
```

Essa é a costura entre os dois modelos: da borda para dentro é RxJS, porque é
assim que o Angular entrega o HTTP; de dentro para a tela é signal
(`readonly products = this._products.asReadonly()`), que é o que os `computed()`
dos componentes consomem. O mesmo `tap` mantém a lista em memória atualizada no
`create`, `update`, `remove` e nas transições de nota, evitando um GET novo a
cada ação.

**Nos componentes não há RxJS.** Todos consomem com `.subscribe({ next, error })`
e guardam o resultado em signals; **não existe `| async` em nenhum template**.
Nada é desinscrito, e não precisa ser: o `HttpClient` completa a Observable
depois da resposta.

**Nos testes o uso é maior**, em 10 arquivos: `of` para resposta imediata,
`throwError` para o caminho de erro e `Subject` quando o teste precisa segurar a
resposta pendurada, para verificar o estado "Salvando…" antes de completar.

Vale registrar o que **não** foi usado, já que seria o caminho tradicional: o
polling da listagem é `setInterval` dentro de um `effect`, não
`interval().pipe(switchMap(...))`; o filtro de busca é `computed()` sobre um
signal, não `valueChanges.pipe(debounceTime())`. Com uma requisição por ação e o
resultado indo para signal, o resto da biblioteca não teria trabalho a fazer.

## Outras bibliotecas utilizadas

### Frontend

| Biblioteca | Finalidade |
|---|---|
| `@angular/core`, `common`, `forms`, `router`, `platform-browser`, `compiler` | o próprio framework — signals, formulários reativos, roteamento |
| `rxjs` | o que o `HttpClient` devolve; usado como descrito acima |
| `tslib` | runtime do TypeScript |
| `@angular/build`, `@angular/cli`, `@angular/compiler-cli` | build e dev server |
| `typescript` | linguagem |
| `vitest` + `jsdom` | execução dos testes (551 testes em 26 arquivos) |
| `prettier` | formatação |

São todas as dependências do `package.json` — não há nenhuma outra.

### Backend

| Biblioteca | Serviço | Finalidade |
|---|---|---|
| `gin-gonic/gin` | ambos | roteamento HTTP, binding e validação de JSON |
| `gorm.io/gorm` + `gorm.io/driver/postgres` | ambos | acesso ao Postgres |
| `jackc/pgx/v5` | billing (direto) | ler o `SQLSTATE` do erro — `23505` vira `ErrInvoiceDuplicated` |
| `go-playground/validator/v10` | ambos | as regras das tags `binding:"..."`, usadas diretamente para traduzir as mensagens |
| `rabbitmq/amqp091-go` | ambos | publicação e consumo das filas |
| `google/uuid` | ambos | `eventId` de cada evento do outbox |
| `go-pdf/fpdf` | billing | geração da DANFE |
| `anthropics/anthropic-sdk-go` | billing | extração do rascunho de nota a partir do prompt |
| `stretchr/testify` | ambos | asserções dos testes |

No inventory o `pgx` é dependência indireta: ele chega pelo driver do GORM, mas
o código não o importa.

## Bibliotecas de componentes visuais

**Nenhuma.** Não há Angular Material, PrimeNG, Bootstrap, Tailwind nem
biblioteca de ícones — as dependências de produção do `package.json` são só o
Angular, o RxJS e o `tslib`.

O que faz esse papel é a pasta `src/app/shared/`, com quatro componentes
escritos à mão e reutilizados pelas telas:

- **`menu-button`** — o menu de ações das tabelas, com `position: fixed` para
  escapar do `overflow` da tabela, navegação por teclado e fechamento em clique
  fora, scroll e resize
- **`confirm-dialog`** — as confirmações de imprimir e excluir, com estado
  ocupado durante a requisição
- **`flash`** — mensagens de sucesso e erro que sobrevivem à navegação
- **`navbar`**

O visual vem de SCSS próprio em duas camadas: `src/styles.scss` define os tokens
em variáveis CSS (paleta, raios, sombras, tipografia) e importa seis parciais em
`src/styles/` (`_button`, `_card`, `_field`, `_form`, `_page`, `_table`) com as
classes utilitárias usadas nos templates; o resto é estilo por componente, no
`.scss` ao lado de cada um. Os únicos ícones são dois SVG inline no
`menu-button`.

## Gerenciamento de dependências no Go

**Go Modules**, um módulo por serviço — `module billing` e `module inventory`,
cada um com seu `go.mod` e `go.sum`. Não há `go.work` nem módulo na raiz: os
dois serviços são independentes, e o único contrato entre eles é o payload dos
eventos, nunca um pacote Go compartilhado. Compartilhar código aqui criaria
acoplamento de build entre serviços que se comunicam justamente para não estarem
acoplados.

- **`go.sum` versionado**, com o hash de cada módulo — build reprodutível e
  verificação de integridade no download.
- **Sem `vendor/`.** O cache de módulos é um volume Docker (`gomodcache`)
  compartilhado pelos dois serviços, então o download acontece uma vez e não
  polui o repositório.
- **`go mod tidy`** mantém a separação entre dependência direta e `// indirect`
  fiel ao que o código realmente importa.

## Frameworks utilizados no Go

Dois, e ambos ficam confinados na camada de infraestrutura:

- **Gin** — servidor HTTP. Aparece só em `internal/infra/web/`: roteamento,
  binding de JSON e as tags de validação.
- **GORM** — ORM sobre o Postgres. Aparece só em `internal/infra/db/`, e o
  `AutoMigrate` no boot dispensa passo de migração separado.

O núcleo não depende de nenhum dos dois. `internal/model` não importa `gin` nem
`gorm`: a decisão de negócio é função pura, como

```go
func ResolveInvoiceStock(
    request InvoiceStockRequest,
    products map[int]Product,
    alreadyApplied bool,
) (InvoiceStockDecision, error)
```

e o repositório é uma interface declarada no próprio `model`, implementada em
`infra/db`. O ganho prático está nos testes: precedência de recusa, agregação
por produto e comparação de saldo são testadas sem banco e sem HTTP, em
milissegundos.

Não há container de injeção de dependência nem geração de código: a montagem é
explícita, no `router.go` e no `main.go`.

## Outbox pattern

**Os dois serviços usam outbox transacional.** O motivo e o desenho — por que não
publicar direto do handler, o formato da tabela, o que acontece com o broker fora
do ar — estão em [Outbox transacional](arquitetura.md#outbox-transacional). Aqui
está como o padrão aparece no código.

### Onde cada peça mora

| Peça | Arquivo | Papel |
|---|---|---|
| `OutboxEvent` e `OutboxRepository` | `internal/model/outbox.go` | a entidade e a interface do repositório |
| construção do evento | `internal/model/invoice_event.go`, `product_event.go` | funções puras que devolvem um `OutboxEvent` pronto, com `eventId` e payload já serializado |
| persistência | `internal/infra/db/outbox_repository.go` | `ClaimEvents`, `MarkPublished`, `MarkFailed` |
| publicação | `internal/infra/messaging/relay.go` | a goroutine que drena a tabela |
| montagem | `internal/infra/messaging/register.go:37` | `NewRelay(outboxRepository, BillingExchange).Start()` |

A divisão é a mesma nos dois serviços e segue a regra do resto do backend: o
núcleo declara, a infra implementa. O relay depende de `model.OutboxRepository`
(`relay.go:30`), não de GORM — nada em `internal/model/outbox.go` sabe que existe
Postgres ou RabbitMQ.

### A escrita: o evento entra no mesmo `tx` do dado de negócio

É a única parte que sustenta a garantia, e por isso **nenhum evento é gravado
fora de uma `Transaction`**: as oito gravações de evento em código de produção —
três no billing, cinco no inventory — estão todas dentro de uma.

```go
func (ir *InvoiceRepository) CloseInvoice(id int, event model.OutboxEvent) error {
	return ir.connection.Transaction(func(tx *gorm.DB) error {
		if err := startTransition(tx, id, model.InvoiceStatusClosing); err != nil {
			return err
		}

		return tx.Create(&event).Error
	})
}
```

Ou a nota entra em `CLOSING` e o evento fica gravado, ou nada acontece.

Quem monta o evento é sempre uma função do `model`; o que muda é de onde ela é
chamada, e a razão é diferente em cada caso:

- **billing** — o usecase chama `model.NewInvoiceCloseRequested(invoice)` e passa
  o evento pronto ao repositório (`invoice_usecase.go:77`), que só grava.
- **inventory, fluxo de nota** — o evento é *resultado da decisão de negócio*:
  `ResolveInvoiceStock` devolve um `InvoiceStockDecision{Movements, Event}`, e o
  repositório grava os movimentos e o evento na mesma transação
  (`stock_movement_repository.go:99-103`). Aplicar o estoque e anunciar que ele
  foi aplicado deixam de poder divergir.
- **inventory, catálogo** — a função roda dentro do próprio `tx`
  (`product_repository.go:41`), porque `NewProductCreated` precisa do ID que o
  insert acabou de gerar.

### O relay

Os parâmetros ficam num bloco de constantes no topo do arquivo, sem configuração
externa:

```go
const (
	relayInterval  = time.Second
	relayBatch     = 50
	relayLease     = 30 * time.Second
	publishTimeout = 5 * time.Second

	backoffBase = 2 * time.Second
	backoffCap  = time.Minute
)
```

O `SKIP LOCKED` que permite várias réplicas sem lock distribuído é uma cláusula
do GORM, não SQL escrito à mão:

```go
tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
```

`ClaimEvents` abre a transação só para selecionar e arrendar; a publicação
acontece depois, fora dela. O `MarkPublished` só roda depois do ack do broker,
com `PublishWithDeferredConfirmWithContext` e `mandatory = true`.

São **três falhas distintas tratadas como o mesmo desfecho** — "não publicou":
erro de rede, `nack` do broker (`errBrokerRefused`) e mensagem devolvida por falta
de binding (`errUnroutable`, lida do canal de returns logo depois do confirm — sem
isso, uma routing key sem fila ligada seria contada como sucesso). Qualquer uma
cai no `fail()`, que grava a causa em `last_error` e adia com
`backoff(event.Attempts)`, dobrando de 2s até o teto de 1 minuto.

Uma decisão do `drain`: a primeira falha encerra o lote e derruba a conexão, em
vez de insistir com os 49 eventos restantes contra um broker que acabou de
recusar. Os outros seguem arrendados e voltam no tique seguinte, na mesma ordem
de `id`.

### O código é duplicado entre os serviços, de propósito

`relay.go` e `outbox_repository.go` são cópias: um `diff` entre os dois serviços
acusa o path do import e mais uma linha — `CorrelationId: event.CausationID`
(`inventory/internal/infra/messaging/relay.go:121`). O campo `CausationID` existe
só no outbox do inventory, para amarrar cada resultado ao pedido que o originou.

Extrair isso para um pacote comum criaria acoplamento de build entre dois serviços
que se falam por evento justamente para não estarem acoplados — a mesma razão de
não haver `go.work` nem módulo na raiz.

### O que está testado

`outbox_repository_test.go`, seis casos em cada serviço, contra Postgres real:
`ClaimEvents` devolve o pendente na ordem de `id`, arrenda o que entrega (uma
segunda chamada não vê o mesmo evento), respeita o limite; `MarkPublished` tira o
evento da fila; `MarkFailed` incrementa `attempts`, grava a causa e segura o
evento até o backoff vencer; id inexistente devolve erro em vez de sucesso
silencioso.

O relay em si não tem teste unitário — o que ele adiciona sobre o repositório é
I/O com o broker. Esse caminho é verificado de ponta a ponta no cenário
[B. RabbitMQ fora do ar](arquitetura.md#b-rabbitmq-fora-do-ar): a API responde
`202`, o evento fica na tabela com `published_at IS NULL` e o fluxo se completa
sozinho quando o broker volta.

## Tratamento de erros e exceções no backend

A regra que organiza tudo: **erro de negócio é valor tipado do `model`; erro de
infraestrutura é `error` de Go e sobe até quem sabe traduzi-lo.**

### 1. Erros de negócio são sentinelas declaradas no modelo

```go
// billing/internal/model/invoice.go
ErrInvoiceDuplicated    = errors.New("invoice number already used in this series")
ErrInvoiceClosed        = errors.New("closed invoice")
ErrInvoiceOpen          = errors.New("open invoice")
ErrInvoiceProcessing    = errors.New("invoice being processed")
ErrInvoiceNotProcessing = errors.New("invoice is not being processed")
```

Mais `ErrDraftUnavailable` (IA não configurada) e, no inventory,
`ErrMovementFromInvoice`. O modelo não conhece HTTP nem AMQP — devolve o
sentinel, e quem está na borda decide o que fazer com ele.

### 2. O controller traduz sentinel em status HTTP

Cada handler encadeia `errors.Is`, do caso específico para o genérico, e o `500`
é só o que sobra:

```go
if errors.Is(err, model.ErrInvoiceProcessing) { 409 "Esta nota fiscal está em processamento." }
if errors.Is(err, model.ErrInvoiceClosed)     { 409 "Notas fiscais fechadas não podem ser alteradas." }
if errors.Is(err, gorm.ErrRecordNotFound)     { 404 }
500
```

A mensagem é escrita para o usuário final e muda conforme a operação: o mesmo
`ErrInvoiceClosed` vira "não podem ser alteradas" no `UpdateInvoice` e "não podem
ser excluídas" no `DeleteInvoice`.

### 3. Erro de driver vira erro de domínio na borda do banco

O código `23505` do Postgres é detalhe de driver e não pode vazar para cima. É
convertido na saída do repositório:

```go
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    return model.ErrInvoiceDuplicated
}
```

O usecase e o controller nunca veem `pgconn.PgError`.

### 4. Erro de validação tem canal próprio

O pacote `infra/web/apierr` converte as `validator.ValidationErrors` do Gin num
mapa `campo → mensagem` e devolve `400 {"errors": {...}}`; corpo malformado
devolve `400 {"message": ...}`. São dois formatos de propósito: o primeiro o
frontend distribui embaixo de cada campo, o segundo vira banner.

Um detalhe que o `init()` do pacote resolve: sem registrar um `TagNameFunc`, o
validator devolveria a chave `Quantity` em vez de `quantity` e o frontend não
acharia o campo.

### 5. Toda falha desconhecida é registrada antes de virar 500

`apierr.Internal` é o único caminho para o `500` nos dois serviços — ele loga a
causa e responde com a mensagem escrita para o usuário:

```go
func Log(ctx *gin.Context, err error) {
	fmt.Fprintf(gin.DefaultErrorWriter, "%s %s: %v\n", ctx.Request.Method, ctx.FullPath(), err)
}

func Internal(ctx *gin.Context, message string, err error) {
	Log(ctx, err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": message})
}
```

O prefixo sai do próprio request (`GET /invoices/:id: ...`), então não existe
rótulo escrito à mão para divergir da rota. Emparelhar log e resposta na mesma
função é o que impede o modo de falha que motivou isso: uma tela dizendo
"não foi possível" enquanto a causa real não aparece em lugar nenhum.

A causa **nunca** entra no corpo da resposta — o cliente recebe a mensagem em
português, e o operador lê o resto no log.

### 6. No consumidor de mensagens, o erro tem três destinos

É a parte com mais decisão, porque nem todo `error` significa a mesma coisa:

- **Recusa de negócio não é erro.** Estoque insuficiente devolve `nil` e a
  mensagem recebe `Ack` normal — o resultado dela é publicar
  `invoice.stock.rejected`. Tratar como erro faria a mensagem esgotar as
  tentativas justamente quando o sistema funcionou como deveria.
- **Mensagem venenosa vai direto para a fila morta.** O `json.Unmarshal` que
  falha é embrulhado em `msgerr.Poison(err)` — um tipo com `Unwrap()`,
  reconhecido por `errors.As`. Corpo que não decodifica nunca vai decodificar,
  então `Nack(requeue=false)` na hora, sem espera.
- **Falha transitória retenta.** Banco fora do ar, deadlock, timeout: espera com
  backoff exponencial (2s, teto de 30s) e `Nack(requeue=true)`, até dez
  tentativas; esgotado o orçamento, fila morta. A contagem é do consumidor, pelo
  `x-acquired-count` da entrega, e o motivo está em
  [Nova tentativa e fila morta](arquitetura.md#nova-tentativa-e-fila-morta).

Todo caminho loga com o `MessageId` antes de decidir.

### 7. Falha de publicação fica gravada na tabela

O relay não perde a causa: `MarkFailed(event.ID, cause.Error(), ...)` grava a
mensagem em `outbox_event.last_error` e agenda a próxima tentativa. Uma mensagem
que não roteou — binding faltando, devolvida pelo `mandatory` — deixa rastro em
vez de sumir.

### 8. `panic` só no boot

Seis ocorrências, todas em `cmd/main.go`: falha ao criar o banco, ao conectar ou
ao migrar. São condições em que subir o serviço não faz sentido. **Não há
`panic` no caminho de request ou de mensagem, e nenhum `recover`** — um panic ali
seria bug, não fluxo previsto.

## C# e LINQ

Não se aplica: o backend é inteiramente em Go.
