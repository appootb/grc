package controller

import (
	"net/http"
	"strings"

	"github.com/appootb/grc/dashboard/model"
	"github.com/gin-gonic/gin"
)

func getServiceConfig(c *gin.Context) {
	displayService := c.Param("service")
	data, err := model.NewConfig().GetKeys(displayService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, NewResponse(data))
}

func updateServiceConfig(c *gin.Context) {
	displayService := c.Param("service")
	key := strings.ReplaceAll(c.Param("key"), ".", "/")
	cfg := &model.Config{}
	if err := c.BindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequest(err.Error()))
		return
	}
	if err := cfg.UpdateKey(displayService, key); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, NewResponse(""))
}

func deleteServiceConfig(c *gin.Context) {
	displayService := c.Param("service")
	key := strings.ReplaceAll(c.Param("key"), ".", "/")
	if err := model.NewConfig().DeleteKey(displayService, key); err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusOK, NewResponse(""))
}
