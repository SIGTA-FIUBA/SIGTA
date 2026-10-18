package rest

type Server struct {
	*GreeterHandler
}

func NewServer(greeter *GreeterHandler) *Server {
	return &Server{GreeterHandler: greeter}
}
