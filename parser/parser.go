package parser

import "github.com/ChrisTrenkamp/xsel/node"

type Parser func() (node.Node, bool, error)
