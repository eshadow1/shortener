package reset

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ParsedStruct хранит информацию о структуре, требующей генерации
type ParsedStruct struct {
	TypeName string
	Receiver string
	Fields   []*ast.Field
}

// ProcessFile парсит один Go-файл и возвращает найденные структуры для генерации
func ProcessFile(path string) ([]ParsedStruct, string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, "", err
	}

	var structs []ParsedStruct

	for _, decl := range file.Decls {
		genDecl, okDecl := decl.(*ast.GenDecl)
		if !okDecl || genDecl.Tok != token.TYPE {
			continue
		}

		hasComment := hasGenerateResetComment(genDecl.Doc)

		for _, spec := range genDecl.Specs {
			typeSpec, okSpec := spec.(*ast.TypeSpec)
			if !okSpec {
				continue
			}

			if !hasComment {
				hasComment = hasGenerateResetComment(typeSpec.Comment)
			}

			if !hasComment {
				continue
			}

			structType, okType := typeSpec.Type.(*ast.StructType)
			if !okType {
				continue
			}

			typeName := typeSpec.Name.Name
			receiver := strings.ToLower(typeName[:1])
			if receiver == typeName {
				receiver = "x"
			}

			structs = append(structs, ParsedStruct{
				TypeName: typeName,
				Receiver: receiver,
				Fields:   structType.Fields.List,
			})
		}
	}

	return structs, file.Name.Name, nil
}

func hasGenerateResetComment(group *ast.CommentGroup) bool {
	if group == nil {
		return false
	}
	for _, comment := range group.List {
		if strings.Contains(comment.Text, "generate:reset") {
			return true
		}
	}
	return false
}
