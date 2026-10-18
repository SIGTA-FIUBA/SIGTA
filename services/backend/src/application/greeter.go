package application

type GreeterService struct{}

func (g *GreeterService) Greet(name string) string {
	return "Hello " + name + "!"
}
