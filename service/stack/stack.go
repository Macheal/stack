package stack

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/stack-labs/stack-rpc/client"
	"github.com/stack-labs/stack-rpc/cmd"
	"github.com/stack-labs/stack-rpc/debug/profile"
	"github.com/stack-labs/stack-rpc/debug/profile/pprof"
	"github.com/stack-labs/stack-rpc/debug/service/handler"
	"github.com/stack-labs/stack-rpc/env"
	"github.com/stack-labs/stack-rpc/server"
	"github.com/stack-labs/stack-rpc/service"
	"github.com/stack-labs/stack-rpc/util/log"
	"github.com/stack-labs/stack-rpc/util/wrapper"
)

type stackService struct {
	opts service.Options

	once sync.Once
}

func (s *stackService) Name() string {
	return s.opts.Server.Options().Name
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *stackService) Init(opts ...service.Option) error {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}

	// service name
	serviceName := s.opts.Server.Options().Name

	// wrap client to inject From-Service header on any calls
	s.opts.Client = wrapper.FromService(serviceName, s.opts.Client)

	if err := s.opts.Cmd.Init(
		cmd.Broker(&s.opts.Broker),
		cmd.Registry(&s.opts.Registry),
		cmd.Transport(&s.opts.Transport),
		cmd.Client(&s.opts.Client),
		cmd.Server(&s.opts.Server),
		cmd.Selector(&s.opts.Selector),
		cmd.Logger(&s.opts.Logger),
		cmd.Config(&s.opts.Config),
		cmd.Auth(&s.opts.Auth),
	); err != nil {
		log.Errorf("cmd init error: %s", err)
		return err
	}

	return nil
}

func (s *stackService) Options() service.Options {
	return s.opts
}

func (s *stackService) Client() client.Client {
	return s.opts.Client
}

func (s *stackService) Server() server.Server {
	return s.opts.Server
}

func (s *stackService) String() string {
	return "stack"
}

func (s *stackService) Start() error {
	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.opts.Server.Start(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *stackService) Stop() error {
	var gerr error

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	if err := s.opts.Server.Stop(); err != nil {
		return err
	}

	if err := s.opts.Config.Close(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (s *stackService) Run() error {
	// register the debug handler
	if err := s.opts.Server.Handle(
		s.opts.Server.NewHandler(
			handler.DefaultHandler,
			server.InternalHandler(true),
		),
	); err != nil {
		return err
	}

	// start the profiler
	// TODO: set as an option to the service, don't just use pprof
	if prof := os.Getenv(env.StackDebugProfile); len(prof) > 0 {
		service := s.opts.Server.Options().Name
		version := s.opts.Server.Options().Version
		id := s.opts.Server.Options().Id
		profiler := pprof.NewProfile(
			profile.Name(service + "." + version + "." + id),
		)
		if err := profiler.Start(); err != nil {
			return err
		}
		defer profiler.Stop()
	}

	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	if s.opts.Signal {
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	}

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-s.opts.Context.Done():
	}

	return s.Stop()
}

func NewService(opts ...service.Option) service.Service {
	options := service.NewOptions(opts...)

	return &stackService{
		opts: options,
	}
}
