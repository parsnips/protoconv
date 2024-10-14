package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type protoField struct {
	protoType string
	name      string
	fieldNum  int
	comment   []string
}

// convertFieldToProto converts a Go struct field to a protobuf message field with proper numbering.
func convertFieldToProto(field *ast.Field, fieldNum int) string {
	if len(field.Names) == 0 {
		return ""
	}
	name := field.Names[0].Name
	protoType := goTypeToProtoType(field.Type)
	comment := ""
	if field.Doc != nil {
		var lines []string
		for _, c := range field.Doc.List {
			lines = append(lines, c.Text)
		}
		comment = strings.Join(lines, "\n") + "\n"
	}
	return fmt.Sprintf("%s  %s %s = %d;", comment, protoType, name, fieldNum)
}

// goTypeToProtoType maps Go types to protobuf types.
func goTypeToProtoType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "int", "int32":
			return "int32"
		case "int64":
			return "int64"
		case "float32":
			return "float"
		case "float64":
			return "double"
		case "bool":
			return "bool"
		case "time.Time":
			return "google.protobuf.Timestamp"
		case "error":
			return "string"
		case "Batcher":
			return "Batch"
		default:
			return t.Name
		}
	case *ast.ArrayType:
		return "repeated " + goTypeToProtoType(t.Elt)
	case *ast.StarExpr:
		return goTypeToProtoType(t.X)
	case *ast.SelectorExpr:
		// Handle imported types (just need to know package.Type format)
		return fmt.Sprintf("%s.%s", t.X, t.Sel.Name)
	default:
		return "string" // default to string for unmapped complex types
	}
}

// parseFiles parses a list of Go files and translates structs to Protobuf messages.
func parseFiles(filenames []string) {
	fileSet := token.NewFileSet()

	for _, filename := range filenames {
		if strings.Contains(filename, "_test") {
			continue
		}
		node, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("// Protobuf definitions generated from %s\n\n", filename)

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if genDecl.Doc != nil {
					for _, comment := range genDecl.Doc.List {
						fmt.Println(comment.Text)
					}
				}

				fmt.Printf("message %s {\n", typeSpec.Name.Name)

				fieldNum := 1
				for _, field := range structType.Fields.List {
					fmt.Println(convertFieldToProto(field, fieldNum))
					fieldNum++
				}
				fmt.Println("}")
			}
		}
	}
}

func isStructType(typeSpec *ast.TypeSpec) bool {
	_, ok := typeSpec.Type.(*ast.StructType)
	return ok
}

func generateToProtoFunc(structName, protoName string) string {
	return fmt.Sprintf(`
// To%sProto converts %s to %s.
func To%sProto(s *%s) *%s {
	if s == nil {
		return nil
	}
	return &%s{
		// TODO: Add field mappings here.
	}
}
`, protoName, structName, protoName, protoName, structName, protoName, protoName)
}

func generateFromProtoFunc(protoName, structName string) string {
	return fmt.Sprintf(`
// From%sProto converts %s to %s.
func From%sProto(p *%s) *%s {
	if p == nil {
		return nil
	}
	return &%s{
		// TODO: Add field mappings here.
	}
}
`, protoName, protoName, structName, protoName, protoName, structName, structName)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename1.go> <filename2.go> ...")
		return
	}
	filenames := os.Args[1:]
	parseFiles(filenames)
}
