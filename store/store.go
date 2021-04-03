package store

import "github.com/ChrisTrenkamp/xsel/node"

type Cursor interface {
	Pos() int
	Node() node.Node
	Namespaces() []Cursor
	Attributes() []Cursor
	Nodes() []Cursor
	Parent() Cursor
}

func GetAttribute(c Cursor, space, local string) (node.Attribute, bool) {
	for _, a := range c.Attributes() {
		attr := a.Node().(node.Attribute)

		if attr.Space() == space && attr.Local() == local {
			return attr, true
		}
	}

	return nil, false
}
