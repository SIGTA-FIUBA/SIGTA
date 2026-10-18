package rest

import (
	ports "backend/src/ports/input"
	"context"
)

type GreeterHandler struct {
	greeter ports.Greeter
}

func (h *GreeterHandler) GetHello(ctx context.Context, req GetHelloRequestObject) (GetHelloResponseObject, error) {
	return GetHello200JSONResponse{Message: h.greeter.Greet("world")}, nil
}

func NewGreeterHandler(greeter ports.Greeter) *GreeterHandler {
	return &GreeterHandler{greeter: greeter}
}
