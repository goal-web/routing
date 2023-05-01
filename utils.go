package routing

import (
	"github.com/goal-web/container"
	"github.com/goal-web/contracts"
	"strings"
)

func parseRule(param string) (string, string, bool) {
	name := param[1 : len(param)-1]
	isOptional := name[len(name)-1:] == "?"
	if isOptional {
		name = name[:len(name)-1]
	}
	rule := ""
	items := strings.Split(name, ":")
	itemsLen := len(items)
	if itemsLen > 1 {
		rule = strings.Join(items[1:itemsLen], ":")
		name = items[0]
	} else {
		rule = ".*"
	}
	return name, rule, isOptional
}

func parseRoute(route string) ([]string, string) {
	params := paramReg.FindAllStringIndex(route, -1)
	var signature string
	var results []string
	var routeLen = len(route)
	var end int
	for i, param := range params {
		if i == 0 && param[0] != 0 {
			paramStr := route[:param[0]]
			results = append(results, paramStr)
			signature += paramStr
		}

		if i != 0 && param[0]-end > 0 {
			paramStr := route[end:param[0]]
			results = append(results, paramStr)
			signature += paramStr
		}

		paramStr := route[param[0]:param[1]]
		_, rule, _ := parseRule(paramStr)
		signature += rule

		results = append(results, paramStr)

		end = param[1]
	}
	if end < routeLen {
		paramStr := route[end:]
		results = append(results, paramStr)
		signature += paramStr
	}
	return results, signature
}

func ConvertToMiddlewares(middlewares ...any) (results []contracts.MagicalFunc) {
	for _, middleware := range middlewares {
		magicalFunc, isMiddleware := middleware.(contracts.MagicalFunc)
		if !isMiddleware {
			magicalFunc = container.NewMagicalFunc(middleware)
		}
		if magicalFunc.NumOut() != 1 {
			panic(MiddlewareError)
		}
		results = append(results, magicalFunc)
	}
	return
}
