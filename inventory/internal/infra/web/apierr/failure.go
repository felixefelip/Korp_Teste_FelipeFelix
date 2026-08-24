package apierr

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Log(ctx *gin.Context, err error) {
	fmt.Fprintf(gin.DefaultErrorWriter, "%s %s: %v\n", ctx.Request.Method, ctx.FullPath(), err)
}

func Internal(ctx *gin.Context, message string, err error) {
	Log(ctx, err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": message})
}
