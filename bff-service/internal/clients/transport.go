package clients

import (
	"bff-service/pkg/middleware"
	"net/http"
)

type HeaderPropagator struct {
	Base http.RoundTripper
}

func (h *HeaderPropagator) RoundTrip(req *http.Request) (*http.Response, error) {
	if uuid, ok := req.Context().Value(middleware.AccountUUIDKey).(string); ok {
		req.Header.Set("X-Account-Uuid", uuid)
	}

	return h.Base.RoundTrip(req)
}
