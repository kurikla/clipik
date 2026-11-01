package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/grandcat/zeroconf"
	"github.com/kurikla/clipik/pkg/model"
)

type ZeroconfDiscoverer struct {
	filter Filter
	logger *slog.Logger

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewZeroconfDiscoverer(filter Filter, logger *slog.Logger) *ZeroconfDiscoverer {
	return &ZeroconfDiscoverer{
		filter: filter,
		logger: logger,
	}
}

func (d *ZeroconfDiscoverer) Start(observer *Observer) error {
	d.mu.Lock()
	if d.cancel != nil {
		d.mu.Unlock()
		return errors.New("zeroconf: already started")
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.done = make(chan struct{})
	d.mu.Unlock()

	go d.browse(ctx, observer)
	return nil
}

func (d *ZeroconfDiscoverer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cancel == nil {
		return nil
	}

	d.cancel()
	<-d.done
	d.cancel = nil
	d.done = nil
	return nil
}

func (d *ZeroconfDiscoverer) browse(ctx context.Context, observer *Observer) {
	defer close(d.done)

	resolver, err := zeroconf.NewResolver(nil)

	if err != nil {
		d.logger.Warn("zeroconf: resolver", "err", err)
		return
	}

	entries := make(chan *zeroconf.ServiceEntry, 10)
	if err := resolver.Browse(ctx, "_clipik._tcp", "local.", entries); err != nil {
		d.logger.Warn("zeroconf: browse", "err", err)
		return
	}

	seen := make(map[string]struct{})
	for {
		select {
		case <-ctx.Done():
			d.logger.Info("zeroconf: browse stopped")
			return
		case e, ok := <-entries:
			if !ok {
				return
			}
			d.handleEntry(e, observer, seen)
		}
	}
}

func (d *ZeroconfDiscoverer) handleEntry(e *zeroconf.ServiceEntry, observer *Observer, seen map[string]struct{}) {
	var ip string
	if len(e.AddrIPv4) > 0 {
		ip = e.AddrIPv4[0].String()
	} else if len(e.AddrIPv6) > 0  {
		ip = e.AddrIPv6[0].String()
	}

	if ip == "" {
		return
	}

	svc := model.Service{
		Ip: ip,
		Port: e.Port,
		Name: e.Instance,
	}

	addr := fmt.Sprintf("%s:%d", svc.Ip, svc.Port)

	if _, ok := seen[addr]; ok {
		return
	}

	seen[addr] = struct{}{}

	d.logger.Info("zeroconf: service found",
		"name", svc.Name,
		"addr", addr)

	observer.OnFound(svc)

	if d.filter != nil && d.filter.Accept(svc) {
		d.logger.Info("zeroconf: service accepted, connecting",
			"name", svc.Name,
			"addr", addr)
		observer.OnAccept(svc)
	} else {
		d.logger.Info("zeroconf: service not auto-accepted, waiting for user",
			"name", svc.Name,
			"addr", addr)
	}
}
