package routing

import (
	"regexp"
)

type RouterNode[T any] struct {
	data     T
	optional bool
	name     string
	rule     string
	reg      *regexp.Regexp
	nodes    map[string][]*RouterNode[T]
}

func NewRouteNode[T any](param string, data T) *RouterNode[T] {
	name, rule, isOptional := parseRule(param)
	return &RouterNode[T]{
		rule:     rule,
		data:     data,
		optional: isOptional,
		name:     name,
		reg:      regexp.MustCompile(rule),
		nodes:    make(map[string][]*RouterNode[T]),
	}
}

func (router *RouterNode[T]) IsSame(node *RouterNode[T]) bool {
	return router.optional == node.optional && router.rule == node.rule && node.name == router.name
}
