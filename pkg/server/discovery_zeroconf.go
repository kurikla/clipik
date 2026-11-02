package server

import (
	"fmt"
	"net"

	"github.com/grandcat/zeroconf"
)

type ZeroconfDiscoverer struct {
	srv *zeroconf.Server
}

func (d *ZeroconfDiscoverer) Register(instance, service string, port int, txt []string, ifaces []string) error {
	nets :=  make([]net.Interface, 0, len(ifaces))

	for _, name := range ifaces {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			return fmt.Errorf("interface %q: %w", name, err)
		}
		nets = append(nets, *iface)
	}

	srv, err := zeroconf.Register(instance, service, "local.", port, txt, nets)

	if err != nil {
		return fmt.Errorf("zeroconf register: %w", err)
	}

	d.srv = srv

	return nil
}

func (d *ZeroconfDiscoverer) Unregister() error {
	if d.srv == nil {
		return nil
	}

	d.srv.Shutdown()
	d.srv = nil

	return nil
}
