# xsel

[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate?business=PJDCE35ARU76Q&currency_code=USD) [![Go Reference](https://pkg.go.dev/badge/github.com/ChrisTrenkamp/xsel.svg)](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel)


`xsel` is a library that (almost) implements the XPath 1.0 specification.  The non-compliant bits are:

* `xsel` does not implement the [id](https://www.w3.org/TR/xpath-10/#function-id) function.
* The grammar as defined in the XPath 1.0 spec doesn't explicitly allow function calls in the middle of a path expression, such as `/path/function-call()/path`.  `xsel` allows function calls in the middle of path expressions.
* `xsel` allows name lookups with a wildcard for the namespace, such as `/*:path`.

## Basic usage

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

func main() {
	xml := `
<root>
	<a>This is an XML node.</a>
</root>
`

	xpath := grammar.MustBuild(`/root/a`)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, _ := store.CreateInMemory(parser)
	result, _ := exec.Exec(cursor, &xpath)

	fmt.Println(result) // This is an XML node.
}
```

`xsel` lets you define your own namespaces, methods, functions for use in your expressions:

## Binding variables and namespaces

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

func main() {
	xml := `
<root>
	<node>2.50</node>
	<node>3.14</node>
	<node>0.30</node>
</root>
`

	contextSettings := func(c *exec.ContextSettings) {
		c.NamespaceDecls["ns"] = "http://some.namespace.com"
		c.Variables[exec.Name("http://some.namespace.com", "mynum")] = exec.Number(3.14)
	}

	xpath := grammar.MustBuild(`//node()[. = $ns:mynum]`)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, _ := store.CreateInMemory(parser)
	result, _ := exec.Exec(cursor, &xpath, contextSettings)

	fmt.Println(result) //3.14
}
```

## Binding functions

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

func main() {
	xml := `
<root>
	<a>This is an element.</a>
	<!-- This is a comment. -->
</root>
`

	isComment := func(context exec.Context, args ...exec.Result) (exec.Result, error) {
		nodeSet, isNodeSet := context.Result().(exec.NodeSet)

		if !isNodeSet || len(nodeSet) == 0 {
			return exec.Bool(false), nil
		}

		_, isComment := nodeSet[0].Node().(node.Comment)
		return exec.Bool(isComment), nil
	}

	contextSettings := func(c *exec.ContextSettings) {
		c.FunctionLibrary[exec.Name("", "is-comment")] = isComment
	}

	xpath := grammar.MustBuild(`//node()[is-comment()]`)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, _ := store.CreateInMemory(parser)
	result, _ := exec.Exec(cursor, &xpath, contextSettings)

	fmt.Println(result) // This is a comment.
}
```

## Extensible

`xsel` supplies an XML parser (using the `encoding/xml` package) out of the box, but the XPath logic does not depend directly on XML.  It instead depends on the interfaces defined in the [node](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node) and [store](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/store) packages.  This means it's possible to use `xsel` for querying against non-XML documents, such as JSON.

To build a custom document, implement your own [Parser](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/parser#Parser) method, and build [Element](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Element)'s, [Attribute](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Attribute)'s [Character Data](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#CharData), [Comment](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Comment)'s, [Processing Instruction](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#ProcInst)'s, and [Namespace](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Namespace)'s.
