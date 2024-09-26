package routing

import (
	"errors"
	"fmt"
	"github.com/goal-web/container"
	"github.com/goal-web/contracts"
	"github.com/modood/table"
	"net/http"
	"net/url"
	"strings"
)

var (
	MiddlewareError = errors.New("middleware error") // 中间件必须有一个返回值
)

var (
	methodList = [...]string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
)

type HttpRouter struct {
	app          contracts.Application
	groups       []contracts.RouteGroup
	routes       []contracts.Route
	routers      map[string]contracts.Router[contracts.Route]
	hostsRouters contracts.Router[map[string]contracts.Router[contracts.Route]]

	// 全局中间件
	middlewares []contracts.MagicalFunc
}

func NewHttpRouter(app contracts.Application) contracts.HttpRouter {
	router := &HttpRouter{
		app:          app,
		routes:       make([]contracts.Route, 0),
		groups:       make([]contracts.RouteGroup, 0),
		middlewares:  make([]contracts.MagicalFunc, 0),
		routers:      map[string]contracts.Router[contracts.Route]{},
		hostsRouters: NewRouter[map[string]contracts.Router[contracts.Route]](),
	}

	return router
}

func (httpRouter *HttpRouter) addHostRoute(hostRouters map[string]map[string]contracts.Router[contracts.Route], route contracts.Route) []string {
	if host := route.GetHost(); host != "" {
		if hostRouters[host] == nil {
			hostRouters[host] = map[string]contracts.Router[contracts.Route]{}
		}

		return httpRouter.addRoute(hostRouters[host], route)
	}

	return nil
}

type RoutePrintItem struct {
	Path       string
	Method     string
	Controller string
	Middleware string
	Host       string
}

func (httpRouter *HttpRouter) Print() {
	var list []RoutePrintItem
	var routes = httpRouter.routes

	for _, group := range httpRouter.groups {
		routes = append(routes, group.Routes()...)
	}
	for _, router := range httpRouter.routers {
		routes = append(routes, router.All()...)
	}

	for _, router := range httpRouter.hostsRouters.All() {
		for _, subRouter := range router {
			routes = append(routes, subRouter.All()...)
		}
	}

	for _, route := range routes {
		var middlewares []string
		for _, mid := range route.Middlewares() {
			middlewares = append(middlewares, mid.Signature())
		}
		list = append(list, RoutePrintItem{
			Path:       route.GetPath(),
			Host:       route.GetHost(),
			Method:     strings.Join(route.Method(), ","),
			Controller: route.Handler().Signature(),
			Middleware: strings.Join(middlewares, ","),
		})
	}

	table.Output(list)
}

func (httpRouter *HttpRouter) addRoute(routers map[string]contracts.Router[contracts.Route], route contracts.Route) []string {
	var failedSignatures []string
	for _, method := range route.Method() {
		if routers[method] == nil {
			routers[method] = NewRouter[contracts.Route]()
		}

		signature, err := routers[method].Add(route.GetPath(), route)
		if err != nil {
			failedSignatures = append(failedSignatures, fmt.Sprintf("[%s] %s", method, signature))
		}
	}
	return failedSignatures
}

func (httpRouter *HttpRouter) Mount() error {
	var failedSignatures []string
	var hostRoutersMap = make(map[string]map[string]contracts.Router[contracts.Route])

	for _, route := range httpRouter.routes {
		tmpFailedSignatures := httpRouter.addRoute(httpRouter.routers, route)
		if len(tmpFailedSignatures) > 0 {
			failedSignatures = append(failedSignatures, tmpFailedSignatures...)
		}

		tmpFailedSignatures = httpRouter.addHostRoute(hostRoutersMap, route)
		if len(tmpFailedSignatures) > 0 {
			failedSignatures = append(failedSignatures, tmpFailedSignatures...)
		}
	}

	for _, group := range httpRouter.groups {
		for _, route := range group.Routes() {
			tmpFailedSignatures := httpRouter.addRoute(httpRouter.routers, route)
			if len(tmpFailedSignatures) > 0 {
				failedSignatures = append(failedSignatures, tmpFailedSignatures...)
			}
			tmpFailedSignatures = httpRouter.addHostRoute(hostRoutersMap, route)
			if len(tmpFailedSignatures) > 0 {
				failedSignatures = append(failedSignatures, tmpFailedSignatures...)
			}
		}
	}

	if len(hostRoutersMap) > 0 {
		httpRouter.hostsRouters = NewRouter[map[string]contracts.Router[contracts.Route]]()
		for host, router := range hostRoutersMap {
			signature, err := httpRouter.hostsRouters.Add(host, router)
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

func (httpRouter *HttpRouter) Add(method any, path string, handler any, middlewares ...any) contracts.Route {
	if strings.HasSuffix(path, "/") && path != "/" {
		path = path[:len(path)-1]
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

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

func (httpRouter *HttpRouter) Route(method string, url *url.URL) (contracts.Route, contracts.RouteParams, error) {
	route, params, err := httpRouter.route(method, url)
	if err == nil {
		return route, params, nil
	}
	for _, item := range methodList {
		if item == method {
			continue
		}
		route, params, err = httpRouter.route(item, url)
		if err == nil {
			return route, params, MethodNotAllowErr
		}
	}
	return nil, nil, NotFoundErr
}

func (httpRouter *HttpRouter) route(method string, url *url.URL) (contracts.Route, contracts.RouteParams, error) {
	path := url.Path
	if strings.HasSuffix(path, "/") && path != "/" {
		path = path[:len(path)-1]
	}

	if !httpRouter.hostsRouters.IsEmpty() {
		routers, hostParams, hostErr := httpRouter.hostsRouters.Find(url.Host)
		if hostErr == nil {
			if routers[method] != nil {
				route, params, err := routers[method].Find(path)
				if err == nil {
					for key, value := range hostParams {
						params[key] = value
					}
					return route, params, nil
				}
			}
		}
	}

	router := httpRouter.routers[method]
	if router == nil {
		return nil, nil, NotFoundErr
	}

	route, params, err := router.Find(path)
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

func (httpRouter *HttpRouter) GET(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodGet, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Post(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPost, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) POST(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPost, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Delete(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodDelete, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) DELETE(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodDelete, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Put(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPut, path, handler, middlewares...)
}
func (httpRouter *HttpRouter) PUT(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPut, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Patch(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPatch, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) PATCH(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodPatch, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Options(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodOptions, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) OPTIONS(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodOptions, path, handler, middlewares...)
}

func (httpRouter *HttpRouter) Trace(path string, handler any, middlewares ...any) contracts.Route {
	return httpRouter.Add(http.MethodTrace, path, handler, middlewares...)
}
func (httpRouter *HttpRouter) TRACE(path string, handler any, middlewares ...any) contracts.Route {
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
