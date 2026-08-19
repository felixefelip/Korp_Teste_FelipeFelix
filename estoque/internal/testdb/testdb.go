// Package testdb concentra o setup do banco compartilhado pelos testes de
// integracao dos varios pacotes.
package testdb

import (
	"fmt"
	"os"
	"testing"

	"estoque/internal/db"
	"estoque/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Setup garante que o banco de teste existe e esta migrado, devolvendo a conexao.
// Deve ser chamado a partir do TestMain de cada pacote.
func Setup() (*gorm.DB, error) {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "test")
	}

	// Trava de seguranca: os testes truncam tabelas, entao so podem rodar
	// contra o banco de teste.
	if db.Env() != "test" {
		return nil, fmt.Errorf("recusando rodar com GO_ENV=%q; use GO_ENV=test", db.Env())
	}

	if err := db.EnsureDatabase(); err != nil {
		return nil, fmt.Errorf("criando o banco: %w", err)
	}

	connection, err := db.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("conectando no banco: %w", err)
	}

	if err := connection.AutoMigrate(&model.Product{}); err != nil {
		return nil, fmt.Errorf("migrando o banco: %w", err)
	}

	return connection, nil
}

// Reset devolve o banco ao estado vazio, para que um teste nao enxergue o
// estado deixado por outro.
func Reset(t testing.TB, connection *gorm.DB) {
	t.Helper()

	err := connection.Exec("TRUNCATE TABLE product RESTART IDENTITY CASCADE").Error
	require.NoError(t, err, "limpando a tabela product")
}
