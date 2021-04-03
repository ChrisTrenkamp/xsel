package exec

import (
	"bytes"
	"math"
	"testing"

	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

func exec(t *testing.T, expr, xml string, expected Result) {
	xpath := grammar.MustBuild(expr)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, err := store.CreateInMemory(parser)

	if err != nil {
		t.Error(err)
		return
	}

	result, err := Exec(cursor, &xpath)

	if err != nil {
		t.Error(err)
	}

	if result != expected {
		t.Errorf("Result != '%s'. Received '%s'", expected, result)
	}
}

func execNodesToString(t *testing.T, expr, xml string, expected string) {
	xpath := grammar.MustBuild(expr)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, err := store.CreateInMemory(parser)

	if err != nil {
		t.Error(err)
		return
	}

	result, err := Exec(cursor, &xpath)

	if err != nil {
		t.Error(err)
	}

	resultString := result.String()

	if resultString != expected {
		t.Errorf("Result != '%s'. Received '%s'", expected, resultString)
	}
}

func execNodes(t *testing.T, expr, xml string, settings ...ContextApply) NodeSet {
	xpath := grammar.MustBuild(expr)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, err := store.CreateInMemory(parser)

	if err != nil {
		t.Error(err)
		return nil
	}

	result, err := Exec(cursor, &xpath, settings...)

	if err != nil {
		t.Error(err)
	}

	return result.(NodeSet)
}

func TestInfinity(t *testing.T) {
	num := Number(math.Inf(1))

	if num.String() != "Infinity" {
		t.Error("Not Infinity")
	}

	num = Number(math.Inf(-1))

	if num.String() != "-Infinity" {
		t.Error("Not -Infinity")
	}
}

func TestAdd(t *testing.T) {
	exec(t, "1.2+2.3", `<root/>`, Number(3.5))
}

func TestSubtract(t *testing.T) {
	exec(t, "5-3", `<root/>`, Number(2))
}

func TestMultiply(t *testing.T) {
	exec(t, "3*4", `<root/>`, Number(12))
}

func TestDivide(t *testing.T) {
	exec(t, "15 div 3", `<root/>`, Number(5))
	execNodesToString(t, "0 div 0", `<root/>`, Number(math.NaN()).String())
	exec(t, "1 div 0", `<root/>`, Number(math.Inf(1)))
	exec(t, "-1 div 0", `<root/>`, Number(math.Inf(-1)))
}

func TestMod(t *testing.T) {
	exec(t, "4 mod 3", `<root/>`, Number(1))
}

func TestString(t *testing.T) {
	exec(t, "'foo'", `<root/>`, String("foo"))
}

func TestNegate(t *testing.T) {
	exec(t, "8", `<root/>`, Number(8))
	exec(t, "-8", `<root/>`, Number(-8))
	exec(t, "--8", `<root/>`, Number(8))
	exec(t, "---8", `<root/>`, Number(-8))
}

func TestEquality(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<b>b</b>
	<one>1</one>
</root>
`
	exec(t, "/root/a = /root/b", xml, Bool(false))
	exec(t, "/root/a = /root/a", xml, Bool(true))

	exec(t, "/root/one = 1", xml, Bool(true))
	exec(t, "1 = /root/one", xml, Bool(true))
	exec(t, "2 = /root/one", xml, Bool(false))
	exec(t, "/root/one = 2", xml, Bool(false))

	exec(t, "/root/a = 'a'", xml, Bool(true))
	exec(t, "'a' = /root/a", xml, Bool(true))
	exec(t, "'b' = /root/a", xml, Bool(false))
	exec(t, "/root/a = 'b'", xml, Bool(false))

	exec(t, "1 = 1", xml, Bool(true))
	exec(t, "1 = 2", xml, Bool(false))

	exec(t, "1 = '1'", xml, Bool(true))
	exec(t, "1 = '2'", xml, Bool(false))

	exec(t, "'1' = '1'", xml, Bool(true))
	exec(t, "'1' = '2'", xml, Bool(false))

	exec(t, "/root/a = true()", xml, Bool(true))
	exec(t, "true() = /root/a", xml, Bool(true))
	exec(t, "true() = 1", xml, Bool(true))
	exec(t, "true() = 0", xml, Bool(false))
}

func TestNotEqual(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<b>b</b>
	<one>1</one>
</root>
`
	exec(t, "/root/a != /root/b", xml, Bool(true))
	exec(t, "/root/a != /root/a", xml, Bool(false))

	exec(t, "/root/one != 1", xml, Bool(false))
	exec(t, "1 != /root/one", xml, Bool(false))
	exec(t, "2 != /root/one", xml, Bool(true))
	exec(t, "/root/one != 2", xml, Bool(true))

	exec(t, "/root/a != 'a'", xml, Bool(false))
	exec(t, "'a' != /root/a", xml, Bool(false))
	exec(t, "'b' != /root/a", xml, Bool(true))
	exec(t, "/root/a != 'b'", xml, Bool(true))

	exec(t, "1 != 1", xml, Bool(false))
	exec(t, "1 != 2", xml, Bool(true))

	exec(t, "1 != '1'", xml, Bool(false))
	exec(t, "1 != '2'", xml, Bool(true))

	exec(t, "'1' != '1'", xml, Bool(false))
	exec(t, "'1' != '2'", xml, Bool(true))

	exec(t, "/root/a != true()", xml, Bool(false))
	exec(t, "true() != /root/a", xml, Bool(false))
	exec(t, "true() != 1", xml, Bool(false))
	exec(t, "true() != 0", xml, Bool(true))
}

func TestLessThan(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	exec(t, "/root/one < /root/two", xml, Bool(true))
	exec(t, "1 < /root/two", xml, Bool(true))
	exec(t, "/root/two < /root/one", xml, Bool(false))

	exec(t, "/root/two < 1", xml, Bool(false))
	exec(t, "3 < /root/two", xml, Bool(false))
	exec(t, "/root/one < 2", xml, Bool(true))

	exec(t, "'1' < /root/two", xml, Bool(true))
	exec(t, "/root/one < '2'", xml, Bool(true))
	exec(t, "'3' < /root/two", xml, Bool(false))
	exec(t, "/root/two < '1'", xml, Bool(false))

	exec(t, "'1' < '2'", xml, Bool(true))
}

func TestLessThanOrEqual(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	exec(t, "/root/one <= /root/two", xml, Bool(true))
	exec(t, "/root/two <= /root/two", xml, Bool(true))
	exec(t, "1 <= /root/two", xml, Bool(true))
	exec(t, "2 <= /root/two", xml, Bool(true))
	exec(t, "/root/two <= /root/one", xml, Bool(false))
	exec(t, "/root/two <= /root/two", xml, Bool(true))

	exec(t, "/root/two <= 1", xml, Bool(false))
	exec(t, "3 <= /root/two", xml, Bool(false))
	exec(t, "/root/one <= 2", xml, Bool(true))
	exec(t, "/root/two <= 2", xml, Bool(true))

	exec(t, "'1' <= /root/two", xml, Bool(true))
	exec(t, "'2' <= /root/two", xml, Bool(true))
	exec(t, "/root/one <= '2'", xml, Bool(true))
	exec(t, "/root/two <= '2'", xml, Bool(true))
	exec(t, "'3' <= /root/two", xml, Bool(false))
	exec(t, "/root/two <= '1'", xml, Bool(false))

	exec(t, "'1' <= '2'", xml, Bool(true))
	exec(t, "'2' <= '2'", xml, Bool(true))
}

func TestGreaterThan(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	exec(t, "/root/one > /root/two", xml, Bool(false))
	exec(t, "1 > /root/two", xml, Bool(false))
	exec(t, "/root/two > /root/one", xml, Bool(true))

	exec(t, "/root/two > 1", xml, Bool(true))
	exec(t, "3 > /root/two", xml, Bool(true))
	exec(t, "/root/one > 2", xml, Bool(false))

	exec(t, "'1' > /root/two", xml, Bool(false))
	exec(t, "/root/one > '2'", xml, Bool(false))
	exec(t, "'3' > /root/two", xml, Bool(true))
	exec(t, "/root/two > '1'", xml, Bool(true))

	exec(t, "'1' > '2'", xml, Bool(false))
}

func TestGreaterThanOrEqual(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	exec(t, "/root/one >= /root/two", xml, Bool(false))
	exec(t, "1 >= /root/two", xml, Bool(false))
	exec(t, "/root/two >= /root/one", xml, Bool(true))
	exec(t, "/root/two >= /root/two", xml, Bool(true))

	exec(t, "/root/two >= 1", xml, Bool(true))
	exec(t, "/root/two >= 2", xml, Bool(true))
	exec(t, "3 >= /root/two", xml, Bool(true))
	exec(t, "2 >= /root/two", xml, Bool(true))
	exec(t, "/root/one >= 2", xml, Bool(false))

	exec(t, "'1' >= /root/two", xml, Bool(false))
	exec(t, "/root/one >= '2'", xml, Bool(false))
	exec(t, "'3' >= /root/two", xml, Bool(true))
	exec(t, "'2' >= /root/two", xml, Bool(true))
	exec(t, "/root/two >= '1'", xml, Bool(true))
	exec(t, "/root/two >= '2'", xml, Bool(true))

	exec(t, "'1' >= '2'", xml, Bool(false))
	exec(t, "'2' >= '2'", xml, Bool(true))
}

func TestOr(t *testing.T) {
	exec(t, "1 or 0", `<root/>`, Bool(true))
	exec(t, "0 or 0", `<root/>`, Bool(false))
}

func TestAnd(t *testing.T) {
	exec(t, "1 and 0", `<root/>`, Bool(false))
	exec(t, "1 and 1", `<root/>`, Bool(true))
}

func TestAbsoluteLocationPathOnly(t *testing.T) {
	execNodesToString(t, "/", `b <root>a root node</root> c`, "b a root node c")
}

func TestAbsoluteLocationPathWithRelative(t *testing.T) {
	xml := `
a root node
<Node>node value</Node>
other text
`
	execNodesToString(t, "/ Node", xml, "node value")
}

func TestRelativeLocationPath(t *testing.T) {
	xml := `
text
<Root>text2
<node>a</node>
<node>b</node>
text3
</Root>
text4
`
	execNodesToString(t, "/Root/node", xml, "a")
}

func TestPredicate(t *testing.T) {
	xml := `
text
<Root>text2
<Node>a</Node>
<Node>b</Node>
text3
</Root>
text4
`
	execNodesToString(t, "/Root/Node[2]", xml, "b")
}

func TestUnion(t *testing.T) {
	xml := `
text
<Root>text2
<Node>a</Node>
<Node>b</Node>
text3
</Root>
text4
`
	nodes := execNodes(t, "/ Root/ Node [ 1 ] | /Root/Node[2]", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	a := NodeSet{nodes[0]}
	b := NodeSet{nodes[1]}

	if a.String() != "a" {
		t.Error("Node not 'a'")
	}

	if b.String() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestUnionNoDuplicates(t *testing.T) {
	xml := `
text
<Root>text2
<Node>a</Node>
<Node>b</Node>
text3
</Root>
text4
`
	nodes := execNodes(t, "/Root/Node[1] | /Root/Node[1]", xml)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	a := NodeSet{nodes[0]}

	if a.String() != "a" {
		t.Error("Node not 'a'")
	}
}

func TestNodeTest(t *testing.T) {
	xml := `
<root>foo<node>bar</node></root>
`
	execNodesToString(t, "/root/node ( ) ", xml, "foo")

	xml = `
<!--some comment-->
<comment>node</comment>
`
	execNodesToString(t, "/comment ( ) ", xml, "some comment")
	execNodesToString(t, "/comment", xml, "node")

	xml = `
<?foo bar?>
<processing-instruction>proc</processing-instruction>
<?eggs spam?>
`
	execNodesToString(t, "/processing-instruction ( ) ", xml, "bar")
	execNodesToString(t, "/processing-instruction ( 'eggs' ) ", xml, "spam")
	execNodesToString(t, "/processing-instruction", xml, "proc")

	xml = `some text<text>other text</text>`
	execNodesToString(t, "/text ( ) ", xml, "some text")
	execNodesToString(t, "/text", xml, "other text")
}

func TestAnyElement(t *testing.T) {
	xml := `
<root>root text<data>data text</data></root>
`
	execNodesToString(t, "/root/*", xml, "data text")
}

func TestChild(t *testing.T) {
	xml := `
a root node
<Node>node value</Node>
other text
`
	execNodesToString(t, "/child::Node", xml, "node value")
}

func TestAnyAttr(t *testing.T) {
	xml := `
<root foo="bar" eggs="ham"></root>
`
	nodes := execNodes(t, "/root/attribute::*", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	foo := NodeSet{nodes[0]}
	eggs := NodeSet{nodes[1]}

	if foo.String() != "bar" {
		t.Error("Node not 'bar'")
	}

	if eggs.String() != "ham" {
		t.Error("Node not 'ham'")
	}
}

func TestAttrAbbreviated(t *testing.T) {
	xml := `
<root foo="bar" eggs="ham"></root>
`
	nodes := execNodes(t, "/root/@eggs", xml)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	eggs := NodeSet{nodes[0]}

	if eggs.String() != "ham" {
		t.Error("Node not 'ham'")
	}
}

func TestAncestor(t *testing.T) {
	xml := `
<root>
	<a>
		<should-not-appear/>
		<b>
			<should-not-appear/>
		</b>
		<should-not-appear/>
	</a>
</root>
`
	nodes := execNodes(t, "/root/a/b/ancestor::*", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	a := nodes[0]
	root := nodes[1]

	if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
		t.Error("Node not 'a'")
	}

	if (store.Cursor)(root).Node().(node.Element).Local() != "root" {
		t.Error("Node not 'root'")
	}
}

func TestAncestorOrSelf(t *testing.T) {
	xml := `
<root>
	<a>
		<should-not-appear/>
		<b>
			<should-not-appear/>
		</b>
		<should-not-appear/>
	</a>
</root>
`
	nodes := execNodes(t, "/root/a/b/ancestor-or-self::*", xml)

	if len(nodes) != 3 {
		t.Error("Size is not 3")
	}

	b := nodes[0]
	a := nodes[1]
	root := nodes[2]

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}

	if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
		t.Error("Node not 'a'")
	}

	if (store.Cursor)(root).Node().(node.Element).Local() != "root" {
		t.Error("Node not 'root'")
	}
}

func TestDescendent(t *testing.T) {
	xml := `<root><a><b/></a></root>`
	nodes := execNodes(t, "/root/descendant::*", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	a := nodes[0]
	b := nodes[1]

	if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
		t.Error("Node not 'a'")
	}

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestDescendentOrSelf(t *testing.T) {
	xml := `<root><a><b/></a></root>`
	nodes := execNodes(t, "/root/descendant-or-self::*", xml)

	if len(nodes) != 3 {
		t.Error("Size is not 3")
	}

	root := nodes[0]
	a := nodes[1]
	b := nodes[2]

	if (store.Cursor)(root).Node().(node.Element).Local() != "root" {
		t.Error("Node not 'root'")
	}

	if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
		t.Error("Node not 'a'")
	}

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestFollowing(t *testing.T) {
	xml := `
<root>
	<a>
		<b/>
		<c/>
	</a>
	<d>
		<e/>
	</d>
</root>`
	nodes := execNodes(t, "/root/a/b/following::*", xml)

	if len(nodes) != 3 {
		t.Error("Size is not 3")
	}

	c := nodes[0]
	d := nodes[1]
	e := nodes[2]

	if (store.Cursor)(c).Node().(node.Element).Local() != "c" {
		t.Error("Node not 'c'")
	}

	if (store.Cursor)(d).Node().(node.Element).Local() != "d" {
		t.Error("Node not 'd'")
	}

	if (store.Cursor)(e).Node().(node.Element).Local() != "e" {
		t.Error("Node not 'e'")
	}
}

func TestFollowingSibling(t *testing.T) {
	xml := `
<root>
	<a>
		<b/>
		<c/>
		<d/>
		<e/>
	</a>
	<f/>
</root>`
	nodes := execNodes(t, "/root/a/c/following-sibling::*", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	d := nodes[0]
	e := nodes[1]

	if (store.Cursor)(d).Node().(node.Element).Local() != "d" {
		t.Error("Node not 'd'")
	}

	if (store.Cursor)(e).Node().(node.Element).Local() != "e" {
		t.Error("Node not 'e'")
	}
}

func TestParent(t *testing.T) {
	xml := `
<root>
	<a>
		<b/>
	</a>
	<b/>
</root>`

	execs := []string{"/root/a/b/parent::*", "/root/a/b/.."}

	for _, i := range execs {
		nodes := execNodes(t, i, xml)

		if len(nodes) != 1 {
			t.Error("Size is not 1")
		}

		a := nodes[0]

		if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
			t.Error("Node not 'a'")
		}
	}
}

func TestPreceding(t *testing.T) {
	xml := `
<root>
	<a>
		<b/>
		<c/>
	</a>
	<d>
		<e/>
	</d>
</root>`
	nodes := execNodes(t, "/root/d/e/preceding::*", xml)

	if len(nodes) != 3 {
		t.Error("Size is not 3")
	}

	c := nodes[0]
	b := nodes[1]
	a := nodes[2]

	if (store.Cursor)(c).Node().(node.Element).Local() != "c" {
		t.Error("Node not 'c'")
	}

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}

	if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
		t.Error("Node not 'a'")
	}
}

func TestPrecedingSibling(t *testing.T) {
	xml := `
<root>
	<f/>
	<a>
		<b/>
		<c/>
		<d/>
		<e/>
	</a>
</root>`
	nodes := execNodes(t, "/root/a/d/preceding-sibling::*", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	c := nodes[0]
	b := nodes[1]

	if (store.Cursor)(c).Node().(node.Element).Local() != "c" {
		t.Error("Node not 'c'")
	}

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestAbbreviatedAbsoluteLocation(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<a>b</a>
</root>`
	nodes := execNodes(t, "//a", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	a := nodes[0]
	b := nodes[1]

	if getCursorString(a) != "a" {
		t.Error("Node not 'a'")
	}

	if getCursorString(b) != "b" {
		t.Error("Node not 'b'")
	}
}

func TestAbbreviatedRelativeLocation(t *testing.T) {
	xml := `
<root>
	<foo>
		<a>a</a>
		<a>b</a>
		<b>z<a>c</a>z</b>
	</foo>
	<bar>
		<a>d</a>
	</bar>
	<a>e</a>
</root>`
	nodes := execNodes(t, "/root/foo//a", xml)

	if len(nodes) != 3 {
		t.Error("Size is not 3")
	}

	a := nodes[0]
	b := nodes[1]
	c := nodes[2]

	if getCursorString(a) != "a" {
		t.Error("Node not 'a'")
	}

	if getCursorString(b) != "b" {
		t.Error("Node not 'b'")
	}

	if getCursorString(c) != "c" {
		t.Error("Node not 'c'")
	}
}

func TestNamespaces(t *testing.T) {
	xml := `
<root root:xmlns="http://root">
	<a a:xmlns="http://a"/>
	<b xmlns="http://b"/>
</root>`
	nodes := execNodes(t, "/root/a/namespace::*", xml)

	if len(nodes) != 3 {
		t.Error("Size is not 3")
	}

	xmlns := nodes[0]
	a := nodes[1]
	root := nodes[2]

	if getCursorString(xmlns) != "http://www.w3.org/XML/1998/namespace" {
		t.Error("Node not 'http://www.w3.org/XML/1998/namespace'")
	}

	if getCursorString(a) != "http://a" {
		t.Error("Node not 'http://a'")
	}

	if getCursorString(root) != "http://root" {
		t.Error("Node not 'http://root'")
	}
}

func TestNamespaceOverride(t *testing.T) {
	xml := `
<root root:xmlns="http://root">
	<a root:xmlns="http://a"/>
	<b root:xmlns="http://b"/>
</root>`
	nodes := execNodes(t, "/root/a/namespace::*", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	xmlns := nodes[0]
	a := nodes[1]

	if getCursorString(xmlns) != "http://www.w3.org/XML/1998/namespace" {
		t.Error("Node not 'http://www.w3.org/XML/1998/namespace'")
	}

	if getCursorString(a) != "http://a" {
		t.Error("Node not 'http://a'")
	}
}

func TestDefaultNamespace(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
</root>`
	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["foo"] = "http://root"
		c.NamespaceDecls["bar"] = "http://a"
	}

	nodes := execNodes(t, "/foo:root/bar:a", xml, namespaces)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	a := nodes[0]

	if (store.Cursor)(a).Node().(node.Element).Local() != "a" {
		t.Error("Node not 'a'")
	}
}

func TestDefaultNamespaceOverrides(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
</root>`
	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["foo"] = "http://root"
		c.NamespaceDecls["bar"] = "http://a"
	}

	nodes := execNodes(t, "/foo:root/bar:a/namespace::*", xml, namespaces)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	xmlns := nodes[0]
	a := nodes[1]

	if getCursorString(xmlns) != "http://www.w3.org/XML/1998/namespace" {
		t.Error("Node not 'http://www.w3.org/XML/1998/namespace'")
	}

	if getCursorString(a) != "http://a" {
		t.Error("Node not 'http://a'")
	}
}

func TestNamespaceSelect(t *testing.T) {
	xml := `<root xmlns="http://root"/>`
	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["foo"] = "http://root"
	}

	nodes := execNodes(t, "/foo:root/namespace::foo", xml, namespaces)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	root := nodes[0]

	if getCursorString(root) != "http://root" {
		t.Error("Node not 'http://root'")
	}
}

func TestNamespaceAnyLocal(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
	<b xmlns="http://b"/>
</root>`
	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["b"] = "http://b"
	}

	nodes := execNodes(t, "//b:*", xml, namespaces)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	b := nodes[0]

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestLocalAnyNamespace(t *testing.T) {
	xml := `
<root>
	<a xmlns="http://a"/>
	<a xmlns="http://b"/>
</root>`
	nodes := execNodes(t, "//*:a", xml)

	if len(nodes) != 2 {
		t.Error("Size is not 2")
	}

	a := nodes[0]
	b := nodes[1]

	if (store.Cursor)(a).Node().(node.Element).Space() != "http://a" {
		t.Error("Node not 'http://a'")
	}

	if (store.Cursor)(b).Node().(node.Element).Space() != "http://b" {
		t.Error("Node not 'http://b'")
	}
}

func TestFunctionLast(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<a>b</a>
</root>`

	execNodesToString(t, "/root/a[last()]", xml, "b")
}

func TestFunctionPosition(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<a>b</a>
	<a>c</a>
</root>`

	execNodesToString(t, "/root/a[position() = 2]", xml, "b")
}

func TestFunctionCount(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<a>b</a>
	<a>c</a>
</root>`

	execNodesToString(t, "count(/root/a)", xml, "3")
}

func TestFunctionLocalName(t *testing.T) {
	xml := `
<root>
</root>`

	execNodesToString(t, "/root/local-name()", xml, "root")
	execNodesToString(t, "local-name(/root)", xml, "root")
}

func TestFunctionNamespaceUri(t *testing.T) {
	xml := `
<root xmlns="http://foo">
</root>`

	execNodesToString(t, "/*/namespace-uri()", xml, "http://foo")
	execNodesToString(t, "namespace-uri(/*)", xml, "http://foo")
}

func TestFunctionName(t *testing.T) {
	xml := `
<root>
</root>`

	execNodesToString(t, "/*/name()", xml, "root")
	execNodesToString(t, "name(/*)", xml, "root")

	xml = `
<root xmlns="http://foo">
</root>`

	execNodesToString(t, "/*/name()", xml, "{http://foo}root")
	execNodesToString(t, "name(/*)", xml, "{http://foo}root")
}

func TestFunctionString(t *testing.T) {
	xml := `<root>1</root>`

	execNodesToString(t, "string(/root)", xml, "1")
	execNodesToString(t, "/root/string()", xml, "1")
}

func TestFunctionConcat(t *testing.T) {
	execNodesToString(t, "concat('foo', 'bar')", ``, "foobar")
}

func TestFunctionStartsWith(t *testing.T) {
	execNodesToString(t, "starts-with('abcd', 'ab')", ``, "true")
	execNodesToString(t, "starts-with('abcd', 'b')", ``, "false")
}

func TestFunctionContains(t *testing.T) {
	execNodesToString(t, "contains('abcd', 'bc')", ``, "true")
	execNodesToString(t, "contains('abcd', 'z')", ``, "false")
}

func TestFunctionSubstringBefore(t *testing.T) {
	execNodesToString(t, `substring-before("1999/04/01","/")`, ``, "1999")
	execNodesToString(t, `substring-before("1999/04/01","2")`, ``, "")
}

func TestFunctionSubstringAfter(t *testing.T) {
	execNodesToString(t, `substring-after("1999/04/01","/")`, ``, "04/01")
	execNodesToString(t, `substring-after("1999/04/01","19")`, ``, "99/04/01")
	execNodesToString(t, `substring-after("1999/04/01","a")`, ``, "")
}

func TestFunctionSubstring(t *testing.T) {
	execNodesToString(t, `substring("12345", 2, 3)`, ``, "234")
	execNodesToString(t, `substring("12345", 2)`, ``, "2345")
	execNodesToString(t, `substring('abcd', -2, 5)`, ``, "ab")
	execNodesToString(t, `substring('abcd', 0)`, ``, "abcd")
	execNodesToString(t, `substring('abcd', 1, 4)`, ``, "abcd")
	execNodesToString(t, `substring("12345", 1.5, 2.6)`, ``, "234")
	execNodesToString(t, `substring("12345", 0 div 0, 3)`, ``, "")
	execNodesToString(t, `substring("12345", 1, 0 div 0)`, ``, "")
	execNodesToString(t, `substring("12345", -42, 1 div 0)`, ``, "12345")
	execNodesToString(t, `substring("12345", -1 div 0, 1 div 0)`, ``, "")
}

func TestFunctionStringLength(t *testing.T) {
	xml := `<root>1234</root>`

	execNodesToString(t, "string-length(/root)", xml, "4")
	execNodesToString(t, "/root/string-length()", xml, "4")
}

func TestFunctionNormalizeSpace(t *testing.T) {
	xml := `<root>  1234   </root>`

	execNodesToString(t, "normalize-space(/root)", xml, "1234")
	execNodesToString(t, "/root/normalize-space()", xml, "1234")
}

func TestFunctionTranslate(t *testing.T) {
	execNodesToString(t, `translate("bar","abc","ABC")`, ``, "BAr")
	execNodesToString(t, `translate("--aaa--","abc-","ABC")`, ``, "AAA")
}

func TestFunctionNot(t *testing.T) {
	exec(t, `not(1)`, ``, Bool(false))
	exec(t, `not(0)`, ``, Bool(true))
}

func TestFunctionTrue(t *testing.T) {
	exec(t, `true()`, ``, Bool(true))
}

func TestFunctionFalse(t *testing.T) {
	exec(t, `false()`, ``, Bool(false))
}

func TestFunctionLang(t *testing.T) {
	xml := `
<p1>
	<p xml:lang="en">I went up a floor.</p>
	<p xml:lang="en-GB">I took the lift.</p>
	<p xml:lang="en-US">I rode the elevator.</p>
</p1>`

	execNodesToString(t, `count(//p[lang('en')])`, xml, "3")
	execNodesToString(t, `count(//text()[lang('en-GB')])`, xml, "1")
	execNodesToString(t, `count(//p[lang('en-US')])`, xml, "1")
	execNodesToString(t, `count(//p[lang('de')])`, xml, "0")
	execNodesToString(t, `count(/p1[lang('en')])`, xml, "0")
}

func TestFunctionNumber(t *testing.T) {
	xml := `<root>1234</root>`

	exec(t, "number(/root)", xml, Number(1234))
	exec(t, "/root/number()", xml, Number(1234))
}

func TestFunctionSum(t *testing.T) {
	xml := `
<root>
	<a>1</a>
	<a>2</a>
	<a>3</a>
</root>
`

	exec(t, "sum(/root/a)", xml, Number(6))
}

func TestFunctionFloor(t *testing.T) {
	exec(t, "floor(2.2)", ``, Number(2))
}

func TestFunctionCeiling(t *testing.T) {
	exec(t, "ceiling(2.2)", ``, Number(3))
}

func TestFunctionRound(t *testing.T) {
	exec(t, `round(-1.5)`, ``, Number(-2))
	exec(t, `round(1.5)`, ``, Number(2))
	exec(t, `round(0)`, ``, Number(0))
}

func TestCustomFunction(t *testing.T) {
	xml := `
<root>
	<a>5</a>
	<b>2.5</b>
	<c>6</c>
</root>
`
	variables := func(c *ContextSettings) {
		c.NamespaceDecls["foo"] = "http://root"
		c.FunctionLibrary[XmlName{"http://root", "bar"}] = func(context Context, args ...Result) (Result, error) {
			return Number(2.5), nil
		}
	}

	nodes := execNodes(t, "//*[. = foo:bar()]", xml, variables)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	b := nodes[0]

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestVariableReference(t *testing.T) {
	xml := `
<root>
	<a>5</a>
	<b>2.5</b>
	<c>6</c>
</root>
`
	variables := func(c *ContextSettings) {
		c.NamespaceDecls["foo"] = "http://root"
		c.Variables[XmlName{"http://root", "bar"}] = Number(2.5)
	}

	nodes := execNodes(t, "//*[. = $foo:bar]", xml, variables)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	b := nodes[0]

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestNewContext(t *testing.T) {
	xml := `<root><a>5</a><b>2.5</b><c>6</c></root>`

	xpath, err := grammar.Build("a[1]")

	if err != nil {
		t.Error(err)
		return
	}

	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, err := store.CreateInMemory(parser)

	if err != nil {
		t.Error(err)
		return
	}

	result, err := Exec(cursor, &xpath, func(c *ContextSettings) {
		c.Context = cursor.Children()[0]
	})

	if err != nil {
		t.Error(err)
	}

	if result.String() != "5" {
		t.Errorf("Result != '5'")
	}
}
