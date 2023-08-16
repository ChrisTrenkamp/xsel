package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
)

type keyValuePair map[string]string

func (n keyValuePair) String() string {
	ret := bytes.Buffer{}

	for k, v := range n {
		fmt.Fprintf(&ret, "%s=%s", k, v)
	}

	return ret.String()
}

func (n keyValuePair) Set(value string) error {
	nsMap := strings.Split(value, "=")

	if len(nsMap) != 2 {
		return fmt.Errorf("invalid namespace mapping: %s", value)
	}

	n[nsMap[0]] = nsMap[1]
	return nil
}

var VALID_FILE_TYPES = map[string]bool{
	"xml":  true,
	"html": true,
	"json": true,
}

var concurrent = flag.Int("c", 1, "Run queries in the given number of concurrent workers (beware that results will have no predictable order)")
var printAllNodes = flag.Bool("a", false, "If the result is a NodeSet, print the string value of all the nodes instead of just the first")
var suppressFileNames = flag.Bool("n", false, "Suppress filenames")
var recursive = flag.Bool("r", false, "Recursively traverse directories")
var asXml = flag.Bool("m", false, "If the result is a NodeSet, print all the results as XML")
var unstrict = flag.Bool("u", false, "Turns off strict XML decoding")
var fileType = flag.String("t", "", "Force xsel to parse files as the given type.  Can be 'xml', 'html', or 'json'.  If unspecified, the file will be detected by its MIME type.  Must be specified when reading from stdin.")
var entities = make(keyValuePair)
var namespaces = make(keyValuePair)
var fileSync = sync.WaitGroup{}
var xpath grammar.Grammar
var variableBindings = make(map[exec.XmlName]exec.Result)
var semaphore chan struct{}

func main() {
	variableDeclarations := make(keyValuePair)
	xpathExpr := flag.String("x", "", "XPath expression to execute (required)")
	flag.Var(namespaces, "s", "Namespace mapping. e.g. -ns companyns=http://company.com")
	flag.Var(variableDeclarations, "v", "Bind a variable (all variables are bound as string types) e.g. -v var=value or -v companyns:var=value")
	flag.Var(entities, "e", "Bind an entity value e.g. entityname=entityval")
	flag.Parse()

	max := *concurrent

	if max < 1 {
		max = 1
	}

	semaphore = make(chan struct{}, max)
	args := flag.Args()

	if *fileType != "" {
		if !VALID_FILE_TYPES[*fileType] {
			fmt.Fprintf(os.Stderr, "Invalid file type, '%s'", *fileType)
			return
		}
	}

	if *xpathExpr == "" {
		fmt.Fprintln(os.Stderr, "Missing XPath expression")
		return
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Missing file arguments")
		return
	}

	xpathTest, err := grammar.Build(*xpathExpr)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Bad XPath expression:", err)
		return
	}

	xpath = xpathTest

	for name, value := range variableDeclarations {
		qName, err := exec.GetQName(name, namespaces)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		variableBindings[qName] = exec.String(value)
	}

	for _, file := range args {
		if file == "-" {
			semaphore <- struct{}{}
			fileSync.Add(1)

			go runXpathOnStdin()
		} else {
			filepath.WalkDir(file, walker)
		}
	}

	fileSync.Wait()
}

func walker(path string, d fs.DirEntry, err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error traversing %s: %s\n", path, err)
		return nil
	}

	if !d.IsDir() {
		semaphore <- struct{}{}
		fileSync.Add(1)

		if *concurrent > 1 {
			go runXpathOnFile(path)
		} else {
			runXpathOnFile(path)
		}

		return nil
	}

	if *recursive {
		return nil
	}

	fmt.Fprintf(os.Stderr, "%s is a directory\n", path)
	return fs.SkipDir
}

func runXpathOnFile(path string) {
	defer fileSync.Done()
	defer func() {
		<-semaphore
	}()

	parseType := *fileType

	if parseType == "" {
		media, _, err := mime.ParseMediaType(mime.TypeByExtension(filepath.Ext(path)))

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting file MIME type for file %s: %s\n", path, err)
			return
		}

		if strings.Contains(media, "xml") {
			parseType = "xml"
		} else if strings.Contains(media, "html") {
			parseType = "html"
		} else if strings.Contains(media, "json") {
			parseType = "json"
		} else {
			fmt.Fprintf(os.Stderr, "Unsupported MIME type for file %s: %s\n", path, media)
			return
		}
	}

	file, err := os.Open(path)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", path, err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file %s: %s\n", path, err)
		}
	}()

	parser, err := createCursor(file, parseType)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", path, err)
		return
	}

	executeXpath(parser, path)
}

func runXpathOnStdin() {
	defer fileSync.Done()
	defer func() {
		<-semaphore
	}()

	if *fileType == "" {
		fmt.Fprintln(os.Stderr, "Must specify file type when reading from stdin")
		return
	}

	parser, err := createCursor(os.Stdin, *fileType)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %s\n", err)
		return
	}

	executeXpath(parser, "-")
}

func buildXmlParserSettings(d *xml.Decoder) {
	d.Strict = !*unstrict
	d.Entity = entities
}

func createCursor(in io.Reader, parseType string) (parser.Parser, error) {
	switch parseType {
	case "xml":
		return parser.ReadXml(in, buildXmlParserSettings), nil
	case "html":
		return parser.ReadHtml(in)
	case "json":
		return parser.ReadJson(in), nil
	}

	return nil, fmt.Errorf("unsupported file type")
}

func executeXpath(parser parser.Parser, path string) {
	cursor, err := store.CreateInMemory(parser)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file %s: %s\n", path, err)
		return
	}

	result, err := exec.Exec(cursor, &xpath, buildContextSettings)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing XPath function on file %s: %s\n", path, err)
		return
	}

	nodeSet, isNodeSet := result.(exec.NodeSet)

	if isNodeSet && len(nodeSet) == 0 {
		return
	}

	buffer := bytes.Buffer{}

	if isNodeSet && *asXml {
		writeXmlResult(&buffer, path, nodeSet)
	} else if isNodeSet && *printAllNodes {
		for _, node := range nodeSet {
			writeResult(&buffer, path, exec.NodeSet{node})
		}
	} else {
		writeResult(&buffer, path, result)
	}

	fmt.Print(buffer.String())
}

func buildContextSettings(c *exec.ContextSettings) {
	c.NamespaceDecls = namespaces
	c.Variables = variableBindings
}

func writeResult(buffer *bytes.Buffer, path string, result exec.Result) {
	if *suppressFileNames || path == "-" {
		fmt.Fprintf(buffer, "%s\n", result.String())
	} else {
		fmt.Fprintf(buffer, "%s: %s\n", path, result.String())
	}
}

func writeXmlResult(buffer *bytes.Buffer, path string, result exec.NodeSet) {
	for _, i := range result {
		nextResult := bytes.Buffer{}

		if !*suppressFileNames && path != "-" {
			fmt.Fprintf(&nextResult, "%s: ", path)
		}

		encoder := xml.NewEncoder(&nextResult)
		err := encodeCursorToXml(encoder, i)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error printing results for file %s: %s\n", path, err)
			return
		}

		if err = encoder.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing results for file %s: %s\n", path, err)
			return
		}

		result := bytes.ReplaceAll(nextResult.Bytes(), ([]byte)("\n"), ([]byte)("&#10;"))
		buffer.Write(result)
		buffer.WriteString("\n")
	}
}

func encodeCursorToXml(encoder *xml.Encoder, cursor store.Cursor) error {
	switch n := cursor.Node().(type) {
	case node.Attribute:
		return writeAttributeAsProcInst(encoder, n)
	case node.CharData:
		val := ([]byte)(n.CharDataValue())
		t := xml.CharData(val)

		return encoder.EncodeToken(t)
	case node.Comment:
		val := ([]byte)(n.CommentValue())
		t := xml.Comment(val)

		return encoder.EncodeToken(t)
	case node.Element:
		err := writeElementToken(encoder, cursor)

		if err != nil {
			return err
		}

		for _, i := range cursor.Children() {
			err = encodeCursorToXml(encoder, i)

			if err != nil {
				return err
			}
		}

		err = writeEndElement(encoder, cursor)

		if err != nil {
			return err
		}
	case node.Namespace:
		return writeNamespaceAsProcInst(encoder, n)
	case node.ProcInst:
		t := xml.ProcInst{
			Target: n.Target(),
			Inst:   []byte(n.ProcInstValue()),
		}

		return encoder.EncodeToken(t)
	case node.Root:
		for _, i := range cursor.Children() {
			err := encodeCursorToXml(encoder, i)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeAttributeAsProcInst(encoder *xml.Encoder, attr node.Attribute) error {
	target := "attribute:"

	if attr.Space() != "" {
		target += attr.Space() + ":"
	}

	target += attr.Local()

	val := ([]byte)(attr.AttributeValue())

	t := xml.ProcInst{
		Target: target,
		Inst:   val,
	}

	return encoder.EncodeToken(t)
}

func writeNamespaceAsProcInst(encoder *xml.Encoder, ns node.Namespace) error {
	target := "namespace:" + ns.Prefix()
	val := ([]byte)(ns.NamespaceValue())

	t := xml.ProcInst{
		Target: target,
		Inst:   val,
	}

	return encoder.EncodeToken(t)
}

func writeElementToken(encoder *xml.Encoder, elem store.Cursor) error {
	n := elem.Node().(node.Element)
	t := xml.StartElement{
		Name: xml.Name{
			Space: n.Space(),
			Local: n.Local(),
		},
	}

	for _, i := range elem.Attributes() {
		attr := i.Node().(node.Attribute)
		attrTok := xml.Attr{
			Name:  createXmlName(attr),
			Value: attr.AttributeValue(),
		}

		t.Attr = append(t.Attr, attrTok)
	}

	return encoder.EncodeToken(t)
}

func writeEndElement(encoder *xml.Encoder, elem store.Cursor) error {
	n := elem.Node().(node.Element)
	t := xml.EndElement{
		Name: createXmlName(n),
	}

	return encoder.EncodeToken(t)
}

func createXmlName(n node.NamedNode) xml.Name {
	return xml.Name{
		Space: n.Space(),
		Local: n.Local(),
	}
}
