// Package main implements a code generator that creates Reset() methods
// for structures marked with the "// generate:reset" comment.
//
// The generator scans all packages in the project, finds structures with
// the special comment, and generates Reset() methods that restore objects
// to their default state. Generated methods are placed in reset.gen.go
// files in the corresponding packages.
//
// Usage:
//
//	go run cmd/reset/main.go
//
// Reset Rules:
//   - Basic types (int, string, bool, etc.) are reset to their zero values
//   - Slices are truncated to length 0 (not nil)
//   - Maps are cleared using the clear() function
//   - Nested structures with Reset() methods have that method called
//   - Non-nil pointers have their values reset according to these rules
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// StructInfo holds information about a structure that needs a Reset method.
type StructInfo struct {
	Name       string
	Fields     []FieldInfo
	PackagName string
	PackagPath string
}

// FieldInfo holds information about a structure field.
type FieldInfo struct {
	Name string
	Type ast.Expr
}

func main() {
	log.Println("Starting reset code generator...")

	// Find the project root (where go.mod is located)
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Failed to find project root: %v", err)
	}

	log.Printf("Project root: %s", projectRoot)

	// Scan all packages and find structures with generate:reset comment
	structsByPackage, err := scanPackages(projectRoot)
	if err != nil {
		log.Fatalf("Failed to scan packages: %v", err)
	}

	log.Printf("Found %d packages with resetable structures", len(structsByPackage))

	// Generate Reset methods for each package
	for pkgPath, structs := range structsByPackage {
		log.Printf("Generating reset methods for package: %s (%d structs)", pkgPath, len(structs))
		if err := generateResetFile(projectRoot, pkgPath, structs); err != nil {
			log.Fatalf("Failed to generate reset file for package %s: %v", pkgPath, err)
		}
	}

	log.Println("Reset code generation completed successfully!")
}

// findProjectRoot finds the project root by looking for go.mod file.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// scanPackages scans all packages in the project and finds structures with generate:reset comment.
func scanPackages(projectRoot string) (map[string][]StructInfo, error) {
	structsByPackage := make(map[string][]StructInfo)
	fset := token.NewFileSet()

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and .git directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git" || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}

		// Only process .go files (skip test and generated files)
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), "_test.go") &&
			!strings.HasSuffix(info.Name(), ".gen.go") {

			structs, pkgName, err := parseFileForResetStructs(fset, path)
			if err != nil {
				log.Printf("Warning: failed to parse %s: %v", path, err)
				return nil
			}

			if len(structs) > 0 {
				pkgPath := filepath.Dir(path)
				relPath, _ := filepath.Rel(projectRoot, pkgPath)

				for i := range structs {
					structs[i].PackagName = pkgName
					structs[i].PackagPath = relPath
				}

				structsByPackage[relPath] = append(structsByPackage[relPath], structs...)
			}
		}

		return nil
	})

	return structsByPackage, err
}

// parseFileForResetStructs parses a Go file and extracts structures with generate:reset comment.
func parseFileForResetStructs(fset *token.FileSet, filePath string) ([]StructInfo, string, error) {
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, "", err
	}

	var structs []StructInfo
	pkgName := file.Name.Name

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// Check if the declaration has the generate:reset comment
		hasResetComment := false
		if genDecl.Doc != nil {
			for _, comment := range genDecl.Doc.List {
				if strings.Contains(comment.Text, "generate:reset") {
					hasResetComment = true
					break
				}
			}
		}

		if !hasResetComment {
			continue
		}

		// Extract struct information
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfo := StructInfo{
				Name:   typeSpec.Name.Name,
				Fields: extractFields(structType),
			}

			structs = append(structs, structInfo)
		}
	}

	return structs, pkgName, nil
}

// extractFields extracts field information from a struct type.
func extractFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fields = append(fields, FieldInfo{
				Name: name.Name,
				Type: field.Type,
			})
		}
	}

	return fields
}

// generateResetFile generates a reset.gen.go file for a package.
func generateResetFile(projectRoot, pkgPath string, structs []StructInfo) error {
	if len(structs) == 0 {
		return nil
	}

	var buf bytes.Buffer

	// Write package header
	pkgName := structs[0].PackagName
	buf.WriteString("// Code generated by cmd/reset. DO NOT EDIT.\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Generate Reset method for each struct
	for _, structInfo := range structs {
		buf.WriteString(generateResetMethod(structInfo))
		buf.WriteString("\n")
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated code: %v\n%s", err, buf.String())
	}

	// Write to file
	outputPath := filepath.Join(projectRoot, pkgPath, "reset.gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("Generated: %s", outputPath)
	return nil
}

// generateResetMethod generates a Reset() method for a struct.
func generateResetMethod(structInfo StructInfo) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("// Reset resets %s to its default state.\n", structInfo.Name))
	buf.WriteString(fmt.Sprintf("func (r *%s) Reset() {\n", structInfo.Name))
	buf.WriteString("\tif r == nil {\n")
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n\n")

	for _, field := range structInfo.Fields {
		resetCode := generateFieldReset(field)
		if resetCode != "" {
			buf.WriteString("\t" + resetCode + "\n")
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}

// generateFieldReset generates reset code for a field based on its type.
func generateFieldReset(field FieldInfo) string {
	switch typ := field.Type.(type) {
	case *ast.Ident:
		// Basic types
		return generateBasicTypeReset(field.Name, typ.Name)

	case *ast.StarExpr:
		// Pointer types
		return generatePointerReset(field.Name, typ)

	case *ast.ArrayType:
		// Slice or array types
		if typ.Len == nil {
			// It's a slice
			return fmt.Sprintf("r.%s = r.%s[:0]", field.Name, field.Name)
		}
		// Arrays - reset elements if needed
		return ""

	case *ast.MapType:
		// Map types
		return fmt.Sprintf("clear(r.%s)", field.Name)

	case *ast.SelectorExpr:
		// Qualified types (e.g., time.Time, cache.Cache)
		return generateQualifiedTypeReset(field.Name, typ)

	default:
		// For other types, try to call Reset() if available
		return generateInterfaceReset(field.Name)
	}
}

// generateBasicTypeReset generates reset code for basic types.
func generateBasicTypeReset(fieldName, typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"byte", "rune":
		return fmt.Sprintf("r.%s = 0", fieldName)
	case "float32", "float64":
		return fmt.Sprintf("r.%s = 0", fieldName)
	case "string":
		return fmt.Sprintf("r.%s = \"\"", fieldName)
	case "bool":
		return fmt.Sprintf("r.%s = false", fieldName)
	default:
		// For custom types, try to call Reset() if available
		return generateInterfaceReset(fieldName)
	}
}

// generatePointerReset generates reset code for pointer types.
func generatePointerReset(fieldName string, starExpr *ast.StarExpr) string {
	var buf bytes.Buffer

	// Check if pointer is not nil
	buf.WriteString(fmt.Sprintf("if r.%s != nil {\n", fieldName))

	// Determine the underlying type
	switch typ := starExpr.X.(type) {
	case *ast.Ident:
		// Pointer to basic type or same-package type
		resetValue := getZeroValue(typ.Name)
		if resetValue != "" {
			buf.WriteString(fmt.Sprintf("\t\t*r.%s = %s\n", fieldName, resetValue))
		} else {
			// Pointer to custom type - try to call Reset() method directly
			buf.WriteString(fmt.Sprintf("\t\tr.%s.Reset()\n", fieldName))
		}

	case *ast.SelectorExpr:
		// Pointer to external package type (e.g., *cache.Cache)
		resetCode := generateExternalTypeReset(fieldName, typ)
		if resetCode != "" {
			buf.WriteString("\t\t" + resetCode + "\n")
		}

	case *ast.StarExpr:
		// Pointer to pointer - try to call Reset()
		buf.WriteString(fmt.Sprintf("\t\tr.%s.Reset()\n", fieldName))

	default:
		// For other types, try to call Reset() method directly
		buf.WriteString(fmt.Sprintf("\t\tr.%s.Reset()\n", fieldName))
	}

	buf.WriteString("\t}")
	return buf.String()
}

// generateExternalTypeReset generates reset code for external package types.
func generateExternalTypeReset(fieldName string, selectorExpr *ast.SelectorExpr) string {
	// Get package and type name
	var pkgName, typeName string
	if pkg, ok := selectorExpr.X.(*ast.Ident); ok {
		pkgName = pkg.Name
		typeName = selectorExpr.Sel.Name
	}

	// Handle known external types
	if pkgName == "cache" && typeName == "Cache" {
		// *cache.Cache should be flushed
		return fmt.Sprintf("r.%s.Flush()", fieldName)
	}

	// For unknown external types, try to call Reset() if available
	return fmt.Sprintf("if resetter, ok := interface{}(r.%s).(interface{ Reset() }); ok {\n\t\t\tresetter.Reset()\n\t\t}", fieldName)
}

// generateQualifiedTypeReset generates reset code for qualified types (e.g., time.Time).
func generateQualifiedTypeReset(fieldName string, selectorExpr *ast.SelectorExpr) string {
	// For qualified types, try to use their zero value or Reset method
	return generateInterfaceReset(fieldName)
}

// generateInterfaceReset generates reset code that tries to call Reset() method.
func generateInterfaceReset(fieldName string) string {
	return fmt.Sprintf("if resetter, ok := interface{}(r.%s).(interface{ Reset() }); ok {\n\t\tresetter.Reset()\n\t}", fieldName)
}

// getZeroValue returns the zero value for a basic type.
func getZeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"byte", "rune", "float32", "float64":
		return "0"
	case "string":
		return "\"\""
	case "bool":
		return "false"
	default:
		return ""
	}
}
