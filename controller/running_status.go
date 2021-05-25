package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-message-center/service"
)

func runningStatusHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, service.RunningStatus())
}
