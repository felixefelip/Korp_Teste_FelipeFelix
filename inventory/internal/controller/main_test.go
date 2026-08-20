package controller_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"inventory/internal/db"
	"inventory/internal/router"
	"inventory/internal/testdb"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testConnection e a conexao compartilhada por todos os testes do pacote.
var testConnection *gorm.DB

func TestMain(m *testing.M) {
	connection, err := testdb.Setup()
	if err != nil {
		fmt.Fprintln(os.Stderr, "setup do banco de teste:", err)
		os.Exit(1)
	}

	testConnection = connection
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

// newServer monta o servidor com as rotas reais da aplicacao (as mesmas que o
// main.go registra, via router.Register) sobre uma tabela product vazia.
func newServer(t *testing.T) *gin.Engine {
	t.Helper()

	testdb.Reset(t, testConnection)

	server := gin.New()
	router.Register(server, testConnection)

	return server
}

// newServerComBancoIndisponivel monta o mesmo servidor sobre uma conexao propria
// que ja foi fechada, para exercitar o caminho de erro dos handlers. A conexao
// compartilhada pelos demais testes nao e afetada.
func newServerComBancoIndisponivel(t *testing.T) *gin.Engine {
	t.Helper()

	connection, err := db.ConnectDB()
	require.NoError(t, err)

	sqlDB, err := connection.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	server := gin.New()
	router.Register(server, connection)

	return server
}

// do dispara uma requisicao contra o servidor e devolve a resposta gravada.
func do(t *testing.T, server *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	return response
}

func get(t *testing.T, server *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	return do(t, server, http.MethodGet, path, "")
}

func post(t *testing.T, server *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	return do(t, server, http.MethodPost, path, body)
}
