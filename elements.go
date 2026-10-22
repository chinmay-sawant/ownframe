// Package ownframe re-exports the retained element model.
package ownframe

import "github.com/chinmay-sawant/ownframe/internal/element"

// Element is a retained element instance.
type Element = element.Node

// ElementKind is an element kind label.
type ElementKind = element.Kind

// ElementRegistry retains element instances by stable key.
type ElementRegistry = element.Registry

// ElementLayout holds CSS layout geometry apart from metadata.
type ElementLayout = element.Layout

// NewElement returns an empty retained node.
func NewElement(id string, kind ElementKind) *Element {
	return element.NewNode(id, kind)
}

// NewElementRegistry returns an empty retained-element registry.
func NewElementRegistry() *ElementRegistry {
	return element.NewRegistry()
}
