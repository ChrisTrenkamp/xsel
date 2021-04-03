package exec

import (
	"fmt"

	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/grammar/parser/bsr"
)

func leftOnlyIndependentResult(context *exprContext, expr *grammar.Grammar) (Result, error) {
	children := make([]*bsr.BSR, 0, 1)

	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			children = append(children, &c)
		}
	}

	if len(children) != 1 {
		return nil, fmt.Errorf("BSR list size not one: %+v", children)
	}

	left := context.copy()

	if err := execContext(&left, expr.Next(children[0])); err != nil {
		return nil, err
	}

	return left.result, nil
}

func leftRightIndependentResult(context *exprContext, expr *grammar.Grammar) (Result, Result, error) {
	children := make([]*bsr.BSR, 0, 2)

	for _, cn := range expr.BSR.GetAllNTChildren() {
		for _, c := range cn {
			children = append(children, &c)
		}
	}

	if len(children) != 2 {
		return nil, nil, fmt.Errorf("BSR list size not two: %+v", children)
	}

	left := context.copy()
	right := context.copy()

	if err := execContext(&left, expr.Next(children[0])); err != nil {
		return nil, nil, err
	}

	if err := execContext(&right, expr.Next(children[1])); err != nil {
		return nil, nil, err
	}

	return left.result, right.result, nil
}

func leftRightIndependentNumber(context *exprContext, expr *grammar.Grammar) (float64, float64, error) {
	left, right, err := leftRightIndependentResult(context, expr)

	if err != nil {
		return 0, 0, err
	}

	leftNum := left.Number()
	rightNum := right.Number()

	return leftNum, rightNum, nil
}

func leftRightDependentResult(context *exprContext, expr *grammar.Grammar) error {
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

	return execContext(context, expr.Next(children[1]))
}
