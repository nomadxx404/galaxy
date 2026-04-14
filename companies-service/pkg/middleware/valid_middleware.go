package middleware

import (
	"companies-service/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func BindJSON[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		//TODO
		//Kafka - send logging
		response.SendFailure(c, 400, "Некорректный формат данных")
		return nil, false
	}
	return &req, true
}

func GetParamInt32(c *gin.Context, name string) (int32, error) {
	param := c.Param(name)
	if param == "" {
		return 0, fmt.Errorf("Параметр %s обязателенs", name)
	}

	val, err := strconv.ParseInt(param, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Неверный параметр: %s", name)
	}
	//TODO
	//Kafka - send logging
	return int32(val), nil
}
