package app

import (
	auth_client "github.com/hse-telescope/gateway/internal/clients/auth"
	core_client "github.com/hse-telescope/gateway/internal/clients/core"
	"github.com/hse-telescope/gateway/internal/config"
	"github.com/hse-telescope/gateway/internal/providers/token"
	"github.com/hse-telescope/gateway/internal/server"
)

type Clients struct {
	auth auth_client.Wrapper
	core core_client.Wrapper
}

type Providers struct {
	token token.Provider
}

type App struct {
	server    *server.Server
	providers Providers
	clients   Clients
}

func newClients(conf config.Config) Clients {
	return Clients{
		auth: auth_client.New(conf.Clients.Auth.URL),
		core: core_client.New(conf.Clients.Core.URL),
	}
}

func newProviders(conf config.Config) Providers {
	return Providers{
		token: token.New(conf.PublicKey),
	}
}

func New(conf config.Config) *App {
	clients := newClients(conf)
	providers := newProviders(conf)
	return &App{
		server: server.New(
			conf,
			providers.token,
			clients.auth,
			clients.core,
		),
		clients:   clients,
		providers: providers,
	}
}

func (a *App) Start() error {
	return a.server.Start()
}
