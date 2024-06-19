package routing

import (
	"regexp"
)

type ParamNode[T any] struct {
	optional bool
	name     string
	rule     string
	reg      *regexp.Regexp
	nodes    map[string]Tree[T]
}

func NewRouteNode[T any](param string) *ParamNode[T] {
	name, rule, isOptional := parseRule(param)
	return &ParamNode[T]{
		rule:     rule,
		optional: isOptional,
		name:     name,
		reg:      regexp.MustCompile(rule),
		nodes:    make(map[string]Tree[T]),
	}
}

type Tree[T any] struct {
	Data  T
	Nodes []*ParamNode[T]
}

func (router *ParamNode[T]) IsSame(node *ParamNode[T]) bool {
	return router.optional == node.optional && router.rule == node.rule && node.name == router.name
}
