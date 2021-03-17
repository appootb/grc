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

func deleteService(c *gin.Context) {
	displayService := c.Param("service")
	if err := model.NewService().Delete(displayService); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, NewResponse(""))
}
