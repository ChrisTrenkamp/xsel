package exec_test

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

func ExampleExec() {
	xml := `
<root>
	<a>This is an XML node.</a>
</root>
`

	xpath := grammar.MustBuild(`/root/a`)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, _ := store.CreateInMemory(parser)
	result, _ := exec.Exec(cursor, &xpath)

	fmt.Println(result)
}

func ExampleExec_namespace_binding() {
	xml := `
<root xmlns="http://some.namespace.com">
	<a xmlns="http://some.namespace.com">This is an XML node with a namespace prefix.</a>
</root>
`

	contextSettings := func(c *exec.ContextSettings) {
		c.NamespaceDecls["ns"] = "http://some.namespace.com"
	}

	xpath := grammar.MustBuild(`/ns:root/ns:a`)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, _ := store.CreateInMemory(parser)
	result, _ := exec.Exec(cursor, &xpath, contextSettings)

	fmt.Println(result)
}

func ExampleExec_custom_function() {
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

	fmt.Println(result)
}

func ExampleExec_custom_variables() {
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

	fmt.Println(result)
}
