package response

type ApiError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *ApiError) Error() string {
	if err, ok := e.Data.(error); ok {
		return e.Message + ": " + err.Error()
	}
	return e.Message
}

func NewError(status int, message string, data interface{}) error {
	return &ApiError{
		Status:  status,
		Message: message,
		Data:    data,
	}
}
