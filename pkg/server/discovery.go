package server

type Discoverer interface {
	Register(instance, service string, port int, txt []string, ifaces []string) error
	Unregister() error
}

func (s *Server) WithDiscoverer(d Discoverer) *Server {
	if d != nil {
		s.discoverer = d
	}

	return s
}
