package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"inventory/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type produtoResposta struct {
	ID    int     `json:"id"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

const produtoValido = `{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":12}`

func decodeProduto(t *testing.T, body []byte) produtoResposta {
	t.Helper()

	var produto produtoResposta
	require.NoError(t, json.Unmarshal(body, &produto))

	return produto
}

func decodeErros(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var corpo struct {
		Errors map[string]string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(body, &corpo))

	return corpo.Errors
}

func TestCreateProductRetorna201ComOProdutoCriado(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", produtoValido)

	require.Equal(t, http.StatusCreated, response.Code)

	criado := decodeProduto(t, response.Body.Bytes())

	assert.NotZero(t, criado.ID, "a resposta deveria trazer o id gerado")
	assert.Equal(t, "Camiseta", criado.Name)
	assert.Equal(t, 30.99, criado.Price)
	assert.Equal(t, 12, criado.Stock)
}

func TestCreateProductSemStockRetornaZero(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", `{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Zero(t, decodeProduto(t, response.Body.Bytes()).Stock)
}

func TestCreateProductPersisteNoBanco(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", produtoValido)
	require.Equal(t, http.StatusCreated, response.Code)

	var salvos []model.Product
	require.NoError(t, testConnection.Find(&salvos).Error)

	require.Len(t, salvos, 1)
	assert.Equal(t, "Camiseta", salvos[0].Name)
	assert.Equal(t, 30.99, salvos[0].Price)
	assert.Equal(t, 12, salvos[0].Stock)
}

func TestCreateProductComJSONInvalidoRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", `{"name":`)

	assert.Equal(t, http.StatusBadRequest, response.Code)

	var corpo map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &corpo))
	assert.NotEmpty(t, corpo["message"], "JSON quebrado nao tem campo culpado, entao vira mensagem")

	var salvos []model.Product
	require.NoError(t, testConnection.Find(&salvos).Error)
	assert.Empty(t, salvos, "nada deveria ter sido gravado")
}

func TestCreateProductComPrecoDeTipoErradoRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":"muito caro"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "tipo invalido", decodeErros(t, response.Body.Bytes())["price"])
}

func TestCreateProductComStockDeTipoErradoRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":"muitos"}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "tipo invalido", decodeErros(t, response.Body.Bytes())["stock"])
}

func TestCreateProductSemOsCamposObrigatoriosRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", `{}`)

	require.Equal(t, http.StatusBadRequest, response.Code)

	erros := decodeErros(t, response.Body.Bytes())
	assert.Equal(t, "obrigatorio", erros["code"])
	assert.Equal(t, "obrigatorio", erros["name"])
	assert.Equal(t, "obrigatorio", erros["unit"])
	assert.Equal(t, "obrigatorio", erros["price"])
	assert.NotContains(t, erros, "stock", "estoque ausente e valido e vira zero")

	var salvos []model.Product
	require.NoError(t, testConnection.Find(&salvos).Error)
	assert.Empty(t, salvos, "requisicao invalida nao pode gravar nada")
}

func TestCreateProductComPrecoEEstoqueZeroRetorna201(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Amostra gratis","unit":"UN","price":0,"stock":0}`)

	require.Equal(t, http.StatusCreated, response.Code)

	criado := decodeProduto(t, response.Body.Bytes())
	assert.Zero(t, criado.Price)
	assert.Zero(t, criado.Stock)
}

func TestCreateProductComPrecoNegativoRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":-10}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErros(t, response.Body.Bytes())["price"])
}

func TestCreateProductComEstoqueNegativoRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99,"stock":-5}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErros(t, response.Body.Bytes())["stock"])
}

func TestCreateProductComUnidadeForaDaListaRetorna400(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"BANANA","price":30.99}`)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErros(t, response.Body.Bytes())["unit"])
}

func TestCreateProductComTextoAcimaDoLimiteRetorna400(t *testing.T) {
	server := newServer(t)

	codigoLongo := ""
	for range 31 {
		codigoLongo += "X"
	}

	response := post(t, server, "/products",
		fmt.Sprintf(`{"code":%q,"name":"Camiseta","unit":"UN","price":30.99}`, codigoLongo))

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotEmpty(t, decodeErros(t, response.Body.Bytes())["code"],
		"o limite do varchar(30) precisa virar 400, nao erro do banco")
}

func TestCreateProductIgnoraOIDEnviadoPeloCliente(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"id":999,"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.NotEqual(t, 999, decodeProduto(t, response.Body.Bytes()).ID)
}

func TestCreateProductNormalizaCodigoEDescricao(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products",
		`{"code":"  prd-0001 ","name":"  Camiseta  ","unit":"UN","price":30.99}`)

	require.Equal(t, http.StatusCreated, response.Code)

	criado := decodeProduto(t, response.Body.Bytes())
	assert.Equal(t, "PRD-0001", criado.Code)
	assert.Equal(t, "Camiseta", criado.Name)

	var salvos []model.Product
	require.NoError(t, testConnection.Find(&salvos).Error)
	require.Len(t, salvos, 1)
	assert.Equal(t, "PRD-0001", salvos[0].Code, "o banco guarda o valor normalizado")
}

func TestRespostaExpoeApenasOsCamposDoContrato(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", produtoValido)
	require.Equal(t, http.StatusCreated, response.Code)

	var corpo map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &corpo))

	campos := make([]string, 0, len(corpo))
	for campo := range corpo {
		campos = append(campos, campo)
	}
	sort.Strings(campos)

	assert.Equal(t, []string{"code", "id", "name", "price", "stock", "unit"}, campos)
}

func TestGetProductsRetorna200ComOsProdutos(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, post(t, server, "/products", produtoValido).Code)
	require.Equal(t, http.StatusCreated, post(t, server, "/products",
		`{"code":"PRD-0002","name":"Calca Jeans","unit":"UN","price":89.99,"stock":3}`).Code)

	response := get(t, server, "/products")

	require.Equal(t, http.StatusOK, response.Code)

	var produtos []produtoResposta
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &produtos))

	require.Len(t, produtos, 2)
	assert.Equal(t, "Camiseta", produtos[0].Name)
	assert.Equal(t, 30.99, produtos[0].Price)
	assert.Equal(t, 12, produtos[0].Stock)
	assert.Equal(t, "Calca Jeans", produtos[1].Name)
	assert.Equal(t, 89.99, produtos[1].Price)
	assert.Equal(t, 3, produtos[1].Stock)
}

func TestGetProductsQuandoNaoHaProdutos(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/products")

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[]`, response.Body.String())
}

func TestGetProductsQuandoOBancoFalhaRetorna500(t *testing.T) {
	server := newServerComBancoIndisponivel(t)

	response := get(t, server, "/products")

	require.Equal(t, http.StatusInternalServerError, response.Code)

	var corpo any
	assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &corpo),
		"o corpo precisa ser um JSON unico e valido, nao dois concatenados")
}

func TestGetProductByIDRetorna200ComOProduto(t *testing.T) {
	server := newServer(t)

	criado := post(t, server, "/products", produtoValido)
	require.Equal(t, http.StatusCreated, criado.Code)

	esperado := decodeProduto(t, criado.Body.Bytes())

	response := get(t, server, fmt.Sprintf("/products/%d", esperado.ID))

	require.Equal(t, http.StatusOK, response.Code)

	product := decodeProduto(t, response.Body.Bytes())

	assert.Equal(t, esperado.ID, product.ID)
	assert.Equal(t, "Camiseta", product.Name)
	assert.Equal(t, 30.99, product.Price)
	assert.Equal(t, 12, product.Stock)
}

func TestGetProductByIDQuandoNaoExisteRetorna404(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/products/404")

	require.Equal(t, http.StatusNotFound, response.Code)

	var corpo map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &corpo))
	assert.NotEmpty(t, corpo["message"])
}

func TestGetProductByIDComIDNaoNumericoRetorna400(t *testing.T) {
	server := newServer(t)

	response := get(t, server, "/products/abc")

	require.Equal(t, http.StatusBadRequest, response.Code)

	var corpo map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &corpo))
	assert.NotEmpty(t, corpo["message"])
}

func TestCreateProductRetornaCodigoEUnidade(t *testing.T) {
	server := newServer(t)

	response := post(t, server, "/products", produtoValido)

	require.Equal(t, http.StatusCreated, response.Code)

	criado := decodeProduto(t, response.Body.Bytes())
	assert.Equal(t, "PRD-0001", criado.Code)
	assert.Equal(t, "UN", criado.Unit)
}

func TestCreateProductPersisteCodigoEUnidade(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"CX","price":30.99}`).Code)

	var salvos []model.Product
	require.NoError(t, testConnection.Find(&salvos).Error)

	require.Len(t, salvos, 1)
	assert.Equal(t, "PRD-0001", salvos[0].Code)
	assert.Equal(t, "CX", salvos[0].Unit)
}

func TestGetProductsRetornaCodigoEUnidade(t *testing.T) {
	server := newServer(t)

	require.Equal(t, http.StatusCreated, post(t, server, "/products", produtoValido).Code)

	response := get(t, server, "/products")
	require.Equal(t, http.StatusOK, response.Code)

	var produtos []produtoResposta
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &produtos))

	require.Len(t, produtos, 1)
	assert.Equal(t, "PRD-0001", produtos[0].Code)
	assert.Equal(t, "UN", produtos[0].Unit)
}

func TestCreateProductAceitaCodigoDuplicado(t *testing.T) {
	server := newServer(t)

	primeiro := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Camiseta","unit":"UN","price":30.99}`)
	segundo := post(t, server, "/products",
		`{"code":"PRD-0001","name":"Calca Jeans","unit":"UN","price":89.99}`)

	assert.Equal(t, http.StatusCreated, primeiro.Code)
	assert.Equal(t, http.StatusCreated, segundo.Code, "codigo duplicado e permitido")

	var salvos []model.Product
	require.NoError(t, testConnection.Find(&salvos).Error)
	assert.Len(t, salvos, 2)
}
