package response

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool        `json:"success"`
	Status    int         `json:"status"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestId string      `json:"request_id,omitempty"`
}

func Success(status int, data interface{}, message string) Response {
	return Response{Success: true, Status: status, Data: data, Message: message}
}

func Failure(status int, message string) Response {
	return Response{Success: false, Status: status, Message: message}
}

func SendSuccess(c *gin.Context, status int, data interface{}, message string) {
	c.JSON(status, Success(status, data, message))
}

func SendFailure(c *gin.Context, status int, message string) {
	c.JSON(status, Failure(status, message))
}

func HandleError(c *gin.Context, err error) {
	status := 500
	message := "Внутренняя ошибка сервера"
	var detailedErr interface{}

	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		status = apiErr.Status
		message = apiErr.Message
		detailedErr = apiErr.Data
	}

	log.Printf("[API-ERROR] %s %s | Status: %d | Msg: %s | Cause: %v",
		c.Request.Method,
		c.Request.URL.Path,
		status,
		err.Error(),
		detailedErr,
	)

	SendFailure(c, status, message)
}

func ToResponse(c *gin.Context, data interface{}, err error) {
	if err != nil {
		status := 500
		message := "Internal Server Error"

		var apiErr *ApiError
		if errors.As(err, &apiErr) {
			status = apiErr.Status
			message = apiErr.Message
		}

		c.JSON(status, Response{
			Success: false,
			Status:  status,
			Message: message,
		})
		return
	}

	c.JSON(200, Response{Success: true, Status: 200, Data: data})
}
