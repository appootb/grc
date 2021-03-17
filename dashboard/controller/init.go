package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init(router *gin.RouterGroup) error {
	cfg := router.Group("/config")
	{
		cfg.GET("/:service", getServiceConfig)
		cfg.PUT("/:service/:key", updateServiceConfig)
		cfg.DELETE("/:service/:key", deleteServiceConfig)
	}
	svc := router.Group("/service")
	{
		svc.GET("/", getServiceNames)
		svc.PUT("/", updateServices)
		svc.DELETE("/:service", deleteService)
	}

	return nil
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"error"`
	Data    interface{} `json:"data"`
}

func (resp *Response) String() string {
	v, _ := json.Marshal(resp)
	return string(v)
}

func NewResponse(data interface{}) *Response {
	return &Response{
		Data: data,
	}
}

func NewError(err error) *Response {
	return &Response{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
	}
}

func NewBadRequest(msg string) *Response {
	return &Response{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}
