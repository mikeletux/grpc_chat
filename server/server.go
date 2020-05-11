package server

/*
Server interface needed to implement for the different server structs
*/
type Server interface {
	Serve() error
}

type Config struct {
	Addr     string
	Port     int
	Protocol string
}
