package routing

import (
	"errors"
	"github.com/goal-web/container"
	"github.com/goal-web/contracts"
	"github.com/labstack/echo/v4"
)

var (
	MethodTypeError = errors.New("http method type unknown")
)

type Group struct {
	prefix      string
	host        string
	middlewares []contracts.MagicalFunc
	routes      []contracts.Route
	groups      []contracts.RouteGroup
}

func (group *Group) GetHost() string {
	return group.host
}

func (group *Group) Host(host string) contracts.RouteGroup {
	group.host = host
	return group
}

func NewGroup(prefix string, middlewares ...any) contracts.RouteGroup {
	return &Group{
		prefix:      prefix,
		routes:      make([]contracts.Route, 0),
		groups:      make([]contracts.RouteGroup, 0),
		middlewares: ConvertToMiddlewares(middlewares...),
	}
}

// Group 添加一个子组
func (group *Group) Group(prefix string, middlewares ...any) contracts.RouteGroup {
	var groupInstance = &Group{
		prefix:      group.prefix + prefix,
		routes:      make([]contracts.Route, 0),
		groups:      make([]contracts.RouteGroup, 0),
		middlewares: append(group.middlewares, ConvertToMiddlewares(middlewares...)...),
	}

	group.groups = append(group.groups, groupInstance)

	return groupInstance
}

// Add 添加路由，method 只允许字符串或者字符串数组
func (group *Group) Add(method any, path string, handler any, middlewares ...any) contracts.RouteGroup {
	methods := make([]string, 0)
	switch r := method.(type) {
	case string:
		methods = []string{r}
	case []string:
		methods = r
	default:
		panic(MethodTypeError)
	}
	group.routes = append(group.routes, &Route{
		method:      methods,
		path:        group.prefix + path,
		middlewares: append(group.middlewares, ConvertToMiddlewares(middlewares...)...),
		handler:     container.NewMagicalFunc(handler),
	})

	return group
}

func (group *Group) Get(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.GET, path, handler, middlewares...)
}

func (group *Group) Post(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.POST, path, handler, middlewares...)
}

func (group *Group) Delete(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.DELETE, path, handler, middlewares...)
}

func (group *Group) Put(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.PUT, path, handler, middlewares...)
}

func (group *Group) Trace(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.TRACE, path, handler, middlewares...)
}

func (group *Group) Patch(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.PATCH, path, handler, middlewares...)
}

func (group *Group) Options(path string, handler any, middlewares ...any) contracts.RouteGroup {
	return group.Add(echo.OPTIONS, path, handler, middlewares...)
}

func (group *Group) Routes() []contracts.Route {
	routes := group.routes

	for _, subGroup := range group.groups {
		routes = append(routes, subGroup.Routes()...)
	}

	return routes
}
