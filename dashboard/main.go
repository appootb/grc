package main

import (
	"github.com/appootb/grc/dashboard/config"
	"github.com/appootb/grc/dashboard/controller"
	"github.com/appootb/grc/dashboard/model"
	"github.com/gin-gonic/gin"
)

func main() {
	// Init config
	if err := config.Init(); err != nil {
		panic("init config failed, err: " + err.Error())
	}

	// Init model
	if err := model.Init(); err != nil {
		panic("init model failed, err: " + err.Error())
	}

	// Init controller
	router := gin.Default()
	if err := controller.Init(router.Group("/api")); err != nil {
		panic("register router controller failed, err: " + err.Error())
	}

	// Serve
	if err := router.Run(config.GlobalConfig.Dashboard.ServeAddress); err != nil {
		panic("gin serve failed, err: " + err.Error())
	}
}
