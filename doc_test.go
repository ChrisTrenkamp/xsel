package xsel_test

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func ExampleExec() {
	xml := `
<root>
	<a>This is an XML node.</a>
</root>
`

	xpath := xsel.MustBuildExpr(`/root/a`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.Exec(cursor, &xpath)

	fmt.Println(result)
	// Output: This is an XML node.
}

func ExampleExecAsString() {
	xml := `
<root>
	<a>This is the first node.</a>
	<a>This is the second node.</a>
</root>
`

	xpath := xsel.MustBuildExpr(`/root/a`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.ExecAsString(cursor, &xpath)

	// Be careful when returning the string result of NodeSet's.
	// Only the first node's string value will be returned.
	// If you want the string value of all node's, use ExecAsNodeset.

	fmt.Println(result)
	// Output: This is the first node.
}

func ExampleExecAsNumber() {
	xml := `
<root>
	<a>3.14</a>
	<a>9001</a>
</root>
`

	xpath := xsel.MustBuildExpr(`/root/a`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.ExecAsNumber(cursor, &xpath)

	// Be careful when returning the number result of NodeSet's.
	// Only the first node's value value will be returned.

	fmt.Println(result)
	// Output: 3.14
}

func ExampleExecAsNodeset() {
	xml := `
<root>
	<a>This is the first node.</a>
	<a>This is the second node.</a>
</root>
`

	xpath := xsel.MustBuildExpr(`/root/a`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.ExecAsNodeset(cursor, &xpath)

	for _, i := range result {
		fmt.Println(xsel.GetCursorString(i))
	}

	// Output: This is the first node.
	// This is the second node.
}

func ExampleExecAsNodeset_subqueries() {
	xml := `
<root>
	<a><b>Some text inbetween b and c. <c>A descendant c element.</c></b></a>
	<a><d>A d element.</d><c>A c element.</c></a>
</root>
`

	aQuery := xsel.MustBuildExpr(`/root/a`)
	rootCursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	aElements, _ := xsel.ExecAsNodeset(rootCursor, &aQuery)

	cQuery := xsel.MustBuildExpr(`.//c`)

	for i, a := range aElements {
		cElements, _ := xsel.ExecAsNodeset(a, &cQuery)
		fmt.Printf("%d: %s\n", i, cElements.String())
	}

	// Output: 0: A descendant c element.
	// 1: A c element.
}

func ExampleWithNS() {
	xml := `
<root xmlns="http://some.namespace.com">
	<a xmlns="http://some.namespace.com">This is an XML node with a namespace prefix.</a>
</root>
`

	xpath := xsel.MustBuildExpr(`/ns:root/ns:a`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.Exec(cursor, &xpath, xsel.WithNS("ns", "http://some.namespace.com"))

	fmt.Println(result)
	// Output: This is an XML node with a namespace prefix.
}

func ExampleWithVariableNS() {
	xml := `
<root>
	<node>2.50</node>
	<node>3.14</node>
	<node>0.30</node>
</root>
`

	const NS = "http://some.namespace.com"

	xpath := xsel.MustBuildExpr(`//node()[. = $ns:mynum]`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.Exec(cursor, &xpath, xsel.WithNS("ns", NS), xsel.WithVariableNS(NS, "mynum", xsel.Number(3.14)))

	fmt.Println(result)
	// Output: 3.14
}

func ExampleWithFunction() {
	xml := `
<root>
	<a>This is an element.</a>
	<!-- This is a comment. -->
</root>
`

	isComment := func(context xsel.Context, args ...xsel.Result) (xsel.Result, error) {
		nodeSet, isNodeSet := context.Result().(xsel.NodeSet)

		if !isNodeSet || len(nodeSet) == 0 {
			return xsel.Bool(false), nil
		}

		_, isComment := nodeSet[0].Node().(xsel.Comment)
		return xsel.Bool(isComment), nil
	}

	xpath := xsel.MustBuildExpr(`//node()[is-comment()]`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.Exec(cursor, &xpath, xsel.WithFunction("is-comment", isComment))

	fmt.Println(result)
	// Output: This is a comment.
}

func ExampleReadJson() {
	json := `
{
	"states": ["AK", ["MD", "FL"] ]
}
`

	xpath := xsel.MustBuildExpr(`/#obj/states/#arr/text()`)
	cursor, _ := xsel.ReadJson(bytes.NewBufferString(json))
	result, _ := xsel.Exec(cursor, &xpath)

	fmt.Println(result)

	xpath = xsel.MustBuildExpr(`/#obj/states/#arr/#arr/text()[2]`)
	result, _ = xsel.Exec(cursor, &xpath)

	fmt.Println(result)
	// Output: AK
	// FL
}

func ExampleUnmarshal() {
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

	contextSettings := xsel.WithNS("NS", "http://www.adventure-works.com")
	xpath := xsel.MustBuildExpr(`/NS:Root`)
	cursor, _ := xsel.ReadXml(bytes.NewBufferString(xml))
	result, _ := xsel.Exec(cursor, &xpath, contextSettings)

	customers := Customers{}
	_ = xsel.Unmarshal(result, &customers, contextSettings) // Remember to check for errors

	fmt.Printf("%+v\n", customers)
	// Output: {Customers:[{Id:GREAL Name:Great Lakes Food Market ContactName:Howard Snyder Address:{Address:2732 Baker Blvd. City:Eugene Region:OR}} {Id:HUNGC Name:Hungry Coyote Import Store ContactName:Yoshi Latimer Address:{Address:City Center Plaza 516 Main St. City:Walla Walla Region:WA}}]}
}
