package main

import (
	"fmt"
	"go/types"
)

type Expr string

type ConversionState struct {
	Type types.Type
	Expr Expr
}

type Conversion interface {
	Matches(t types.Type) bool
	Apply(ConversionState) ConversionState
}

type Converter struct {
	Rules []Conversion
}

func (c Converter) next(state ConversionState) []ConversionState {
	results := []ConversionState{}
	for _, rule := range c.Rules {
		if rule.Matches(state.Type) {
			results = append(results, rule.Apply(state))
		}
	}
	return results
}

func (c Converter) Convert(from ConversionState, to types.Type) (Expr, error) {
	maxDepth := 20

	visited := map[ConversionState]bool{}
	stack := []ConversionState{from}
	for range maxDepth {
		if len(stack) == 0 {
			return "", fmt.Errorf("stack is empty")
		}

		state := stack[0]
		stack = stack[1:]

		if visited[state] {
			continue
		}
		if types.Identical(state.Type, to) {
			return state.Expr, nil
		}

		visited[state] = true
		stack = append(stack, c.next(state)...)
	}
	return "", fmt.Errorf("max depth reached")
}

// BasicToPtr converts a basic type to a pointer.
type BasicToPtr struct{}

func (c BasicToPtr) Matches(t types.Type) bool {
	_, basic := t.Underlying().(*types.Basic)
	_, strct := t.Underlying().(*types.Struct)
	return basic || strct
}

func (c BasicToPtr) Apply(state ConversionState) ConversionState {
	return ConversionState{Type: types.NewPointer(state.Type), Expr: Expr(fmt.Sprintf("&%s", state.Expr))}
}

// PtrToBasic converts a pointer to a basic type.
type PtrToBasic struct{}

func (c PtrToBasic) Matches(t types.Type) bool {
	ptr, isPointer := t.Underlying().(*types.Pointer)
	if !isPointer {
		return false
	}
	_, basic := ptr.Elem().Underlying().(*types.Basic)
	_, strct := ptr.Elem().Underlying().(*types.Struct)
	return basic || strct
}

func (c PtrToBasic) Apply(state ConversionState) ConversionState {
	ptr, _ := state.Type.Underlying().(*types.Pointer)

	return ConversionState{Type: ptr.Elem(), Expr: Expr(fmt.Sprintf("*%s", state.Expr))}
}
