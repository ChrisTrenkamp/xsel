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

func ExampleExec_unmarshal() {
	xml := `
<Root xmlns="http://www.adventure-works.com">
	<Customers>
		<Customer CustomerID="GREAL">
			<CompanyName>Great Lakes Food Market</CompanyName>
			<ContactName>Howard Snyder</ContactName>
			<ContactTitle>Marketing Manager</ContactTitle>
			<FullAddress>
				<Address>2732 Baker Blvd.</Address>
				<City>Eugene</City>
				<Region>OR</Region>
			</FullAddress>
		</Customer>
		<Customer CustomerID="HUNGC">
		  <CompanyName>Hungry Coyote Import Store</CompanyName>
		  <ContactName>Yoshi Latimer</ContactName>
		  <FullAddress>
			<Address>City Center Plaza 516 Main St.</Address>
			<City>Walla Walla</City>
			<Region>WA</Region>
		  </FullAddress>
		</Customer>
	</Customers>
</Root>
`

	type Address struct {
		Address string `xsel:"NS:Address"`
		City    string `xsel:"NS:City"`
		Region  string `xsel:"NS:Region"`
	}

	type Customer struct {
		Id          string  `xsel:"@CustomerID"`
		Name        string  `xsel:"NS:CompanyName"`
		ContactName string  `xsel:"NS:ContactName"`
		Address     Address `xsel:"NS:FullAddress"`
	}

	type Customers struct {
		Customers []Customer `xsel:"NS:Customers/NS:Customer"`
	}

	contextSettings := func(c *exec.ContextSettings) {
		c.NamespaceDecls["NS"] = "http://www.adventure-works.com"
	}

	xpath := grammar.MustBuild(`/NS:Root`)
	parser := parser.ReadXml(bytes.NewBufferString(xml))
	cursor, _ := store.CreateInMemory(parser)
	result, _ := exec.Exec(cursor, &xpath, contextSettings)

	customers := Customers{}
	exec.Unmarshal(result, &customers, contextSettings) // Remember to check for errors

	fmt.Printf("%+v\n", customers)
	//{Customers:[{Id:GREAL Name:Great Lakes Food Market ContactName:Howard Snyder Address:{Address:2732 Baker Blvd. City:Eugene Region:OR}}
	// {Id:HUNGC Name:Hungry Coyote Import Store ContactName:Yoshi Latimer Address:{Address:City Center Plaza 516 Main St. City:Walla Walla Region:WA}}]}
}
