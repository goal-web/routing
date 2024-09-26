package routing

import "github.com/goal-web/contracts"

type ServiceProvider struct {
}

func NewService() contracts.ServiceProvider {
	return ServiceProvider{}
}

func (s ServiceProvider) Register(application contracts.Application) {
	application.Call(func(console contracts.Console) {
		console.RegisterCommand("route:list", NewRouteList)
	})
	application.Singleton("HttpRouter", func(console contracts.Console) contracts.HttpRouter {
		return NewHttpRouter(application)
	})
}

func (s ServiceProvider) Start() error {
	return nil
}

func (s ServiceProvider) Stop() {
}
