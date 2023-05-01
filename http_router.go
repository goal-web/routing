package routing

import (
	"errors"
	"fmt"
	"github.com/goal-web/container"
	"github.com/goal-web/contracts"
	"net/http"
	"net/url"
	"strings"
)

var (
	MiddlewareError = errors.New("middleware error") // 中间件必须有一个返回值
)

type HttpRouter struct {
	app          contracts.Application
	engine       contracts.HttpEngine
	groups       []contracts.RouteGroup
	routes       []contracts.Route
	router       contracts.Router[contracts.Route]
	hostsRouters contracts.Router[contracts.Router[contracts.Route]]

	// 全局中间件
	middlewares []contracts.MagicalFunc
}

func NewHttpRouter(app contracts.Application) contracts.HttpRouter {
	router := &HttpRouter{
		app:          app,
		routes:       make([]contracts.Route, 0),
		groups:       make([]contracts.RouteGroup, 0),
		middlewares:  make([]contracts.MagicalFunc, 0),
		router:       NewRouter[contracts.Route](),
		hostsRouters: NewRouter[contracts.Router[contracts.Route]](),
	}

	return router
}

func (httpRouter *HttpRouter) addHostRoute(hostRouters map[string]contracts.Router[contracts.Route], route contracts.Route) (string, error) {
	if host := route.GetHost(); host != "" {
		if hostRouters[host] == nil {
			hostRouters[host] = NewRouter[contracts.Route]()
		}
		return hostRouters[host].Add(route.GetPath(), route)
	}

	return "", nil
}

func (httpRouter *HttpRouter) Mount() error {
	var failedSignatures []string
	var signature string
	var err error
	var hostRoutersMap = make(map[string]contracts.Router[contracts.Route])

	for _, route := range httpRouter.routes {
		signature, err = httpRouter.router.Add(route.GetPath(), route)
		if err != nil {
			failedSignatures = append(failedSignatures, signature)
		}

		if signature, err = httpRouter.addHostRoute(hostRoutersMap, route); err != nil {
			failedSignatures = append(failedSignatures, signature)
		}
	}

	for _, group := range httpRouter.groups {
		for _, route := range group.Routes() {
			signature, err = httpRouter.router.Add(route.GetPath(), route)
			if err != nil {
				failedSignatures = append(failedSignatures, signature)
			}
			if signature, err = httpRouter.addHostRoute(hostRoutersMap, route); err != nil {
				failedSignatures = append(failedSignatures, signature)
			}
		}
	}

	if len(hostRoutersMap) > 0 {
		httpRouter.hostsRouters = NewRouter[contracts.Router[contracts.Route]]()
		for host, router := range hostRoutersMap {
			signature, err = httpRouter.hostsRouters.Add(host, router)
			if err != nil {
				failedSignatures = append(failedSignatures, signature)
			}
		}
	}

	if len(failedSignatures) > 0 {
		return fmt.Errorf("duplicate route [%s] occurred", strings.Join(failedSignatures, "|"))
	}
	return nil
}

func (httpRouter *HttpRouter) Route(url *url.URL) (contracts.Route, contracts.RouteParams, error) {
	if !httpRouter.hostsRouters.IsEmpty() {
		router, hostParams, hostErr := httpRouter.hostsRouters.Find(url.Host)
		if hostErr == nil {
			route, params, err := router.Find(url.Path)
			if err == nil {
				for key, value := range hostParams {
					params[key] = value
				}
				return route, params, nil
			}
		}
	}

	route, params, err := httpRouter.router.Find(url.Path)
	if err != nil {
		return nil, nil, err
	}

	return route, params, err
}

func (httpRouter *HttpRouter) Group(prefix string, middlewares ...any) contracts.RouteGroup {
	groupInstance := NewGroup(prefix, middlewares...)

	httpRouter.groups = append(httpRouter.groups, groupInstance)

	return groupInstance
}

func (httpRouter *HttpRouter) Get(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodGet, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Post(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPost, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Delete(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodDelete, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Put(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPut, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Patch(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPatch, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Options(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodOptions, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Trace(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodTrace, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Use(middlewares ...any) {
	for _, middleware := range middlewares {
		if magicalFunc, ok := middleware.(contracts.MagicalFunc); ok {
			httpRouter.middlewares = append(httpRouter.middlewares, magicalFunc)
		} else {
			httpRouter.middlewares = append(httpRouter.middlewares, container.NewMagicalFunc(middleware))
		}
	}
}

func (httpRouter *HttpRouter) Middlewares() []contracts.MagicalFunc {
	return httpRouter.middlewares
}

func (httpRouter *HttpRouter) Add(method any, path string, handler any, middlewares ...any) contracts.Route {
	methods := make([]string, 0)
	switch v := method.(type) {
	case string:
		methods = []string{v}
	case []string:
		methods = v
	}
	route := NewRoute(methods, path, ConvertToMiddlewares(middlewares...), container.NewMagicalFunc(handler))
	httpRouter.routes = append(httpRouter.routes, route)
	return route
}
