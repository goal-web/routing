package routing

import (
	"errors"
	"github.com/goal-web/contracts"
	"regexp"
	"strings"
)

var paramReg = regexp.MustCompile(`{([^{}]+)}`)
var NotFound = errors.New("route not found")
var RouteHasExists = errors.New("route is exists")

type Router[T any] struct {
	paths        map[string]T
	paramsRoutes map[string][]*RouterNode[T]
	signatures   map[string]struct{}
}

func NewRouter[T any]() contracts.Router[T] {
	return &Router[T]{
		paths:        map[string]T{},
		paramsRoutes: map[string][]*RouterNode[T]{},
		signatures:   map[string]struct{}{},
	}
}

func (router *Router[T]) IsEmpty() bool {
	return len(router.signatures) == 0
}

func (router *Router[T]) Find(path string) (T, contracts.RouteParams, error) {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	result, ok := router.paths[path]
	if ok {
		return result, nil, nil
	}

	var params = make(contracts.RouteParams)
	tmpResult, err := router.find(path, router.paramsRoutes, params)
	return tmpResult, params, err
}

func (router *Router[T]) find(path string, tree map[string][]*RouterNode[T], params contracts.RouteParams) (T, error) {
	for prefix, nodes := range tree {
		index := strings.Index(path, prefix)
		if index != 0 {
			if prefix[len(prefix)-1:] != "/" || strings.Index(path+"/", prefix) != 0 {
				continue
			} else {
				path += "/"
			}
		}
		value := path[len(prefix):]
		for _, node := range nodes {
			if len(node.nodes) == 0 {
				if !strings.Contains(value, "/") && (node.reg.MatchString(value) || (node.optional && value == "")) {
					params[node.name] = value
					return node.data, nil
				}
			} else {
				for subPrefix, subNodes := range node.nodes {
					if subPrefix == "/" {
						values := strings.Split(value, "/")
						if node.reg.MatchString(values[0]) {
							params[node.name] = values[0]
							value = "/" + strings.Join(values[1:], "/")
							result, err := router.find(value, node.nodes, params)
							if err == nil {
								return result, nil
							}
						}
					}
					index = strings.Index(value, subPrefix)
					if index > -1 {
						subValue := value[:index]
						if !strings.Contains(subValue, "/") && (node.reg.MatchString(subValue) || node.optional) {
							params[node.name] = subValue

							value = value[index:]
							if len(subNodes) == 0 && value == subPrefix {
								return node.data, nil
							}

							result, err := router.find(value, node.nodes, params)
							if err == nil {
								return result, nil
							}
						}
					} else if "/"+value == subPrefix && node.optional && len(subNodes) == 0 {
						params[node.name] = ""
						return node.data, nil
					}
				}
			}
		}
	}
	var result T
	return result, NotFound
}

func (router *Router[T]) Add(route string, data T) (string, error) {
	results, signature := parseRoute(route)
	if _, exists := router.signatures[signature]; exists {
		return signature, RouteHasExists
	}

	if len(results) == 1 {
		router.paths[results[0]] = data
	} else {
		tmpTree := router.paramsRoutes
		var prefix string
		for _, param := range results {
			if !(strings.HasPrefix(param, "{") && strings.HasSuffix(param, "}")) {
				prefix = param
				if tmpTree[prefix] == nil {
					tmpTree[prefix] = make([]*RouterNode[T], 0)
				}
				continue
			}

			node := NewRouteNode(param, data)
			nodes := tmpTree[prefix]

			if len(nodes) == 0 {
				tmpTree[prefix] = append(nodes, node)
				tmpTree = node.nodes
			} else {
				exists := false
				for _, item := range tmpTree[prefix] {
					if item.IsSame(node) {
						tmpTree = item.nodes
						exists = true
					}
				}
				if !exists {
					tmpTree[prefix] = append(nodes, node)
					tmpTree = node.nodes
				}
			}
		}
	}

	router.signatures[signature] = struct{}{}
	return signature, nil
}
