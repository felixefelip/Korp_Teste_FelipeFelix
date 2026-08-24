package apierr_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"billing/internal/infra/web/apierr"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func serve(t *testing.T, handler gin.HandlerFunc) (*httptest.ResponseRecorder, string) {
	t.Helper()

	log := &bytes.Buffer{}
	previous := gin.DefaultErrorWriter
	gin.DefaultErrorWriter = log
	t.Cleanup(func() { gin.DefaultErrorWriter = previous })

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/things/:id", handler)

	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/things/7", nil))

	return response, log.String()
}

func TestInternalAnswersWithTheMessageWrittenForTheUser(t *testing.T) {
	response, _ := serve(t, func(ctx *gin.Context) {
		apierr.Internal(ctx, "erro ao buscar a coisa", errors.New("connection refused"))
	})

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"message": "erro ao buscar a coisa"}`, response.Body.String())
}

func TestInternalKeepsTheCauseOutOfTheResponse(t *testing.T) {
	response, _ := serve(t, func(ctx *gin.Context) {
		apierr.Internal(ctx, "erro ao buscar a coisa", errors.New("connection refused"))
	})

	assert.NotContains(t, response.Body.String(), "connection refused")
}

func TestInternalLogsTheCauseWithTheRoute(t *testing.T) {
	_, log := serve(t, func(ctx *gin.Context) {
		apierr.Internal(ctx, "erro ao buscar a coisa", errors.New("connection refused"))
	})

	assert.Contains(t, log, "GET /things/:id")
	assert.Contains(t, log, "connection refused")
}

func TestLogWritesWithoutTouchingTheResponse(t *testing.T) {
	response, log := serve(t, func(ctx *gin.Context) {
		apierr.Log(ctx, errors.New("upstream timed out"))
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "tente novamente"})
	})

	assert.Equal(t, http.StatusBadGateway, response.Code)
	assert.Contains(t, log, "upstream timed out")
}
