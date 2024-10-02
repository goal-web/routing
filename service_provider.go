package routing

import "github.com/goal-web/contracts"

type ServiceProvider struct {
}

func NewService() contracts.ServiceProvider {
	return ServiceProvider{}
}

func (s ServiceProvider) Register(application contracts.Application) {

}

func (s ServiceProvider) Start() error {
	return nil
}

func (s ServiceProvider) Stop() {
}
