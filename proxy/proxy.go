// Package proxy is a transparent proxy built on the stack/server
package proxy

import (
	"context"

	"github.com/stack-labs/stack/client"
	"github.com/stack-labs/stack/router"
	"github.com/stack-labs/stack/server"
	"github.com/stack-labs/stack/util/options"
)

// Proxy can be used as a proxy server for stack services
type Proxy interface {
	options.Options
	// ProcessMessage handles inbound messages
	ProcessMessage(context.Context, server.Message) error
	// ServeRequest handles inbound requests
	ServeRequest(context.Context, server.Request, server.Response) error
}

var (
	DefaultEndpoint = "localhost:9090"
)

// WithEndpoint sets a proxy endpoint
func WithEndpoint(e string) options.Option {
	return options.WithValue("proxy.endpoint", e)
}

// WithClient sets the client
func WithClient(c client.Client) options.Option {
	return options.WithValue("proxy.client", c)
}

// WithRouter specifies the router to use
func WithRouter(r router.Router) options.Option {
	return options.WithValue("proxy.router", r)
}

// WithLink sets a link for outbound requests
func WithLink(name string, c client.Client) options.Option {
	return func(o *options.Values) error {
		var links map[string]client.Client
		v, ok := o.Get("proxy.links")
		if ok {
			links = v.(map[string]client.Client)
		} else {
			links = map[string]client.Client{}
		}
		links[name] = c
		// save the links
		o.Set("proxy.links", links)
		return nil
	}
}
