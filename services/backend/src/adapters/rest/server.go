package rest

type Server struct {
	*AuthHandler
}

func NewServer(auth *AuthHandler) *Server {
	return &Server{AuthHandler: auth}
}
