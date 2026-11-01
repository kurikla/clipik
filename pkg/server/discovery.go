package server

type Discoverer interface {
	Register(instance, service string, port int, txt []string, ifaces []string) error
	Unregister() error
}

type Noop struct{}

func (Noop) Register(_, _ string, _ int, _ []string, _[]string) error { return nil }

func (Noop) Unregister() error { return nil }
