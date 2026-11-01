package server

type Discoverer interface {
	Register(instance, service string, port int, txt []string, ifaces []string) error
	Unregister() error
}
