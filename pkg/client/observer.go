package client

import (
	"fmt"
	"github.com/kurikla/clipik/pkg/model"
)

type Observer struct {
	client *Client
}

func (o *Observer) OnFound(service model.Service) {
	o.client.mu.Lock()
	o.client.discovered[serviceKey(service)] = service
	o.client.mu.Unlock()

	o.client.logger.Info("service found",
		"name", service.Name,
		"addr", fmt.Sprintf("%s:%d", service.Ip, service.Port))
}

func (o *Observer) OnAccept(service model.Service) {
	addr := fmt.Sprintf("%s:%d", service.Ip, service.Port)

	go func () {
		if err := o.client.connect(addr); err != nil {
			o.client.logger.Warn("connect", "addr", addr, "err", err)
		}
	}()
}

func (o *Observer) OnLost(service model.Service) {
	key := serviceKey(service)

	o.client.mu.Lock()
	delete(o.client.discovered, key)
	sc, connected := o.client.connections[fmt.Sprintf("%s:%d", service.Ip, service.Port)]
	o.client.mu.Unlock()

	o.client.logger.Info("service lost",
		"name", service.Name,
		"addr", fmt.Sprintf("%s:%d", service.Ip, service.Port))

	if connected {
		o.client.disconnect(sc.addr)
	}
}

func serviceKey(s model.Service) string {
	return fmt.Sprintf("%s:%d", s.Ip, s.Port)
}
