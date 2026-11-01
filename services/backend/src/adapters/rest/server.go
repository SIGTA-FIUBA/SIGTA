package rest

type Server struct {
	*GreeterHandler
	*AuthHandler
	*DocumentHandler
}

func NewServer(greeter *GreeterHandler, auth *AuthHandler, documents *DocumentHandler) *Server {
	return &Server{GreeterHandler: greeter, AuthHandler: auth, DocumentHandler: documents}
}
