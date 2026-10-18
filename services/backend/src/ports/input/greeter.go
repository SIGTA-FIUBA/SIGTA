package ports

type Greeter interface {
	Greet(name string) string
}
