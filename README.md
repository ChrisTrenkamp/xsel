# xsel

[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate?business=PJDCE35ARU76Q&currency_code=USD) [![Go Reference](https://pkg.go.dev/badge/github.com/ChrisTrenkamp/xsel.svg)](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel)


`xsel` is a library that (almost) implements the XPath 1.0 specification.  The non-compliant bits are:

* `xsel` does not implement the [id](https://www.w3.org/TR/xpath-10/#function-id) function.
* The grammar as defined in the XPath 1.0 spec doesn't explicitly allow function calls in the middle of a path expression, such as `/path/function-call()/path`.  `xsel` allows function calls in the middle of path expressions.
* `xsel` allows name lookups with a wildcard for the namespace, such as `/*:path`.
* `xsel` allows the `#` character in element selections.

## Basic usage

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func main() {
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
```

## Binding variables and namespaces

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func main() {
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
```

## Binding variables

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func main() {
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
```

## Binding custom functions

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func main() {
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
```

## Unmarshal result into a struct

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func main() {
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
	xsel.Unmarshal(result, &customers, contextSettings) // Remember to check for errors

	fmt.Printf("%+v\n", customers)
	// Output: {Customers:[{Id:GREAL Name:Great Lakes Food Market ContactName:Howard Snyder Address:{Address:2732 Baker Blvd. City:Eugene Region:OR}} {Id:HUNGC Name:Hungry Coyote Import Store ContactName:Yoshi Latimer Address:{Address:City Center Plaza 516 Main St. City:Walla Walla Region:WA}}]}
}
```

## Extensible

`xsel` supplies an XML parser (using the `encoding/xml` package) out of the box, but the XPath logic does not depend directly on XML.  It instead depends on the interfaces defined in the [node](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node) and [store](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/store) packages.  This means it's possible to use `xsel` for querying against non-XML documents.  The [parser](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/parser) package supplies methods for parsing XML, HTML, and JSON documents.

To build a custom document, implement your own [Parser](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/parser#Parser) method, and build [Element](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Element)'s, [Attribute](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Attribute)'s [Character Data](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#CharData), [Comment](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Comment)'s, [Processing Instruction](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#ProcInst)'s, and [Namespace](https://pkg.go.dev/github.com/ChrisTrenkamp/xsel/node#Namespace)'s.


## HTML documents

Use the `xsel.ReadHtml` function to read HTML documents. Namespaces are completely ignored for HTML documents.  Keep all queries in the default namespace.  Write queries such as `//svg`.  Do not write queries such as `//svg:svg`.

## JSON documents

JSON documents only build elements and character data.  Object and array declarations will omit an element node with the name `#`.  So for example, given the following JSON file:

```
{
	"states": ["AK", ["MD", "FL"] ]
}
```

It would look like this in XML...

```
<#>
	<states>
		<#>
			AK
			<#>
				MD
				FL
			</#>
		</#>
	</states>
</#>
```

... however, `MD` and `FL` are separate text nodes, which is different from XML parsing:


```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ChrisTrenkamp/xsel"
)

func main() {
	json := `
{
	"states": ["AK", ["MD", "FL"] ]
}
`

	xpath := xsel.MustBuildExpr(`/#/states/#/text()`)
	cursor, _ := xsel.ReadJson(bytes.NewBufferString(json))
	result, _ := xsel.Exec(cursor, &xpath)

	fmt.Println(result)

	// Notice the [2] in the text selection.
	xpath = xsel.MustBuildExpr(`/#/states/#/#/text()[2]`)
	result, _ = xsel.Exec(cursor, &xpath)

	fmt.Println(result)
	// Output: AK
	// FL
}
```

## Commandline Utility

`xsel` supplies a grep-like commandline utility for querying XML documents:

```
$ go install github.com/ChrisTrenkamp/xsel/xsel@latest
$ xsel -h
Usage of xsel:
  -a    If the result is a NodeSet, print the string value of all the nodes instead of just the first
  -c int
        Run queries in the given number of concurrent workers (beware that results will have no predictable order) (default 1)
  -e value
        Bind an entity value e.g. entityname=entityval
  -m    If the result is a NodeSet, print all the results as XML
  -n    Suppress filenames
  -r    Recursively traverse directories
  -s value
        Namespace mapping. e.g. -ns companyns=http://company.com
  -t string
        Force xsel to parse files as the given type.  Can be 'xml', 'html', or 'json'.  If unspecified, the file will be detected by its MIME type.  Must be specified when reading from stdin.
  -u    Turns off strict XML decoding
  -v value
        Bind a variable (all variables are bound as string types) e.g. -v var=value or -v companyns:var=value
  -x string
        XPath expression to execute (required)
```

## CLI examples

```
$ cat test.xml
<?xml version="1.0" encoding="UTF-8"?>
<root>
  <a xmlns="http://a">Element a</a>
  <b>Element b</b>
</root>
```

This is a basic query:
```
$ xsel -x '/root/b' test.xml
test.xml: Element b
```

This is a basic query on stdin:
```
$ cat foo.xml | xsel -x '/root/b' -
Element b
```

This query has multiple results, but only the first value is printed:
```
$ xsel -x '/root/*' test.xml
test.xml: Element a
```

This query has multiple results, and all values are printed:
```
$ xsel -x '/root/*' -a test.xml
test.xml: Element a
test.xml: Element b
```

Print all results as XML:
```
$ xsel -x '/root/*' -m test.xml
test.xml: <a xmlns="http://a">Element a</a>
test.xml: <b>Element b</b>
```

Suppress the filename when printing results:
```
$ xsel -x '/root/*' -m -n test.xml
<a xmlns="http://a">Element a</a>
<b>Element b</b>
```

Bind a namespace:
```
$ xsel -x '//a:*' -s a='http://a' -m test.xml
test.xml: <a xmlns="http://a">Element a</a>
```

Bind a variable (variables are bound as strings):
```
$ xsel -x '//*[. = $textval]' -v textval="Element b" test.xml
test.xml: Element b
```
