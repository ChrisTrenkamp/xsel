package exec

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/grammar/parser/bsr"
	"github.com/ChrisTrenkamp/xsel/grammar/parser/symbols"
	"github.com/ChrisTrenkamp/xsel/node"
	"github.com/ChrisTrenkamp/xsel/store"
)

type Result interface {
	String() string
	Number() float64
	Bool() bool
}

type XmlName struct {
	Space string
	Local string
}

func Name(space, local string) XmlName {
	return XmlName{
		Space: space,
		Local: local,
	}
}

func GetQName(input string, namespaces map[string]string) (XmlName, error) {
	spl := strings.Split(input, ":")
	ret := XmlName{}

	if len(spl) == 1 {
		ret.Local = spl[0]
	} else {
		ns, ok := namespaces[strings.TrimSpace(spl[0])]

		if !ok {
			return XmlName{}, fmt.Errorf("unknown namespace binding '%s'", spl[0])
		}

		ret.Space = ns
		ret.Local = spl[1]
	}

	ret.Local = strings.TrimSpace(ret.Local)
	return ret, nil
}

type Function func(context Context, args ...Result) (Result, error)

// ContextSettings allows you to add namespace mappings, create new functions,
// and add variable bindings to your XPath query.
type ContextSettings struct {
	NamespaceDecls  map[string]string
	FunctionLibrary map[XmlName]Function
	Variables       map[XmlName]Result
	// Context is the initial position to run XPath queries.  This is initialized to the root of
	// the document.  This may be overridden to point to a different position in the document.
	// When overriding the Context, it must be a Cursor contained within the root, or bad things can happen!
	Context store.Cursor
}

type ContextApply func(c *ContextSettings)

type exprContext struct {
	root             store.Cursor
	result           Result
	contextPosition  int
	builtinFunctions map[XmlName]Function
	ContextSettings
}

type Context interface {
	Result() Result
	ContextPosition() int
}

func (c *exprContext) Result() Result {
	return c.result
}

func (c *exprContext) ContextPosition() int {
	return c.contextPosition
}

func (e *exprContext) copy() exprContext {
	return exprContext{
		root:             e.root,
		result:           e.result,
		contextPosition:  e.contextPosition,
		builtinFunctions: builtinFunctions,
		ContextSettings:  e.ContextSettings,
	}
}

func cleanup(nextResult NodeSet) NodeSet {
	return unique(nextResult)
}

type execFn func(context *exprContext, expr *grammar.Grammar) error

var execFunctions = map[symbols.NT]execFn{}

func init() {
	execFunctions[symbols.NT_Number] = execNumber
	execFunctions[symbols.NT_Literal] = execLiteral
	execFunctions[symbols.NT_AdditiveExprAdd] = execAdditiveExprAdd
	execFunctions[symbols.NT_AdditiveExprSubtract] = execAdditiveExprSubtract
	execFunctions[symbols.NT_MultiplicativeExprMultiply] = execMultiplicativeExprMultiply
	execFunctions[symbols.NT_MultiplicativeExprDivide] = execMultiplicativeExprDivide
	execFunctions[symbols.NT_MultiplicativeExprMod] = execMultiplicativeExprMod
	execFunctions[symbols.NT_UnaryExprNegate] = execUnaryExprNegate
	execFunctions[symbols.NT_EqualityExprEqual] = execEqualityExprEqual
	execFunctions[symbols.NT_EqualityExprNotEqual] = execEqualityExprNotEqual
	execFunctions[symbols.NT_RelationalExprLessThan] = execRelationalExprLessThan
	execFunctions[symbols.NT_RelationalExprGreaterThan] = execRelationalExprGreaterThan
	execFunctions[symbols.NT_RelationalExprLessThanOrEqual] = execRelationalExprLessThanOrEqual
	execFunctions[symbols.NT_RelationalExprGreaterThanOrEqual] = execRelationalExprGreaterThanOrEqual
	execFunctions[symbols.NT_OrExprOr] = execOrExprOr
	execFunctions[symbols.NT_AndExprAnd] = execAndExprAnd
	execFunctions[symbols.NT_AbsoluteLocationPathOnly] = execAbsoluteLocationPathOnly
	execFunctions[symbols.NT_RelativeLocationPathWithStep] = leftRightDependentResult
	execFunctions[symbols.NT_Step] = execStep
	execFunctions[symbols.NT_NameTestQNameLocalOnly] = execNameTestQNameLocalOnly
	execFunctions[symbols.NT_NodeTestNodeTypeNoArgTestNodeTestConflictResolver] = execNameTestQNameLocalOnly
	execFunctions[symbols.NT_NodeTestAndPredicate] = leftRightDependentResult
	execFunctions[symbols.NT_Predicate] = execPredicate
	execFunctions[symbols.NT_UnionExprUnion] = execUnionExprUnion
	execFunctions[symbols.NT_NodeTestNodeTypeNoArgTest] = execNodeTestNodeTypeNoArgTest
	execFunctions[symbols.NT_NodeTestProcInstTargetTest] = execNodeTestProcInstTargetTest
	execFunctions[symbols.NT_NameTestAnyElement] = execNameTestAnyElement
	execFunctions[symbols.NT_NameTestNamespaceAnyLocal] = execNameTestNamespaceAnyLocal
	execFunctions[symbols.NT_NameTestLocalAnyNamespace] = execNameTestLocalAnyNamespace
	execFunctions[symbols.NT_NameTestQNameNamespaceWithLocal] = execNameTestQNameNamespaceWithLocal
	execFunctions[symbols.NT_StepWithAxisAndNodeTest] = leftRightDependentResult
	execFunctions[symbols.NT_AxisName] = execAxisName
	execFunctions[symbols.NT_AbbreviatedStepParent] = execAbbreviatedStepParent
	execFunctions[symbols.NT_AbbreviatedAxisSpecifier] = execAbbreviatedAxisSpecifier
	execFunctions[symbols.NT_AbbreviatedAbsoluteLocationPath] = execAbbreviatedAbsoluteLocationPath
	execFunctions[symbols.NT_AbbreviatedRelativeLocationPath] = execAbbreviatedRelativeLocationPath
	execFunctions[symbols.NT_FunctionCall] = execFunctionCall
	execFunctions[symbols.NT_VariableReference] = execVariableReference
}

// Executes an XPath query against the given Cursor that is pointing to the node.Root.
func Exec(cursor store.Cursor, expr *grammar.Grammar, settings ...ContextApply) (Result, error) {
	contextSettings := ContextSettings{
		Variables:       make(map[XmlName]Result),
		FunctionLibrary: make(map[XmlName]Function),
		NamespaceDecls:  make(map[string]string),
		Context:         cursor,
	}

	for _, i := range settings {
		i(&contextSettings)
	}

	context := &exprContext{
		root:             cursor,
		result:           Result(NodeSet{contextSettings.Context}),
		contextPosition:  0,
		builtinFunctions: builtinFunctions,
		ContextSettings:  contextSettings,
	}

	err := execContext(context, expr)

	if err != nil {
		return nil, err
	}

	return context.result, nil
}

func execContext(context *exprContext, expr *grammar.Grammar) error {
	name := expr.BSR.Label.Slot().NT
	exec := execFunctions[name]

	if exec != nil {
		return exec(context, expr)
	}

	return execChildren(context, expr)
}

func execChildren(context *exprContext, expr *grammar.Grammar) error {
	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			return execContext(context, expr.Next(&c))
		}
	}

	return nil
}

func execNumber(context *exprContext, expr *grammar.Grammar) error {
	numStr := expr.GetString()
	numResult, err := strconv.ParseFloat(numStr, 64)

	context.result = Number(numResult)
	return err
}

func execLiteral(context *exprContext, expr *grammar.Grammar) error {
	literal := expr.GetString()
	literal = literal[1 : len(literal)-1]

	context.result = String(literal)
	return nil
}

func execAdditiveExprAdd(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentNumber(context, expr)

	if err != nil {
		return err
	}

	context.result = Number(left + right)
	return nil
}

func execAdditiveExprSubtract(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentNumber(context, expr)

	if err != nil {
		return err
	}

	context.result = Number(left - right)
	return nil
}

func execMultiplicativeExprMultiply(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentNumber(context, expr)

	if err != nil {
		return err
	}

	context.result = Number(left * right)
	return nil
}

func execMultiplicativeExprDivide(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentNumber(context, expr)

	if err != nil {
		return err
	}

	if right == 0 {
		if left == 0 {
			context.result = Number(math.NaN())
		} else if left > 0 {
			context.result = Number(math.Inf(1))
		} else {
			context.result = Number(math.Inf(-1))
		}

		return nil
	}

	context.result = Number(left / right)
	return nil
}

func execMultiplicativeExprMod(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentNumber(context, expr)

	if err != nil {
		return err
	}

	context.result = Number(int(left) % int(right))
	return nil
}

func execUnaryExprNegate(context *exprContext, expr *grammar.Grammar) error {
	left, err := leftOnlyIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNum := left.Number()

	context.result = Number(-leftNum)
	return nil
}

func execEqualityExprEqual(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, leftNodeSetOk := left.(NodeSet)
	rightNodeSet, rightNodeSetOk := right.(NodeSet)

	if leftNodeSetOk && rightNodeSetOk {
		for _, leftNode := range leftNodeSet {
			for _, rightNode := range rightNodeSet {
				if getCursorString(leftNode) == getCursorString(rightNode) {
					context.result = Bool(true)
					return nil
				}
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftNumber, leftNumberOk := left.(Number)

	if leftNumberOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftNumber == Number(getStringNumber(getCursorString(rightNode))) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightNumber, rightNumberOk := right.(Number)

	if leftNodeSetOk && rightNumberOk {
		for _, leftNode := range leftNodeSet {
			if Number(getStringNumber(getCursorString(leftNode))) == rightNumber {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftString, leftStringOk := left.(String)

	if leftStringOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftString == String(getCursorString(rightNode)) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightString, rightStringOk := right.(String)

	if leftNodeSetOk && rightStringOk {
		for _, leftNode := range leftNodeSet {
			if String(getCursorString(leftNode)) == rightString {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftBool, leftBoolOk := left.(Bool)

	if leftBoolOk && rightNodeSetOk {
		context.result = Bool(bool(leftBool) == rightNodeSet.Bool())
		return nil
	}

	rightBool, rightBoolOk := right.(Bool)

	if leftNodeSetOk && rightBoolOk {
		context.result = Bool(leftNodeSet.Bool() == bool(rightBool))
		return nil
	}

	if leftBoolOk || rightBoolOk {
		context.result = Bool(left.Bool() == right.Bool())
		return nil
	}

	if leftNumberOk || rightNumberOk {
		context.result = Bool(left.Number() == right.Number())
		return nil
	}

	context.result = Bool(left.String() == right.String())
	return nil
}

func execEqualityExprNotEqual(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, leftNodeSetOk := left.(NodeSet)
	rightNodeSet, rightNodeSetOk := right.(NodeSet)

	if leftNodeSetOk && rightNodeSetOk {
		for _, leftNode := range leftNodeSet {
			for _, rightNode := range rightNodeSet {
				if getCursorString(leftNode) != getCursorString(rightNode) {
					context.result = Bool(true)
					return nil
				}
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftNumber, leftNumberOk := left.(Number)

	if leftNumberOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftNumber != Number(getStringNumber(getCursorString(rightNode))) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightNumber, rightNumberOk := right.(Number)

	if leftNodeSetOk && rightNumberOk {
		for _, leftNode := range leftNodeSet {
			if Number(getStringNumber(getCursorString(leftNode))) != rightNumber {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftString, leftStringOk := left.(String)

	if leftStringOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftString != String(getCursorString(rightNode)) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightString, rightStringOk := right.(String)

	if leftNodeSetOk && rightStringOk {
		for _, leftNode := range leftNodeSet {
			if String(getCursorString(leftNode)) != rightString {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftBool, leftBoolOk := left.(Bool)

	if leftBoolOk && rightNodeSetOk {
		context.result = Bool(bool(leftBool) != rightNodeSet.Bool())
		return nil
	}

	rightBool, rightBoolOk := right.(Bool)

	if leftNodeSetOk && rightBoolOk {
		context.result = Bool(leftNodeSet.Bool() != bool(rightBool))
		return nil
	}

	if leftBoolOk || rightBoolOk {
		context.result = Bool(left.Bool() != right.Bool())
		return nil
	}

	if leftNumberOk || rightNumberOk {
		context.result = Bool(left.Number() != right.Number())
		return nil
	}

	context.result = Bool(left.String() != right.String())
	return nil
}

func execRelationalExprLessThan(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, leftNodeSetOk := left.(NodeSet)
	rightNodeSet, rightNodeSetOk := right.(NodeSet)

	if leftNodeSetOk && rightNodeSetOk {
		for _, leftNode := range leftNodeSet {
			for _, rightNode := range rightNodeSet {
				if getCursorString(leftNode) < getCursorString(rightNode) {
					context.result = Bool(true)
					return nil
				}
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftNumber, leftNumberOk := left.(Number)

	if leftNumberOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftNumber < Number(getStringNumber(getCursorString(rightNode))) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightNumber, rightNumberOk := right.(Number)

	if leftNodeSetOk && rightNumberOk {
		for _, leftNode := range leftNodeSet {
			if Number(getStringNumber(getCursorString(leftNode))) < rightNumber {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftString, leftStringOk := left.(String)

	if leftStringOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftString < String(getCursorString(rightNode)) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightString, rightStringOk := right.(String)

	if leftNodeSetOk && rightStringOk {
		for _, leftNode := range leftNodeSet {
			if String(getCursorString(leftNode)) < rightString {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	context.result = Bool(left.Number() < right.Number())
	return nil
}

func execRelationalExprLessThanOrEqual(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, leftNodeSetOk := left.(NodeSet)
	rightNodeSet, rightNodeSetOk := right.(NodeSet)

	if leftNodeSetOk && rightNodeSetOk {
		for _, leftNode := range leftNodeSet {
			for _, rightNode := range rightNodeSet {
				if getCursorString(leftNode) <= getCursorString(rightNode) {
					context.result = Bool(true)
					return nil
				}
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftNumber, leftNumberOk := left.(Number)

	if leftNumberOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftNumber <= Number(getStringNumber(getCursorString(rightNode))) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightNumber, rightNumberOk := right.(Number)

	if leftNodeSetOk && rightNumberOk {
		for _, leftNode := range leftNodeSet {
			if Number(getStringNumber(getCursorString(leftNode))) <= rightNumber {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftString, leftStringOk := left.(String)

	if leftStringOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftString <= String(getCursorString(rightNode)) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightString, rightStringOk := right.(String)

	if leftNodeSetOk && rightStringOk {
		for _, leftNode := range leftNodeSet {
			if String(getCursorString(leftNode)) <= rightString {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	context.result = Bool(left.Number() <= right.Number())
	return nil
}

func execRelationalExprGreaterThan(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, leftNodeSetOk := left.(NodeSet)
	rightNodeSet, rightNodeSetOk := right.(NodeSet)

	if leftNodeSetOk && rightNodeSetOk {
		for _, leftNode := range leftNodeSet {
			for _, rightNode := range rightNodeSet {
				if getCursorString(leftNode) > getCursorString(rightNode) {
					context.result = Bool(true)
					return nil
				}
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftNumber, leftNumberOk := left.(Number)

	if leftNumberOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftNumber > Number(getStringNumber(getCursorString(rightNode))) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightNumber, rightNumberOk := right.(Number)

	if leftNodeSetOk && rightNumberOk {
		for _, leftNode := range leftNodeSet {
			if Number(getStringNumber(getCursorString(leftNode))) > rightNumber {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftString, leftStringOk := left.(String)

	if leftStringOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftString > String(getCursorString(rightNode)) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightString, rightStringOk := right.(String)

	if leftNodeSetOk && rightStringOk {
		for _, leftNode := range leftNodeSet {
			if String(getCursorString(leftNode)) > rightString {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	context.result = Bool(left.Number() > right.Number())
	return nil
}

func execRelationalExprGreaterThanOrEqual(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, leftNodeSetOk := left.(NodeSet)
	rightNodeSet, rightNodeSetOk := right.(NodeSet)

	if leftNodeSetOk && rightNodeSetOk {
		for _, leftNode := range leftNodeSet {
			for _, rightNode := range rightNodeSet {
				if getCursorString(leftNode) >= getCursorString(rightNode) {
					context.result = Bool(true)
					return nil
				}
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftNumber, leftNumberOk := left.(Number)

	if leftNumberOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftNumber >= Number(getStringNumber(getCursorString(rightNode))) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightNumber, rightNumberOk := right.(Number)

	if leftNodeSetOk && rightNumberOk {
		for _, leftNode := range leftNodeSet {
			if Number(getStringNumber(getCursorString(leftNode))) >= rightNumber {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	leftString, leftStringOk := left.(String)

	if leftStringOk && rightNodeSetOk {
		for _, rightNode := range rightNodeSet {
			if leftString >= String(getCursorString(rightNode)) {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	rightString, rightStringOk := right.(String)

	if leftNodeSetOk && rightStringOk {
		for _, leftNode := range leftNodeSet {
			if String(getCursorString(leftNode)) >= rightString {
				context.result = Bool(true)
				return nil
			}
		}

		context.result = Bool(false)
		return nil
	}

	context.result = Bool(left.Number() >= right.Number())
	return nil
}

func execOrExprOr(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftBool := left.Bool()
	rightBool := right.Bool()

	context.result = Bool(leftBool || rightBool)
	return nil
}

func execAndExprAnd(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftBool := left.Bool()
	rightBool := right.Bool()

	context.result = Bool(leftBool && rightBool)
	return nil
}

func execAbsoluteLocationPathOnly(context *exprContext, expr *grammar.Grammar) error {
	context.result = NodeSet{context.root}
	return nil
}

func execStep(context *exprContext, expr *grammar.Grammar) error {
	children := make([]*bsr.BSR, 0, 1)

	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			children = append(children, &c)
		}
	}

	if len(children) != 1 {
		return fmt.Errorf("BSR list size not one: %+v", children)
	}

	nextBsr := children[0]
	nextSlotNt := nextBsr.Label.Slot().NT

	if nextSlotNt == symbols.NT_NodeTest ||
		nextSlotNt == symbols.NT_NodeTestAndPredicate ||
		nextSlotNt == symbols.NT_NodeTestNodeTypeNoArgTest ||
		nextSlotNt == symbols.NT_NodeTestNodeTypeNoArgTestNodeTestConflictResolver ||
		nextSlotNt == symbols.NT_NodeTestProcInstTargetTest ||
		nextSlotNt == symbols.NT_NameTestAnyElement ||
		nextSlotNt == symbols.NT_NameTestNamespaceAnyLocal ||
		nextSlotNt == symbols.NT_NameTestQNameNamespaceWithLocal ||
		nextSlotNt == symbols.NT_NameTestQNameLocalOnly {
		nodeSet, ok := context.result.(NodeSet)

		if !ok {
			return fmt.Errorf("cannot query nodes on non-NodeSet's")
		}

		context.result = selectChild(nodeSet)
	}

	return execContext(context, expr.Next(nextBsr))
}

func execPredicate(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return fmt.Errorf("cannot run path queries on non-NodeSet's")
	}

	nextResult := make(NodeSet, 0)

	for i := range nodeSet {
		nextContext := context.copy()
		nextContext.result = NodeSet{nodeSet[i]}
		nextContext.contextPosition = i
		left, err := leftOnlyIndependentResult(&nextContext, expr)

		if err != nil {
			return err
		}

		if n, ok := left.(Number); ok {
			if (i + 1) == int(n) {
				nextResult = append(nextResult, nodeSet[i])
			}
		} else if b, ok := left.(Bool); ok {
			if bool(b) {
				nextResult = append(nextResult, nodeSet[i])
			}
		} else if left.Bool() {
			nextResult = append(nextResult, nodeSet[i])
		}
	}

	context.result = nextResult
	return nil
}

func execUnionExprUnion(context *exprContext, expr *grammar.Grammar) error {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return err
	}

	leftNodeSet, lok := left.(NodeSet)
	rightNodeSet, rok := right.(NodeSet)

	if !lok || !rok {
		return fmt.Errorf("cannot union non-NodeSet's")
	}

	context.result = cleanup(append(leftNodeSet, rightNodeSet...))
	return nil
}

func execNodeTestNodeTypeNoArgTest(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return nil
	}

	nodeType := expr.GetString()
	parenIndex := strings.LastIndex(nodeType, "(")
	nodeType = nodeType[:parenIndex]
	nodeType = strings.TrimSpace(nodeType)

	result := make(NodeSet, 0)

	switch nodeType {
	case "comment":
		for _, i := range nodeSet {
			if _, ok := i.Node().(node.Comment); ok {
				result = append(result, i)
			}
		}
	case "text":
		for _, i := range nodeSet {
			if _, ok := i.Node().(node.CharData); ok {
				result = append(result, i)
			}
		}
	case "processing-instruction":
		for _, i := range nodeSet {
			if _, ok := i.Node().(node.ProcInst); ok {
				result = append(result, i)
			}
		}
	case "node":
		return nil
	}

	context.result = result
	return nil
}

func execNodeTestProcInstTargetTest(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return nil
	}

	literal, err := leftOnlyIndependentResult(context, expr)

	if err != nil {
		return err
	}

	literalString := literal.String()
	result := make(NodeSet, 0)

	for _, i := range nodeSet {
		if pi, ok := i.Node().(node.ProcInst); ok && pi.Target() == literalString {
			result = append(result, i)
		}
	}

	context.result = result
	return nil
}

func execNameTestAnyElement(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return nil
	}

	result := make(NodeSet, 0)

	for _, i := range nodeSet {
		if _, ok := i.Node().(node.NamedNode); ok {
			result = append(result, i)
		}

		if _, ok := i.Node().(node.Namespace); ok {
			result = append(result, i)
		}
	}

	context.result = result
	return nil
}

func execNameTestNamespaceAnyLocal(context *exprContext, expr *grammar.Grammar) error {
	namespaceLookup := expr.BSR.GetTChildI(0).LiteralString()
	namespaceValue, ok := context.NamespaceDecls[namespaceLookup]

	if !ok {
		return fmt.Errorf("unknown namespace binding '%s'", namespaceLookup)
	}

	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return nil
	}

	result := make(NodeSet, 0)

	for _, i := range nodeSet {
		if node, ok := i.Node().(node.NamedNode); ok {
			if node.Space() == namespaceValue {
				result = append(result, i)
			}
		}
	}

	context.result = result
	return nil
}

func execNameTestLocalAnyNamespace(context *exprContext, expr *grammar.Grammar) error {
	localValue := expr.BSR.GetTChildI(2).LiteralString()
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return nil
	}

	result := make(NodeSet, 0)

	for _, i := range nodeSet {
		if node, ok := i.Node().(node.NamedNode); ok {
			if node.Local() == localValue {
				result = append(result, i)
			}
		}
	}

	context.result = result
	return nil
}

func execNameTestQNameNamespaceWithLocal(context *exprContext, expr *grammar.Grammar) error {
	namespaceLookup := expr.BSR.GetTChildI(0).LiteralString()
	namespaceValue := context.NamespaceDecls[namespaceLookup]
	local := expr.BSR.GetTChildI(2).LiteralString()

	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return nil
	}

	result := make(NodeSet, 0)

	for _, i := range nodeSet {
		if node, ok := i.Node().(node.NamedNode); ok {
			if node.Local() == local && node.Space() == namespaceValue {
				result = append(result, i)
			}
		}
	}

	context.result = result
	return nil
}

func execNameTestQNameLocalOnly(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return fmt.Errorf("cannot run path queries on non-NodeSet's")
	}

	nextResult := make(NodeSet, 0)
	queryName := expr.GetString()

	for _, child := range nodeSet {
		if elem, ok := child.Node().(node.NamedNode); ok {
			if elem.Space() == "" && elem.Local() == queryName {
				nextResult = append(nextResult, child)
			}
		}

		if ns, ok := child.Node().(node.Namespace); ok {
			namespaceValue := context.NamespaceDecls[queryName]

			if ns.NamespaceValue() == namespaceValue {
				nextResult = append(nextResult, child)
			}
		}
	}

	context.result = nextResult

	return execChildren(context, expr)
}

func execAxisName(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return fmt.Errorf("cannot query nodes on non-NodeSet's")
	}

	axis := expr.GetString()
	var result Result

	switch axis {
	case "child":
		result = selectChild(nodeSet)
	case "attribute":
		result = selectAttributes(nodeSet)
	case "ancestor":
		result = selectAncestor(nodeSet)
	case "ancestor-or-self":
		result = selectAncestorOrSelf(nodeSet)
	case "descendant":
		result = selectDescendant(nodeSet)
	case "descendant-or-self":
		result = selectDescendantOrSelf(nodeSet)
	case "following":
		result = selectFollowing(nodeSet)
	case "following-sibling":
		result = selectFollowingSibling(nodeSet)
	case "namespace":
		result = selectNamespace(nodeSet)
	case "parent":
		result = selectParent(nodeSet)
	case "preceding":
		result = selectPreceding(nodeSet)
	case "preceding-sibling":
		result = selectPrecedingSibling(nodeSet)
	default: // self
		return nil
	}

	context.result = result

	return nil
}

func execAbbreviatedStepParent(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return fmt.Errorf("cannot query nodes on non-NodeSet's")
	}

	context.result = selectParent(nodeSet)
	return nil
}

func execAbbreviatedAxisSpecifier(context *exprContext, expr *grammar.Grammar) error {
	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return fmt.Errorf("cannot query nodes on non-NodeSet's")
	}

	context.result = selectAttributes(nodeSet)
	return nil
}

func execAbbreviatedAbsoluteLocationPath(context *exprContext, expr *grammar.Grammar) error {
	context.result = selectDescendantOrSelf(NodeSet{context.root})

	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			return execContext(context, expr.Next(&c))
		}
	}

	return nil
}

func execAbbreviatedRelativeLocationPath(context *exprContext, expr *grammar.Grammar) error {
	children := make([]*bsr.BSR, 0, 2)

	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			children = append(children, &c)
		}
	}

	if len(children) != 2 {
		return fmt.Errorf("BSR list size not two: %+v", children)
	}

	if err := execContext(context, expr.Next(children[0])); err != nil {
		return err
	}

	nodeSet, ok := context.result.(NodeSet)

	if !ok {
		return fmt.Errorf("cannot query nodes on non-NodeSet's")
	}

	context.result = selectDescendantOrSelf(nodeSet)
	return execContext(context, expr.Next(children[1]))
}

func execFunctionCall(context *exprContext, expr *grammar.Grammar) error {
	children := make([]*bsr.BSR, 0, 2)

	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			children = append(children, &c)
		}
	}

	if len(children) != 2 {
		return fmt.Errorf("BSR list size not two: %+v", children)
	}

	bsrs := make([]*bsr.BSR, 0)
	gatherFunctionArgs(children[1], &bsrs)

	args := make([]Result, 0, len(bsrs))

	for _, i := range bsrs {
		nextContext := context.copy()
		nextExpr := expr.Next(i)

		if err := execContext(&nextContext, nextExpr); err != nil {
			return err
		}

		args = append(args, nextContext.result)
	}

	qname, err := GetQName(expr.Next(children[0]).GetString(), context.NamespaceDecls)

	if err != nil {
		return err
	}

	fn := context.FunctionLibrary[qname]

	if fn == nil {
		fn = context.builtinFunctions[qname]
	}

	if fn == nil {
		return fmt.Errorf("could not find function %s:%s", qname.Space, qname.Local)
	}

	result, err := fn(context, args...)

	if err != nil {
		return fmt.Errorf("error invoking function %s:%s: %s", qname.Space, qname.Local, err)
	}

	context.result = result

	return nil
}

func gatherFunctionArgs(b *bsr.BSR, args *[]*bsr.BSR) {
	children := make([]*bsr.BSR, 0, 2)

	for _, cn := range b.GetAllNTChildren() {
		for _, c := range cn {
			children = append(children, &c)
		}
	}

	name := b.Label.Slot().NT

	if (name == symbols.NT_FunctionSignature || name == symbols.NT_FunctionCallArgumentList) && len(children) == 1 {
		gatherFunctionArgs(children[0], args)
		return
	}

	if len(children) >= 1 {
		*args = append(*args, children[0])
	}

	if len(children) == 2 {
		gatherFunctionArgs(children[1], args)
	}
}

func execVariableReference(context *exprContext, expr *grammar.Grammar) error {
	variableStr := expr.Next(expr.BSR).GetString()
	variableStr = strings.TrimSpace(variableStr)

	if strings.HasPrefix(variableStr, "$") {
		variableStr = variableStr[1:]
	}

	qname, err := GetQName(variableStr, context.NamespaceDecls)

	if err != nil {
		return err
	}

	variable := context.Variables[qname]

	if variable == nil {
		return fmt.Errorf("could not find variable %s:%s", qname.Space, qname.Local)
	}

	context.result = variable
	return nil
}
