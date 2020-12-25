package http

import (
	"github.com/stack-labs/stack-rpc/broker"
	"github.com/stack-labs/stack-rpc/plugin"
)

type httpBrokerPlugin struct {
}

func (h *httpBrokerPlugin) Name() string {
	return "http"
}

func (h *httpBrokerPlugin) Options() []broker.Option {
	return nil
}

func (h *httpBrokerPlugin) New(opts ...broker.Option) broker.Broker {
	return NewBroker(opts...)
}

func init() {
	plugin.BrokerPlugins["http"] = &httpBrokerPlugin{}
}
