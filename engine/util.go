package engine

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Top level function
// Analytical expression and execution
// err is not nil if an error occurs (including arithmetic runtime errors)
func ParseAndExec(s string) (r int, err error) {
	toks, err := Parse(s)
	if err != nil {
		return 0, err
	}
	ast := NewAST(toks, s)
	if ast.Err != nil {
		return 0, ast.Err
	}
	ar := ast.ParseExpression()
	if ast.Err != nil {
		return 0, ast.Err
	}
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return ExprASTResult(ar), err
}

func ErrPos(s string, pos int) string {
	r := strings.Repeat("-", len(s)) + "\n"
	s += "\n"
	for i := 0; i < pos; i++ {
		s += " "
	}
	s += "^\n"
	return r + s + r
}

// the integer power of a number
func Pow(x float64, n float64) float64 {
	return math.Pow(x, n)
}

// Float64ToStr float64 -> string
func Float64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ExprASTResult is a Top level function
// AST traversal
// if an arithmetic runtime error occurs, a panic exception is thrown
func ExprASTResult(expr ExprAST) int {
	var l, r int
	switch expr.(type) {
	case BinaryExprAST:
		ast := expr.(BinaryExprAST)
		l = ExprASTResult(ast.Lhs)
		r = ExprASTResult(ast.Rhs)
		switch ast.Op {
		case "+":
			return l + r
		case "-":
			return l - r
		case "*":
			return l * r
		case "/":
			if r == 0 {
				panic(errors.New(
					fmt.Sprintf("violation of arithmetic specification: a division by zero in ExprASTResult: [%g/%g]",
						l,
						r)))
			}
			return l / r
		case "%":
			return l % r
		case "^":
			return l ^ r
		case ">>":
			return l >> r
		case "<<":
			return l << r
		case ">":
			if l > r {
				return 1
			} else {
				return 0
			}
		case "<":
			if l < r {
				return 1
			} else {
				return 0
			}
		case "&":
			return l & r
		default:

		}
	case NumberExprAST:
		return expr.(NumberExprAST).Val
	}

	return 0.0
}
