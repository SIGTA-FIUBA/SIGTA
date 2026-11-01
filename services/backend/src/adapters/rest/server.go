package rest

type Server struct {
	*GreeterHandler
	*AuthHandler
}

func NewServer(greeter *GreeterHandler, auth *AuthHandler) *Server {
	return &Server{GreeterHandler: greeter, AuthHandler: auth}
}
