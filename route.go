package routing

import (
	"github.com/goal-web/contracts"
)

type Route struct {
	method      []string
	path        string
	middlewares []contracts.MagicalFunc
	handler     contracts.MagicalFunc
	name        string
	host        string
}

func NewRoute(method []string, path string, middlewares []contracts.MagicalFunc, handler contracts.MagicalFunc) contracts.Route {
	return &Route{
		method:      method,
		path:        path,
		middlewares: middlewares,
		handler:     handler,
	}
}

func (route *Route) Name(name string) contracts.Route {
	route.name = name
	return route
}

func (route *Route) Host(host string) contracts.Route {
	route.host = host
	return route
}

func (route *Route) Middlewares() []contracts.MagicalFunc {
	return route.middlewares
}

func (route *Route) Method() []string {
	return route.method
}

func (route *Route) GetPath() string {
	return route.path
}

func (route *Route) GetHost() string {
	return route.host
}

func (route *Route) GetName() string {
	return route.name
}

func (route *Route) Handler() contracts.MagicalFunc {
	return route.handler
}
