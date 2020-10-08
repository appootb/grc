package controller

import (
	"net/http"

	"github.com/appootb/grc/dashboard/model"
	"github.com/gin-gonic/gin"
)

func getServiceNames(c *gin.Context) {
	data := model.NewService().GetNames()
	c.JSON(http.StatusOK, NewResponse(data))
}

func updateServices(c *gin.Context) {
	model.NewService().Sync()
	c.JSON(http.StatusOK, NewResponse(""))
}
