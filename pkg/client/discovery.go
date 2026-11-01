package client

import "github.com/kurikla/clipik-s/pkg/model"

type Discoverer interface {
	Start(observer *Observer) error
	Stop() error
}

type Filter interface {
	Accept(service model.Service) bool
}
