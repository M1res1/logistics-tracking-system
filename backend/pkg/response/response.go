package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, envelope{Data: data})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, envelope{Error: msg})
}
