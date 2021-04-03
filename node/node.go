package node

type Node interface{}

type Root interface {
	Node
}

type NamedNode interface {
	Space() string
	Local() string
}

type Element interface {
	Node
	NamedNode
}

type Namespace interface {
	Node
	Prefix() string
	NamespaceValue() string
}

type Attribute interface {
	Node
	NamedNode
	AttributeValue() string
}

type CharData interface {
	Node
	CharDataValue() string
}

type Comment interface {
	Node
	CommentValue() string
}

type ProcInst interface {
	Node
	Target() string
	ProcInstValue() string
}
