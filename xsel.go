package xsel

import (
	"fmt"
	"io"

	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

type ContextSettings = exec.ContextSettings
type Context = exec.Context
type ContextApply = exec.ContextApply

type Grammar = grammar.Grammar

type Bool = exec.Bool
type Number = exec.Number
type String = exec.String
type NodeSet = exec.NodeSet
type Result = exec.Result
type XmlName = exec.XmlName
type Function = exec.Function

type Node = node.Node
type Root = node.Root
type NamedNode = node.NamedNode
type Element = node.Element
type Namespace = node.Namespace
type Attribute = node.Attribute
type CharData = node.CharData
type Comment = node.Comment
type ProcInst = node.ProcInst

type Parser = parser.Parser
type XmlParseOptions = parser.XmlParseOptions

type Cursor = store.Cursor

// WithNS binds a namespace name to a XPath query.
func WithNS(name, url string) func(c *ContextSettings) {
	return func(c *ContextSettings) {
		c.NamespaceDecls[name] = url
	}
}

// WithVariable binds a variable name with no namespace to a XPath query.
func WithVariable(local string, value Result) func(c *ContextSettings) {
	return WithVariableNS("", local, value)
}

// WithVariableNS binds a variable name with a namespace to a XPath query.
func WithVariableNS(space, local string, value Result) func(c *ContextSettings) {
	return WithVariableName(XmlName{Space: space, Local: local}, value)
}

// WithVariableName binds a variable name with a namespace to a XPath query.
func WithVariableName(name XmlName, value Result) func(c *ContextSettings) {
	return func(c *ContextSettings) {
		c.Variables[name] = value
	}
}

// WithFunction binds a custom function name with no namespace to a XPath query.
func WithFunction(local string, fn Function) func(c *ContextSettings) {
	return WithFunctionNS("", local, fn)
}

// WithFunctionNS binds a custom function name with a namespace to a XPath query.
func WithFunctionNS(space, local string, fn Function) func(c *ContextSettings) {
	return WithFunctionName(XmlName{Space: space, Local: local}, fn)
}

// WithFunctionName binds a custom function name with a namespace to a XPath query.
func WithFunctionName(name XmlName, fn Function) func(c *ContextSettings) {
	return func(c *ContextSettings) {
		c.FunctionLibrary[name] = fn
	}
}

func GetQName(input string, namespaces map[string]string) (XmlName, error) {
	return exec.GetQName(input, namespaces)
}

// BuildExpr creates an XPath query.
func BuildExpr(xpath string) (Grammar, error) {
	return grammar.Build(xpath)
}

// MustBuildExpr is like BuildExpr, but panics if an error is thrown.
func MustBuildExpr(xpath string) Grammar {
	return grammar.MustBuild(xpath)
}

// ReadXml parses the given XML document and stores the node in memory.
func ReadXml(in io.Reader, opts ...XmlParseOptions) (Cursor, error) {
	parser := parser.ReadXml(in, opts...)
	return store.CreateInMemory(parser)
}

// ReadHtml parses the given HTML document and stores the node in memory.
func ReadHtml(in io.Reader) (Cursor, error) {
	parser, err := parser.ReadHtml(in)

	if err != nil {
		return nil, err
	}

	return store.CreateInMemory(parser)
}

// ReadJson parses the given JSON document and stores the node in memory.
func ReadJson(in io.Reader) (Cursor, error) {
	parser := parser.ReadJson(in)
	return store.CreateInMemory(parser)
}

// Exec executes an XPath query against the given Cursor and returns the result.
func Exec(cursor Cursor, expr *Grammar, settings ...ContextApply) (Result, error) {
	return exec.Exec(cursor, expr, settings...)
}

// Like Exec, except it returns the string result of the query.
func ExecAsString(cursor Cursor, expr *Grammar, settings ...ContextApply) (string, error) {
	ret, err := exec.Exec(cursor, expr, settings...)
	if err != nil {
		return "", err
	}

	return ret.String(), nil
}

// Like Exec, except it returns the query as a number.
func ExecAsNumber(cursor Cursor, expr *Grammar, settings ...ContextApply) (float64, error) {
	ret, err := exec.Exec(cursor, expr, settings...)
	if err != nil {
		return 0, err
	}

	return ret.Number(), nil
}

// Like Exec, except it returns the query as a NodeSet.
func ExecAsNodeset(cursor Cursor, expr *Grammar, settings ...ContextApply) (NodeSet, error) {
	ret, err := exec.Exec(cursor, expr, settings...)
	if err != nil {
		return nil, err
	}

	nodeset, ok := ret.(NodeSet)
	if !ok {
		return nil, fmt.Errorf("result is not NodeSet")
	}

	return nodeset, nil
}

// GetCursorString is a convenience method to return the string value of
// an individual Node.
func GetCursorString(c Cursor) string {
	return exec.GetCursorString(c)
}

// Unmarshal maps a XPath result to a struct or slice.
// When unmarshaling a slice, the result must be a NodeSet. When unmarshaling
// a struct, the result must be a NodeSet with one result. To unmarshal a
// value to a struct field, give it a "xsel" tag name, and a XPath expression
// for its value (e.g. `xsel:"//my-struct[@my-id = 'my-value']"`).
//
// For struct fields, Unmarshal can set fields that are ints and uints, bools,
// strings, slices, and nested structs.
//
// For slice elements, Unmarshal can set ints and uints, bools, strings, and
// structs.  It cannot Unmarshal multidimensional slices.
//
// Arrays, maps, and channels are not supported.
func Unmarshal(result Result, value any, settings ...ContextApply) error {
	return exec.Unmarshal(result, value, settings...)
}
