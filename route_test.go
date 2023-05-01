package routing_test

import (
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/routing"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouter(t *testing.T) {
	router := routing.NewRouter[*RouterTest]()

	for route, tests := range routes {
		signature, err := router.Add(route, tests)
		assert.NoError(t, err, signature, route)
		for _, errRoute := range tests.errRoutes {
			signature, err = router.Add(errRoute, tests)
			assert.Error(t, err, signature, route)
		}
	}

	for route, test := range routes {
		for path, params := range test.successfulPaths {
			_, results, err := router.Find(path)
			if err != nil {
				fmt.Println(route, path, err.Error())
			} else {
				for key, value := range params {
					if results[key] != value {
						assert.True(t, results[key] == value)
					}
				}
			}
			assert.NoError(t, err, err)

		}
		for _, path := range test.notFoundPaths {
			_, _, err := router.Find(path)
			if err == nil {
				fmt.Println(route, path)
			}
			assert.Error(t, err, err)
		}
	}
	fmt.Println(router)
}

var routes = map[string]*RouterTest{
	"/books/{name}_description": {
		successfulPaths: map[string]contracts.Fields{
			"/books/docker_description": {"name": "docker"},
			"/books/k8s_description":    {"name": "k8s"},
		},
		notFoundPaths: []string{
			"/books/docker/description",
			"/books/docker/_description",
		},
	},
	"/books1/{name?}_description": {
		successfulPaths: map[string]contracts.Fields{
			"/books1/docker_description": {"name": "docker"},
			"/books1/k8s_description":    {"name": "k8s"},
			"/books1/_description":       {"name": ""},
		},
		notFoundPaths: []string{
			"/books1/docker/description",
			"/books1/docker/_description",
		},
	},
	"/articles/first_{type}": {
		successfulPaths: map[string]contracts.Fields{
			"/articles/first_docker": {"type": "docker"},
			"/articles/first_k8s":    {"type": "k8s"},
		},
		notFoundPaths: []string{
			"/articles/first_/description",
			"/articles/first/_description",
		},
	},
	"/articles1/first_{type?}": {
		successfulPaths: map[string]contracts.Fields{
			"/articles1/first_docker": {"type": "docker"},
			"/articles/first_k8s":     {"type": "k8s"},
			"/articles/first_":        {"type": ""},
		},
		notFoundPaths: []string{
			"/articles1/first_/description",
			"/articles1/first/_description",
		},
	},
	"/users1/{name}/{level?}": {
		errRoutes: []string{"/users1/{xx}/{xxx:.*}"},
		successfulPaths: map[string]contracts.Fields{
			"/users1/xxx":       {"name": "xxx", "level": ""},
			"/users1/xx/sadad":  {"name": "xx", "level": "sadad"},
			"/users1/xx/sadad/": {"name": "xx", "level": "sadad"},
			"/users1/dd/":       {"name": "dd", "level": ""},
		},
		notFoundPaths: []string{
			"/users1/xxx/da/1",
		},
	},
	"/users": {
		successfulPaths: map[string]contracts.Fields{
			"/users":  {},
			"/users/": {},
		},
		notFoundPaths: []string{
			"/usersx",
		},
	},
	"/users/{name}": {
		successfulPaths: map[string]contracts.Fields{
			"/users/xxx": {"name": "xxx"},
		},
		notFoundPaths: []string{
			"/users/xxx/da",
		},
	},
	"/homepage/{name?}/hosts": {
		errRoutes: []string{"/homepage/{xx}/hosts"},
		successfulPaths: map[string]contracts.Fields{
			"/homepage/xxx/hosts": {"name": "xxx"},
			"/homepage/hosts":     {"name": ""},
		},
		notFoundPaths: []string{
			"/homepage/xxx/hosts1",
			"/homepage/hosts1",
		},
	},
	"/homepage/{name?}/news": {
		successfulPaths: map[string]contracts.Fields{
			"/homepage/xxx/news": {"name": "xxx"},
			"/homepage/news":     {"name": ""},
		},
		notFoundPaths: []string{
			"/homepage/xxx/news1",
			"/homepage/news1",
		},
	},
	"/category/{category:[0-9]+}/archive/{archive}": {
		successfulPaths: map[string]contracts.Fields{
			"/category/1/archive/any": {"category": "1", "archive": "any"},
		},
		notFoundPaths: []string{
			"/category/any/archive/any",
		},
	},
	"/category/{category}/posts/{archive:[0-9]+}": {
		successfulPaths: map[string]contracts.Fields{
			"/category/any/posts/1": {"category": "any", "archive": "1"},
		},
		notFoundPaths: []string{
			"/category/any/posts/any",
		},
	},
	"/posts/{id}": {
		successfulPaths: map[string]contracts.Fields{
			"/posts/first": {"id": "first"},
		},
		notFoundPaths: []string{
			"/posts/first/xxx",
		},
		errRoutes: []string{
			"/posts/{name}",
		},
	},
	"/archives/{id:[0-9]+?}": {
		successfulPaths: map[string]contracts.Fields{
			"/archives/1": {"id": "1"},
			"/archives/":  {"id": ""},
			"/archives":   {"id": ""},
		},
		notFoundPaths: []string{
			"/archives/any",
		},
	},
}

type RouterTest struct {
	successfulPaths map[string]contracts.Fields
	notFoundPaths   []string
	errRoutes       []string
}

func BenchmarkName(b *testing.B) {

	router := routing.NewRouter[*RouterTest]()

	for route, tests := range routes {
		router.Add(route, tests)
	}

	for i := 0; i < b.N; i++ {
		for _, test := range routes {
			for _, tests := range test.successfulPaths {
				for path := range tests {
					_, _, _ = router.Find(path)
				}
			}
		}
	}
}
