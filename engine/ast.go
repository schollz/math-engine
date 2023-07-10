package engine

import (
	"errors"
	"fmt"
	"strconv"
)

var precedence = map[string]int{"+": 90, "-": 90, "*": 100, "/": 100, "%": 100, ">": 70, "&": 60, "<": 70, ">>": 80, "<<": 80, "|": 40, "^": 50}

type ExprAST interface {
	toStr() string
}

type NumberExprAST struct {
	Val int
	Str string
}

type BinaryExprAST struct {
	Op string
	Lhs,
	Rhs ExprAST
}

type FunCallerExprAST struct {
	Name string
	Arg  []ExprAST
}

func (n NumberExprAST) toStr() string {
	return fmt.Sprintf(
		"NumberExprAST:%s",
		n.Str,
	)
}

func (b BinaryExprAST) toStr() string {
	return fmt.Sprintf(
		"BinaryExprAST: (%s %s %s)",
		b.Op,
		b.Lhs.toStr(),
		b.Rhs.toStr(),
	)
}

func (n FunCallerExprAST) toStr() string {
	return fmt.Sprintf(
		"FunCallerExprAST:%s",
		n.Name,
	)
}

type AST struct {
	Tokens []*Token

	source    string
	currTok   *Token
	currIndex int
	depth     int

	Err error
}

func NewAST(toks []*Token, s string) *AST {
	a := &AST{
		Tokens: toks,
		source: s,
	}
	if a.Tokens == nil || len(a.Tokens) == 0 {
		a.Err = errors.New("empty token")
	} else {
		a.currIndex = 0
		a.currTok = a.Tokens[0]
	}
	return a
}

func (a *AST) ParseExpression() ExprAST {
	a.depth++ // called depth
	lhs := a.parsePrimary()
	r := a.parseBinOpRHS(0, lhs)
	a.depth--
	if a.depth == 0 && a.currIndex != len(a.Tokens) && a.Err == nil {
		a.Err = errors.New(
			fmt.Sprintf("bad expression, reaching the end or missing the operator\n%s",
				ErrPos(a.source, a.currTok.Offset)))
	}
	return r
}

func (a *AST) getNextToken() *Token {
	a.currIndex++
	if a.currIndex < len(a.Tokens) {
		a.currTok = a.Tokens[a.currIndex]
		return a.currTok
	}
	return nil
}

func (a *AST) getTokPrecedence() int {
	if p, ok := precedence[a.currTok.Tok]; ok {
		return p
	}
	return -1
}

func (a *AST) parseNumber() NumberExprAST {
	f64, err := strconv.Atoi(a.currTok.Tok)
	if err != nil {
		a.Err = errors.New(
			fmt.Sprintf("%v\nwant '(' or '0-9' but get '%s'\n%s",
				err.Error(),
				a.currTok.Tok,
				ErrPos(a.source, a.currTok.Offset)))
		return NumberExprAST{}
	}
	n := NumberExprAST{
		Val: f64,
		Str: a.currTok.Tok,
	}
	a.getNextToken()
	return n
}

func (a *AST) parsePrimary() ExprAST {
	switch a.currTok.Type {
	case Literal:
		return a.parseNumber()
	case Operator:
		if a.currTok.Tok == "(" {
			t := a.getNextToken()
			if t == nil {
				a.Err = errors.New(
					fmt.Sprintf("want '(' or '0-9' but get EOF\n%s",
						ErrPos(a.source, a.currTok.Offset)))
				return nil
			}
			e := a.ParseExpression()
			if e == nil {
				return nil
			}
			if a.currTok.Tok != ")" {
				a.Err = errors.New(
					fmt.Sprintf("want ')' but get %s\n%s",
						a.currTok.Tok,
						ErrPos(a.source, a.currTok.Offset)))
				return nil
			}
			a.getNextToken()
			return e
		} else if a.currTok.Tok == "-" {
			if a.getNextToken() == nil {
				a.Err = errors.New(
					fmt.Sprintf("want '0-9' but get '-'\n%s",
						ErrPos(a.source, a.currTok.Offset)))
				return nil
			}
			bin := BinaryExprAST{
				Op:  "-",
				Lhs: NumberExprAST{},
				Rhs: a.parsePrimary(),
			}
			return bin
		} else {
			return a.parseNumber()
		}
	case COMMA:
		a.Err = errors.New(
			fmt.Sprintf("want '(' or '0-9' but get %s\n%s",
				a.currTok.Tok,
				ErrPos(a.source, a.currTok.Offset)))
		return nil
	default:
		return nil
	}
}

func (a *AST) parseBinOpRHS(execPrec int, lhs ExprAST) ExprAST {
	for {
		tokPrec := a.getTokPrecedence()
		if tokPrec < execPrec {
			return lhs
		}
		binOp := a.currTok.Tok
		if a.getNextToken() == nil {
			a.Err = errors.New(
				fmt.Sprintf("want '(' or '0-9' but get EOF\n%s",
					ErrPos(a.source, a.currTok.Offset)))
			return nil
		}
		rhs := a.parsePrimary()
		if rhs == nil {
			return nil
		}
		nextPrec := a.getTokPrecedence()
		if tokPrec < nextPrec {
			rhs = a.parseBinOpRHS(tokPrec+1, rhs)
			if rhs == nil {
				return nil
			}
		}
		lhs = BinaryExprAST{
			Op:  binOp,
			Lhs: lhs,
			Rhs: rhs,
		}
	}
}
