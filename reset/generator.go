// Package reset генерирует функции очистки для произвольных структур.
package reset

import (
	"fmt"
	"go/ast"
	"strings"
)

const pointerResetTemplate = "if %s.%s != nil {\n" +
	"\tif resetter, ok := interface{}(%s.%s).(interface{ Reset() }); ok {\n" +
	"\t\tresetter.Reset()\n" +
	"\t}\n" +
	"}"

const valueResetTemplate = "if resetter, ok := interface{}(&%s.%s).(interface{ Reset() }); ok {\n" +
	"\tresetter.Reset()\n" +
	"}"

// GenerateResetMethod собирает итоговый текст метода Reset()
func GenerateResetMethod(typeName, receiver string, fields []*ast.Field) string {
	var bodyLines []string

	bodyLines = append(bodyLines,
		fmt.Sprintf("if %s == nil {", receiver),
		"\treturn",
		"}",
		"",
	)

	for _, field := range fields {
		if len(field.Names) == 0 {
			continue
		}
		fieldName := field.Names[0].Name
		resetStmt := generateResetStmt(receiver, fieldName, field.Type)
		if resetStmt != "" {
			bodyLines = append(bodyLines, resetStmt)
		}
	}

	return fmt.Sprintf("func (%s *%s) Reset() {\n%s\n}\n",
		receiver, typeName, strings.Join(bodyLines, "\n"))
}

func generateResetStmt(receiver, fieldName string, expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if isBasicType(t.Name) {
			return fmt.Sprintf("%s.%s = %s", receiver, fieldName, zeroValue(t.Name))
		}
		return fmt.Sprintf(valueResetTemplate, receiver, fieldName)
	case *ast.SelectorExpr:
		return fmt.Sprintf(valueResetTemplate, receiver, fieldName)
	case *ast.StarExpr:
		return handlePointer(receiver, fieldName, t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return fmt.Sprintf("%s.%s = %s.%s[:0]", receiver, fieldName, receiver, fieldName)
		}
	case *ast.MapType:
		return fmt.Sprintf("clear(%s.%s)", receiver, fieldName)
	}
	return ""
}

func handlePointer(receiver, fieldName string, innerExpr ast.Expr) string {
	switch t := innerExpr.(type) {
	case *ast.Ident:
		if isBasicType(t.Name) {
			return fmt.Sprintf("if %s.%s != nil {\n\t*%s.%s = %s\n}", receiver, fieldName, receiver, fieldName, zeroValue(t.Name))
		}
		return fmt.Sprintf(pointerResetTemplate, receiver, fieldName, receiver, fieldName)
	case *ast.SelectorExpr:
		return fmt.Sprintf(pointerResetTemplate, receiver, fieldName, receiver, fieldName)
	case *ast.ArrayType:
		if t.Len == nil {
			return fmt.Sprintf("if %s.%s != nil {\n\t*%s.%s = (*%s.%s)[:0]\n}", receiver, fieldName, receiver, fieldName, receiver, fieldName)
		}
	case *ast.MapType:
		return fmt.Sprintf("if %s.%s != nil {\n\tclear(*%s.%s)\n}", receiver, fieldName, receiver, fieldName)
	}
	return fmt.Sprintf("// unsupported pointer type for %s.%s", receiver, fieldName)
}

func isBasicType(name string) bool {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64", "complex64", "complex128",
		"bool", "string", "byte", "rune":
		return true
	}
	return false
}

func zeroValue(name string) string {
	switch name {
	case "bool":
		return "false"
	case "string":
		return `""`
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64", "complex64", "complex128", "byte", "rune":
		return "0"
	default:
		return "nil"
	}
}
