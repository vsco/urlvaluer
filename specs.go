// Copyright 2014 Parker Moore
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go/ast"
	"log"
	"unicode"
	"unicode/utf8"
)

type StructSpec struct {
	Name   string
	Fields []fieldSpec
}

type fieldSpec struct {
	Name          string
	SnakeCaseName string
	resolvedType  string
	isPointer     bool
}

func NewFieldSpec(name, snakeCase string, astType ast.Expr) fieldSpec {
	spec := fieldSpec{Name: name, SnakeCaseName: snakeCase}
	return *spec.withType(astType)
}

func resolveType(astType ast.Expr) (resolvedType string, isPointer bool) {
	switch field := astType.(type) {
	case *ast.StarExpr:
		isPointer = true
		resolvedType, _ = resolveType(field.X)
	case *ast.Ident:
		if field.Obj != nil {
			if typeSpec, ok := field.Obj.Decl.(*ast.TypeSpec); ok {
				resolvedType, _ = resolveType(typeSpec.Type)
				log.Printf("resolved to %s", resolvedType)
			}
		} else {
			resolvedType = field.String()
		}
	case *ast.ArrayType:
		resolvedType = "array"
	case *ast.SelectorExpr:
		resolvedType, _ = resolveType(field.Sel)
	case *ast.StructType:
		resolvedType = "Struct"
	default:
		resolvedType = fmt.Sprintf("%T", astType)
	}
	return
}

func (f *fieldSpec) withType(astType ast.Expr) *fieldSpec {
	f.resolvedType, f.isPointer = resolveType(astType)
	return f
}

func (f fieldSpec) Accessor(owner string) string {
	if f.isPointer {
		if f.IsStruct() {
			return fmt.Sprintf(`%s.%s.UrlValues("%s")`, owner, f.Name, f.SnakeCaseName)
		}
		return fmt.Sprintf("*%s.%s", owner, f.Name)
	} else {
		return fmt.Sprintf("%s.%s", owner, f.Name)
	}
}

func (f fieldSpec) IsStruct() bool {
	r, _ := utf8.DecodeRuneInString(f.resolvedType[0:1])
	return unicode.IsUpper(r)
}

func (f fieldSpec) Zero() string {
	log.Printf("%v is a %v and isPointer=%t", f.Name, f.resolvedType, f.isPointer)
	if f.isPointer {
		return `nil`
	}

	switch f.resolvedType {
	case "string":
		return `""`
	case "int", "int64", "float", "float64":
		return `0`
	case "bool":
		return `false`
	default:
		return `0`
	}
}

func (f fieldSpec) HasLen() bool {
	return f.resolvedType == "array"
}
