package exec

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

func query(t *testing.T, expr string, parser parser.Parser, settings ...ContextApply) Result {
	xpath := grammar.MustBuild(expr)
	cursor, err := store.CreateInMemory(parser)

	if err != nil {
		t.Error(err)
		return NodeSet{}
	}

	result, err := Exec(cursor, &xpath, settings...)

	if err != nil {
		t.Error(err)
	}

	return result
}

func queryXml(t *testing.T, expr, xml string, settings ...ContextApply) Result {
	parser := parser.ReadXml(bytes.NewBufferString(xml))

	return query(t, expr, parser, settings...)
}

func execXml(t *testing.T, expr, xml string, expected Result, settings ...ContextApply) {
	result := queryXml(t, expr, xml, settings...)

	if result != expected {
		t.Errorf("Result != '%s'. Received '%s'", expected, result)
	}
}

func execXmlNodesToString(t *testing.T, expr, xml string, expected string, settings ...ContextApply) {
	resultString := queryXml(t, expr, xml, settings...).String()

	if resultString != expected {
		t.Errorf("Result != '%s'. Received '%s'", expected, resultString)
	}
}

func execXmlNodes(t *testing.T, expr, xml string, settings ...ContextApply) NodeSet {
	return queryXml(t, expr, xml, settings...).(NodeSet)
}

func queryJson(t *testing.T, expr, json string, settings ...ContextApply) Result {
	parser := parser.ReadJson(bytes.NewBufferString(json))

	return query(t, expr, parser, settings...)
}

func execJsonNodes(t *testing.T, expr, json string, settings ...ContextApply) NodeSet {
	return queryJson(t, expr, json, settings...).(NodeSet)
}

func queryHtml(t *testing.T, expr, html string, settings ...ContextApply) Result {
	parser, err := parser.ReadHtml(bytes.NewBufferString(html))

	if err != nil {
		t.Error(err)
		return nil
	}

	return query(t, expr, parser, settings...)
}

func execHtmlNodes(t *testing.T, expr, html string, settings ...ContextApply) NodeSet {
	return queryHtml(t, expr, html, settings...).(NodeSet)
}

/*
func printTree(cursor store.Cursor, depth int) {
	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}

	switch t := cursor.Node().(type) {
	case node.Element:
		fmt.Println("<" + t.Local() + ">")
	case node.CharData:
		fmt.Println(t.CharDataValue())
	}

	for _, i := range cursor.Children() {
		printTree(i, depth+1)
	}

	if t, isEnd := cursor.Node().(node.Element); isEnd {
		for i := 0; i < depth; i++ {
			fmt.Print("  ")
		}

		fmt.Println("</" + t.Local() + ">")
	}
}
*/

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
	execXml(t, "1.2+2.3", `<root/>`, Number(3.5))
}

func TestSubtract(t *testing.T) {
	execXml(t, "5-3", `<root/>`, Number(2))
}

func TestMultiply(t *testing.T) {
	execXml(t, "3*4", `<root/>`, Number(12))
}

func TestDivide(t *testing.T) {
	execXml(t, "15 div 3", `<root/>`, Number(5))
	execXmlNodesToString(t, "0 div 0", `<root/>`, Number(math.NaN()).String())
	execXml(t, "1 div 0", `<root/>`, Number(math.Inf(1)))
	execXml(t, "-1 div 0", `<root/>`, Number(math.Inf(-1)))
}

func TestMod(t *testing.T) {
	execXml(t, "4 mod 3", `<root/>`, Number(1))
	execXmlNodesToString(t, "4 mod 0", `<root/>`, Number(math.NaN()).String())
}

func TestString(t *testing.T) {
	execXml(t, "'foo'", `<root/>`, String("foo"))
}

func TestNegate(t *testing.T) {
	execXml(t, "8", `<root/>`, Number(8))
	execXml(t, "-8", `<root/>`, Number(-8))
	execXml(t, "--8", `<root/>`, Number(8))
	execXml(t, "---8", `<root/>`, Number(-8))
}

func TestEquality(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<b>b</b>
	<one>1</one>
</root>
`
	execXml(t, "/root/a = /root/b", xml, Bool(false))
	execXml(t, "/root/a = /root/a", xml, Bool(true))

	execXml(t, "/root/one = 1", xml, Bool(true))
	execXml(t, "1 = /root/one", xml, Bool(true))
	execXml(t, "2 = /root/one", xml, Bool(false))
	execXml(t, "/root/one = 2", xml, Bool(false))

	execXml(t, "/root/a = 'a'", xml, Bool(true))
	execXml(t, "'a' = /root/a", xml, Bool(true))
	execXml(t, "'b' = /root/a", xml, Bool(false))
	execXml(t, "/root/a = 'b'", xml, Bool(false))

	execXml(t, "1 = 1", xml, Bool(true))
	execXml(t, "1 = 2", xml, Bool(false))

	execXml(t, "1 = '1'", xml, Bool(true))
	execXml(t, "1 = '2'", xml, Bool(false))

	execXml(t, "'1' = '1'", xml, Bool(true))
	execXml(t, "'1' = '2'", xml, Bool(false))

	execXml(t, "/root/a = true()", xml, Bool(true))
	execXml(t, "true() = /root/a", xml, Bool(true))
	execXml(t, "true() = 1", xml, Bool(true))
	execXml(t, "true() = 0", xml, Bool(false))
}

func TestNotEqual(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<b>b</b>
	<one>1</one>
</root>
`
	execXml(t, "/root/a != /root/b", xml, Bool(true))
	execXml(t, "/root/a != /root/a", xml, Bool(false))

	execXml(t, "/root/one != 1", xml, Bool(false))
	execXml(t, "1 != /root/one", xml, Bool(false))
	execXml(t, "2 != /root/one", xml, Bool(true))
	execXml(t, "/root/one != 2", xml, Bool(true))

	execXml(t, "/root/a != 'a'", xml, Bool(false))
	execXml(t, "'a' != /root/a", xml, Bool(false))
	execXml(t, "'b' != /root/a", xml, Bool(true))
	execXml(t, "/root/a != 'b'", xml, Bool(true))

	execXml(t, "1 != 1", xml, Bool(false))
	execXml(t, "1 != 2", xml, Bool(true))

	execXml(t, "1 != '1'", xml, Bool(false))
	execXml(t, "1 != '2'", xml, Bool(true))

	execXml(t, "'1' != '1'", xml, Bool(false))
	execXml(t, "'1' != '2'", xml, Bool(true))

	execXml(t, "/root/a != true()", xml, Bool(false))
	execXml(t, "true() != /root/a", xml, Bool(false))
	execXml(t, "true() != 1", xml, Bool(false))
	execXml(t, "true() != 0", xml, Bool(true))
}

func TestLessThan(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	execXml(t, "/root/one < /root/two", xml, Bool(true))
	execXml(t, "1 < /root/two", xml, Bool(true))
	execXml(t, "/root/two < /root/one", xml, Bool(false))

	execXml(t, "/root/two < 1", xml, Bool(false))
	execXml(t, "3 < /root/two", xml, Bool(false))
	execXml(t, "/root/one < 2", xml, Bool(true))

	execXml(t, "'1' < /root/two", xml, Bool(true))
	execXml(t, "/root/one < '2'", xml, Bool(true))
	execXml(t, "'3' < /root/two", xml, Bool(false))
	execXml(t, "/root/two < '1'", xml, Bool(false))

	execXml(t, "'1' < '2'", xml, Bool(true))
}

func TestLessThanOrEqual(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	execXml(t, "/root/one <= /root/two", xml, Bool(true))
	execXml(t, "/root/two <= /root/two", xml, Bool(true))
	execXml(t, "1 <= /root/two", xml, Bool(true))
	execXml(t, "2 <= /root/two", xml, Bool(true))
	execXml(t, "/root/two <= /root/one", xml, Bool(false))
	execXml(t, "/root/two <= /root/two", xml, Bool(true))

	execXml(t, "/root/two <= 1", xml, Bool(false))
	execXml(t, "3 <= /root/two", xml, Bool(false))
	execXml(t, "/root/one <= 2", xml, Bool(true))
	execXml(t, "/root/two <= 2", xml, Bool(true))

	execXml(t, "'1' <= /root/two", xml, Bool(true))
	execXml(t, "'2' <= /root/two", xml, Bool(true))
	execXml(t, "/root/one <= '2'", xml, Bool(true))
	execXml(t, "/root/two <= '2'", xml, Bool(true))
	execXml(t, "'3' <= /root/two", xml, Bool(false))
	execXml(t, "/root/two <= '1'", xml, Bool(false))

	execXml(t, "'1' <= '2'", xml, Bool(true))
	execXml(t, "'2' <= '2'", xml, Bool(true))
}

func TestGreaterThan(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	execXml(t, "/root/one > /root/two", xml, Bool(false))
	execXml(t, "1 > /root/two", xml, Bool(false))
	execXml(t, "/root/two > /root/one", xml, Bool(true))

	execXml(t, "/root/two > 1", xml, Bool(true))
	execXml(t, "3 > /root/two", xml, Bool(true))
	execXml(t, "/root/one > 2", xml, Bool(false))

	execXml(t, "'1' > /root/two", xml, Bool(false))
	execXml(t, "/root/one > '2'", xml, Bool(false))
	execXml(t, "'3' > /root/two", xml, Bool(true))
	execXml(t, "/root/two > '1'", xml, Bool(true))

	execXml(t, "'1' > '2'", xml, Bool(false))
}

func TestGreaterThanOrEqual(t *testing.T) {
	xml := `
<root>
	<one>1</one>
	<two>2</two>
</root>
`
	execXml(t, "/root/one >= /root/two", xml, Bool(false))
	execXml(t, "1 >= /root/two", xml, Bool(false))
	execXml(t, "/root/two >= /root/one", xml, Bool(true))
	execXml(t, "/root/two >= /root/two", xml, Bool(true))

	execXml(t, "/root/two >= 1", xml, Bool(true))
	execXml(t, "/root/two >= 2", xml, Bool(true))
	execXml(t, "3 >= /root/two", xml, Bool(true))
	execXml(t, "2 >= /root/two", xml, Bool(true))
	execXml(t, "/root/one >= 2", xml, Bool(false))

	execXml(t, "'1' >= /root/two", xml, Bool(false))
	execXml(t, "/root/one >= '2'", xml, Bool(false))
	execXml(t, "'3' >= /root/two", xml, Bool(true))
	execXml(t, "'2' >= /root/two", xml, Bool(true))
	execXml(t, "/root/two >= '1'", xml, Bool(true))
	execXml(t, "/root/two >= '2'", xml, Bool(true))

	execXml(t, "'1' >= '2'", xml, Bool(false))
	execXml(t, "'2' >= '2'", xml, Bool(true))
}

func TestOr(t *testing.T) {
	execXml(t, "1 or 0", `<root/>`, Bool(true))
	execXml(t, "0 or 0", `<root/>`, Bool(false))
}

func TestAnd(t *testing.T) {
	execXml(t, "1 and 0", `<root/>`, Bool(false))
	execXml(t, "1 and 1", `<root/>`, Bool(true))
}

func TestAbsoluteLocationPathOnly(t *testing.T) {
	execXmlNodesToString(t, "/", `b <root>a root node</root> c`, "b a root node c")
}

func TestAbsoluteLocationPathWithRelative(t *testing.T) {
	xml := `
a root node
<Node>node value</Node>
other text
`
	execXmlNodesToString(t, "/ Node", xml, "node value")
}

func TestRelativeLocationPath(t *testing.T) {
	xml := `
text
<Root>text2
<node>a</node>
<node>b</node>
<attribute>c</attribute>
text3
</Root>
text4
`
	execXmlNodesToString(t, "/Root/node", xml, "a")
	execXmlNodesToString(t, "/Root/attribute", xml, "c")
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
	execXmlNodesToString(t, "/Root/Node[2]", xml, "b")
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
	nodes := execXmlNodes(t, "/ Root/ Node [ 1 ] | /Root/Node[2]", xml)

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
	nodes := execXmlNodes(t, "/Root/Node[1] | /Root/Node[1]", xml)

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
	execXmlNodesToString(t, "/root/node ( ) ", xml, "foo")

	xml = `
<!--some comment-->
<comment>node</comment>
`
	execXmlNodesToString(t, "/comment ( ) ", xml, "some comment")
	execXmlNodesToString(t, "/comment", xml, "node")

	xml = `
<?foo bar?>
<processing-instruction>proc</processing-instruction>
<?eggs spam?>
`
	execXmlNodesToString(t, "/processing-instruction ( ) ", xml, "bar")
	execXmlNodesToString(t, "/processing-instruction ( 'eggs' ) ", xml, "spam")
	execXmlNodesToString(t, "/processing-instruction", xml, "proc")

	xml = `some text<text>other text</text>`
	execXmlNodesToString(t, "/text ( ) ", xml, "some text")
	execXmlNodesToString(t, "/text", xml, "other text")
}

func TestAnyElement(t *testing.T) {
	xml := `
<root>root text<data>data text</data></root>
`
	execXmlNodesToString(t, "/root/*", xml, "data text")
}

func TestChild(t *testing.T) {
	xml := `
a root node
<Node>node value</Node>
other text
`
	execXmlNodesToString(t, "/child::Node", xml, "node value")
}

func TestAnyAttr(t *testing.T) {
	xml := `
<root foo="bar" eggs="ham"></root>
`
	nodes := execXmlNodes(t, "/root/attribute::*", xml)

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
	nodes := execXmlNodes(t, "/root/@eggs", xml)

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
	nodes := execXmlNodes(t, "/root/a/b/ancestor::*", xml)

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
	nodes := execXmlNodes(t, "/root/a/b/ancestor-or-self::*", xml)

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
	nodes := execXmlNodes(t, "/root/descendant::*", xml)

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
	nodes := execXmlNodes(t, "/root/descendant-or-self::*", xml)

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
	nodes := execXmlNodes(t, "/root/a/b/following::*", xml)

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
	nodes := execXmlNodes(t, "/root/a/c/following-sibling::*", xml)

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
		nodes := execXmlNodes(t, i, xml)

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
	nodes := execXmlNodes(t, "/root/d/e/preceding::*", xml)

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
	nodes := execXmlNodes(t, "/root/a/d/preceding-sibling::*", xml)

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
	nodes := execXmlNodes(t, "//a", xml)

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
	nodes := execXmlNodes(t, "/root/foo//a", xml)

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
	nodes := execXmlNodes(t, "/root/a/namespace::*", xml)

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
	nodes := execXmlNodes(t, "/root/a/namespace::*", xml)

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

	nodes := execXmlNodes(t, "/foo:root/bar:a", xml, namespaces)

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

	nodes := execXmlNodes(t, "/foo:root/bar:a/namespace::*", xml, namespaces)

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

	nodes := execXmlNodes(t, "/foo:root/namespace::foo", xml, namespaces)

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

	nodes := execXmlNodes(t, "//b:*", xml, namespaces)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	b := nodes[0]

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestNamespaceAnyLocalConflict(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
	<b xmlns="http://b"/>
	<c xmlns="http://c">c</c>
	<d xmlns="http://c">d</d>
</root>`

	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["attribute"] = "http://b"
	}

	nodes := execXmlNodes(t, "//attribute:*", xml, namespaces)

	if len(nodes) != 1 {
		t.Error("Size is not 1")
	}

	b := nodes[0]

	if (store.Cursor)(b).Node().(node.Element).Local() != "b" {
		t.Error("Node not 'b'")
	}
}

func TestNameTestQNameNamespaceWithLocalReservedNameConflictNamespace(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
	<b xmlns="http://b"/>
	<c xmlns="http://c">c</c>
	<d xmlns="http://c">d</d>
</root>`

	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["descendant"] = "http://c"
	}

	execXmlNodesToString(t, "//descendant:c", xml, "c", namespaces)
}

func TestNameTestQNameNamespaceWithLocalReservedNameConflictLocal(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
	<b xmlns="http://b"/>
	<descendant xmlns="http://c">c</descendant>
	<descendant xmlns="http://c">d</descendant>
</root>`

	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["a"] = "http://c"
	}

	execXmlNodesToString(t, "//a:descendant", xml, "c", namespaces)
}

func TestNameTestQNameNamespaceWithLocalReservedNameConflictBoth(t *testing.T) {
	xml := `
<root xmlns="http://root">
	<a xmlns="http://a"/>
	<b xmlns="http://b"/>
	<descendant xmlns="http://c">c</descendant>
	<descendant xmlns="http://c">d</descendant>
</root>`

	namespaces := func(c *ContextSettings) {
		c.NamespaceDecls["descendant"] = "http://c"
	}

	execXmlNodesToString(t, "//descendant:descendant", xml, "c", namespaces)
}

func TestLocalAnyNamespace(t *testing.T) {
	xml := `
<root>
	<a xmlns="http://a"/>
	<a xmlns="http://b"/>
</root>`
	nodes := execXmlNodes(t, "//*:a", xml)

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

func TestLocalAxisConflict(t *testing.T) {
	xml := `
<root>
	<attribute xmlns="http://a"/>
	<attribute xmlns="http://b"/>
</root>`
	nodes := execXmlNodes(t, "//*: attribute ", xml)

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

	execXmlNodesToString(t, "/root/a[last()]", xml, "b")
}

func TestFunctionPosition(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<a>b</a>
	<a>c</a>
</root>`

	execXmlNodesToString(t, "/root/a[position() = 2]", xml, "b")
}

func TestFunctionCount(t *testing.T) {
	xml := `
<root>
	<a>a</a>
	<a>b</a>
	<a>c</a>
</root>`

	execXmlNodesToString(t, "count(/root/a)", xml, "3")
}

func TestFunctionLocalName(t *testing.T) {
	xml := `
<root>
</root>`

	execXmlNodesToString(t, "/root/local-name()", xml, "root")
	execXmlNodesToString(t, "local-name(/root)", xml, "root")
}

func TestFunctionNamespaceUri(t *testing.T) {
	xml := `
<root xmlns="http://foo">
</root>`

	execXmlNodesToString(t, "/*/namespace-uri()", xml, "http://foo")
	execXmlNodesToString(t, "namespace-uri(/*)", xml, "http://foo")
}

func TestFunctionName(t *testing.T) {
	xml := `
<root>
</root>`

	execXmlNodesToString(t, "/*/name()", xml, "root")
	execXmlNodesToString(t, "name(/*)", xml, "root")

	xml = `
<root xmlns="http://foo">
</root>`

	execXmlNodesToString(t, "/*/name()", xml, "{http://foo}root")
	execXmlNodesToString(t, "name(/*)", xml, "{http://foo}root")
}

func TestFunctionString(t *testing.T) {
	xml := `<root>1</root>`

	execXmlNodesToString(t, "string(/root)", xml, "1")
	execXmlNodesToString(t, "/root/string()", xml, "1")
}

func TestFunctionConcat(t *testing.T) {
	execXmlNodesToString(t, "concat('foo', 'bar')", ``, "foobar")
}

func TestFunctionStartsWith(t *testing.T) {
	execXmlNodesToString(t, "starts-with('abcd', 'ab')", ``, "true")
	execXmlNodesToString(t, "starts-with('abcd', 'b')", ``, "false")
}

func TestFunctionContains(t *testing.T) {
	execXmlNodesToString(t, "contains('abcd', 'bc')", ``, "true")
	execXmlNodesToString(t, "contains('abcd', 'z')", ``, "false")
}

func TestFunctionSubstringBefore(t *testing.T) {
	execXmlNodesToString(t, `substring-before("1999/04/01","/")`, ``, "1999")
	execXmlNodesToString(t, `substring-before("1999/04/01","2")`, ``, "")
}

func TestFunctionSubstringAfter(t *testing.T) {
	execXmlNodesToString(t, `substring-after("1999/04/01","/")`, ``, "04/01")
	execXmlNodesToString(t, `substring-after("1999/04/01","19")`, ``, "99/04/01")
	execXmlNodesToString(t, `substring-after("1999/04/01","a")`, ``, "")
}

func TestFunctionSubstring(t *testing.T) {
	execXmlNodesToString(t, `substring("12345", 2, 3)`, ``, "234")
	execXmlNodesToString(t, `substring("12345", 2)`, ``, "2345")
	execXmlNodesToString(t, `substring('abcd', -2, 5)`, ``, "ab")
	execXmlNodesToString(t, `substring('abcd', 0)`, ``, "abcd")
	execXmlNodesToString(t, `substring('abcd', 1, 4)`, ``, "abcd")
	execXmlNodesToString(t, `substring("12345", 1.5, 2.6)`, ``, "234")
	execXmlNodesToString(t, `substring("12345", 0 div 0, 3)`, ``, "")
	execXmlNodesToString(t, `substring("12345", 1, 0 div 0)`, ``, "")
	execXmlNodesToString(t, `substring("12345", -42, 1 div 0)`, ``, "12345")
	execXmlNodesToString(t, `substring("12345", -1 div 0, 1 div 0)`, ``, "")
}

func TestFunctionStringLength(t *testing.T) {
	xml := `<root>1234</root>`

	execXmlNodesToString(t, "string-length(/root)", xml, "4")
	execXmlNodesToString(t, "/root/string-length()", xml, "4")
}

func TestFunctionNormalizeSpace(t *testing.T) {
	xml := `<root>  1234   </root>`

	execXmlNodesToString(t, "normalize-space(/root)", xml, "1234")
	execXmlNodesToString(t, "/root/normalize-space()", xml, "1234")
}

func TestFunctionTranslate(t *testing.T) {
	execXmlNodesToString(t, `translate("bar","abc","ABC")`, ``, "BAr")
	execXmlNodesToString(t, `translate("--aaa--","abc-","ABC")`, ``, "AAA")
}

func TestFunctionNot(t *testing.T) {
	execXml(t, `not(1)`, ``, Bool(false))
	execXml(t, `not(0)`, ``, Bool(true))
}

func TestFunctionTrue(t *testing.T) {
	execXml(t, `true()`, ``, Bool(true))
}

func TestFunctionFalse(t *testing.T) {
	execXml(t, `false()`, ``, Bool(false))
}

func TestFunctionLang(t *testing.T) {
	xml := `
<p1>
	<p xml:lang="en">I went up a floor.</p>
	<p xml:lang="en-GB">I took the lift.</p>
	<p xml:lang="en-US">I rode the elevator.</p>
</p1>`

	execXmlNodesToString(t, `count(//p[lang('en')])`, xml, "3")
	execXmlNodesToString(t, `count(//text()[lang('en-GB')])`, xml, "1")
	execXmlNodesToString(t, `count(//p[lang('en-US')])`, xml, "1")
	execXmlNodesToString(t, `count(//p[lang('de')])`, xml, "0")
	execXmlNodesToString(t, `count(/p1[lang('en')])`, xml, "0")
}

func TestFunctionNumber(t *testing.T) {
	xml := `<root>1234</root>`

	execXml(t, "number(/root)", xml, Number(1234))
	execXml(t, "/root/number()", xml, Number(1234))
}

func TestFunctionSum(t *testing.T) {
	xml := `
<root>
	<a>1</a>
	<a>2</a>
	<a>3</a>
</root>
`

	execXml(t, "sum(/root/a)", xml, Number(6))
}

func TestFunctionFloor(t *testing.T) {
	execXml(t, "floor(2.2)", ``, Number(2))
}

func TestFunctionCeiling(t *testing.T) {
	execXml(t, "ceiling(2.2)", ``, Number(3))
}

func TestFunctionRound(t *testing.T) {
	execXml(t, `round(-1.5)`, ``, Number(-2))
	execXml(t, `round(1.5)`, ``, Number(2))
	execXml(t, `round(0)`, ``, Number(0))
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

	nodes := execXmlNodes(t, "//*[. = foo:bar()]", xml, variables)

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

	nodes := execXmlNodes(t, "//*[. = $foo:bar]", xml, variables)

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

	result, err := Exec(cursor.Children()[0], &xpath)

	if err != nil {
		t.Error(err)
	}

	if result.String() != "5" {
		t.Errorf("Result != '5'")
	}
}

func TestJsonNestedArray(t *testing.T) {
	json := `
{
	"a": [ 0, ["b", "c", {"d": 2.71828}]],
	"b": {
		"c": 3.14,
		"d": [{"e": "f"}, "g"]
	},
	"nil": null
}
`

	value := execJsonNodes(t, "/#obj/a/#arr/text()[. = '0']", json)

	if c := value[0].Node().(node.CharData); c.CharDataValue() != "0" {
		t.Error("bad array value")
	}

	value = execJsonNodes(t, "/#obj/a/#arr/#arr/text()[. = 'b']", json)

	if c := value[0].Node().(node.CharData); c.CharDataValue() != "b" {
		t.Error("bad nested array value")
	}

	value = execJsonNodes(t, "/#obj/a/#arr/#arr/#obj/d[. = 2.71828]", json)

	if value.String() != "2.71828" || value[0].Node().(node.Element).Local() != "d" {
		t.Error("bad object-in-array value")
	}

	value = execJsonNodes(t, "/#obj/b/#obj/c", json)

	if value.String() != "3.14" || value[0].Node().(node.Element).Local() != "c" {
		t.Error("bad nested object value")
	}

	value = execJsonNodes(t, "/#obj/b/#obj/d/#arr/#obj/e", json)

	if value.String() != "f" || value[0].Node().(node.Element).Local() != "e" {
		t.Error("bad object-in-array-in-object value")
	}

	value = execJsonNodes(t, "/#obj/b/#obj/d/#arr/text()[. = 'g']", json)

	if c := value[0].Node().(node.CharData); c.CharDataValue() != "g" {
		t.Error("bad object-in-array-in-object value")
	}

	value = execJsonNodes(t, "/#obj/nil", json)

	if value.String() != "null" || value[0].Node().(node.Element).Local() != "nil" {
		t.Error("bad nil value")
	}
}

func TestJson(t *testing.T) {
	json := `
	{ "store": {
		"book": [
		  { "category": "reference",
			"author": "Nigel Rees",
			"title": "Sayings of the Century",
			"price": 8.95
		  },
		  { "category": "fiction",
			"author": "Evelyn Waugh",
			"title": "Sword of Honour",
			"price": 12.99
		  },
		  { "category": "fiction",
			"author": "Herman Melville",
			"title": "Moby Dick",
			"isbn": "0-553-21311-3",
			"price": 8.99
		  },
		  { "category": "fiction",
			"author": "J. R. R. Tolkien",
			"title": "The Lord of the Rings",
			"isbn": "0-395-19395-8",
			"price": 22.99
		  }
		],
		"bicycle": {
		  "color": "red",
		  "price": 19.95
		}
	  }
	}
`

	pricedItems := execJsonNodes(t, "//*[price]", json)

	if len(pricedItems) != 5 {
		t.Error("result size not 5")
	}

	for i := 0; i < 4; i++ {
		if pricedItems[i].Node().(node.Element).Local() != "#obj" {
			t.Error("name not '#obj'")
		}
	}

	if pricedItems[4].Parent().Node().(node.Element).Local() != "bicycle" {
		t.Error("name not 'bicycle'")
	}

	nodes := execJsonNodes(t, "/#obj/store/#obj/book/#arr/#obj/author", json)

	if len(nodes) != 4 {
		t.Error("result size not 4")
	}

	for _, i := range nodes {
		if i.Node().(node.Element).Local() != "author" {
			t.Error("name not 'author'")
		}
	}

	if getCursorString(nodes[0]) != "Nigel Rees" {
		t.Error("first node value incorrect")
	}

	if getCursorString(nodes[1]) != "Evelyn Waugh" {
		t.Error("second node value incorrect")
	}

	if getCursorString(nodes[2]) != "Herman Melville" {
		t.Error("third node value incorrect")
	}

	if getCursorString(nodes[3]) != "J. R. R. Tolkien" {
		t.Error("forth node value incorrect")
	}
}

func TestHtmlDocument(t *testing.T) {
	html := `
<!doctype html>
<html lang=en xmlns:svg="http://www.w3.org/2000/svg">
<head><meta charset="utf-8"><title>html title</title></head><body>
<br/>
<p>content</p>
<svg:svg height="110" xmlns="http://www.w3.org/2000/svg">
  <rect width="300" style="fill:rgb(0,0,255)" xlink:href="http://example.com" />
</svg>
</body>
</html>
`

	nodes := execHtmlNodes(t, "/html/body/p", html)
	if nodes.String() != "content" {
		t.Error("result not 'content'")
	}

	nodes = execHtmlNodes(t, "/html/@*", html)

	if len(nodes) != 1 {
		t.Error("result not one lang attribute")
	}

	if nodes[0].Node().(node.Attribute).Local() != "lang" || getCursorString(nodes[0]) != "en" {
		t.Error("lang attribute not 'en'")
	}

	nodes = execHtmlNodes(t, "/html/body/svg", html)

	if len(nodes) != 1 || nodes[0].Node().(node.Element).Local() != "svg" {
		t.Error("svg not selected")
	}

	nodes = execHtmlNodes(t, "/html/body/svg/rect/@href", html)

	if len(nodes) != 1 || nodes[0].Node().(node.Attribute).AttributeValue() != "http://example.com" {
		t.Error("bad href value")
	}
}

func TestNonPointerSliceUnmarshal(t *testing.T) {
	sl := make([]int, 0)
	xml := `
<root>
	<elem>1</elem>
	<elem>2</elem>
	<elem>3</elem>
</root>
`

	nodes := execXmlNodes(t, "/root/elem", xml)
	err := Unmarshal(nodes, sl)
	if err.Error() != "field <slice> is not settable" {
		t.Error("incorrect error:", err)
	}

	err = Unmarshal(nodes, &sl)
	if err != nil {
		t.Error("got error:", err)
	}

	if !reflect.DeepEqual(sl, []int{1, 2, 3}) {
		t.Error("incorrect result:", sl)
	}
}

func TestUnmarshal(t *testing.T) {
	type SubUnmarshalTarget struct {
		A      *string `xsel:"a"`
		Battr  bool    `xsel:"b/@attr"`
		Ignore string
	}

	type SliceUnmarshalTarget struct {
		Elem string `xsel:"."`
	}

	type UnmarshalTarget struct {
		Text        string                   `xsel:"normalize-space(text())"`
		Attr        float32                  `xsel:"node/@attr"`
		Attr64      float64                  `xsel:"node/@attr"`
		Subfield    **SubUnmarshalTarget     `xsel:"node"`
		Slice       *[]*SliceUnmarshalTarget `xsel:"slice/elem"`
		StringSlice []string                 `xsel:"slice/elem"`
		Uint8       uint8                    `xsel:"slice/elem[1]"`
		Int8        int8                     `xsel:"slice/elem[1]"`
		Uint16      uint16                   `xsel:"slice/elem[1]"`
		Int16       int16                    `xsel:"slice/elem[1]"`
		Uint32      uint32                   `xsel:"slice/elem[1]"`
		Int32       int32                    `xsel:"slice/elem[1]"`
		Uint64      uint64                   `xsel:"slice/elem[1]"`
		Int64       int64                    `xsel:"slice/elem[1]"`
		Uint        uint                     `xsel:"slice/elem[1]"`
		Int         int                      `xsel:"slice/elem[1]"`
	}

	xml := `
<root>
	foo
	<node attr="3.14">
		<a>a</a>
		<b attr="true"/>
	</node>
	<slice>
		<elem>1</elem>
		<elem>2</elem>
		<elem>3</elem>
	</slice>
</root>
`
	nodes := execXmlNodes(t, "/root", xml)
	target := UnmarshalTarget{}

	if err := Unmarshal(nodes, &target); err != nil {
		t.Error(err)
	}

	a := "a"
	subExpected := &SubUnmarshalTarget{
		A:     &a,
		Battr: true,
	}
	sliceResults := []*SliceUnmarshalTarget{{"1"}, {"2"}, {"3"}}
	expected := UnmarshalTarget{
		Text:        "foo",
		Attr:        3.14,
		Attr64:      3.14,
		Subfield:    &subExpected,
		Slice:       &sliceResults,
		StringSlice: []string{"1", "2", "3"},
		Uint8:       1,
		Int8:        1,
		Uint16:      1,
		Int16:       1,
		Uint32:      1,
		Int32:       1,
		Uint64:      1,
		Int64:       1,
		Uint:        1,
		Int:         1,
	}

	if !reflect.DeepEqual(expected, target) {
		t.Error("incorrect result")
	}
}
