// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"go/types"
	"log"
	"reflect"
	"strings"

	"github.com/go-openapi/spec"
	"golang.org/x/tools/go/packages"
)

// setPropertyOrderFromGoStructs walks all definitions in the spec and
// sets x-order extensions on their properties to match the field
// declaration order of the corresponding Go struct.
//
// It uses x-go-package and x-go-name extensions (set by codescan)
// to locate the Go type for each definition.
func setPropertyOrderFromGoStructs(swspec *spec.Swagger, pkgPatterns []string, workDir, buildTags string) {
	structMap := loadGoStructs(pkgPatterns, workDir, buildTags)
	if len(structMap) == 0 {
		return
	}

	for defName, schema := range swspec.Definitions {
		st := findStructForSchema(&schema, defName, structMap)
		if st == nil {
			continue
		}

		setXOrderFromStruct(st, &schema)
		swspec.Definitions[defName] = schema
	}
}

// structKey is "pkgPath.TypeName" → *types.Struct.
type structMap map[string]*types.Struct

func loadGoStructs(pkgPatterns []string, workDir, buildTags string) structMap {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedName | packages.NeedImports |
			packages.NeedDeps,
		Dir: workDir,
	}
	if buildTags != "" {
		cfg.BuildFlags = []string{"-tags=" + buildTags}
	}

	pkgs, err := packages.Load(cfg, pkgPatterns...)
	if err != nil {
		log.Printf("warning: could not load Go packages for property ordering: %v", err)
		return nil
	}

	result := make(structMap)

	seen := make(map[string]bool)
	var walk func(*packages.Package)
	walk = func(pkg *packages.Package) {
		if seen[pkg.PkgPath] {
			return
		}
		seen[pkg.PkgPath] = true

		if pkg.Types != nil {
			scope := pkg.Types.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				tn, ok := obj.(*types.TypeName)
				if !ok {
					continue
				}
				if st, ok := tn.Type().Underlying().(*types.Struct); ok {
					result[pkg.PkgPath+"."+name] = st
				}
			}
		}
		for _, dep := range pkg.Imports {
			walk(dep)
		}
	}

	for _, pkg := range pkgs {
		walk(pkg)
	}

	return result
}

func findStructForSchema(schema *spec.Schema, defName string, sm structMap) *types.Struct {
	pkgPath, _ := schema.Extensions.GetString("x-go-package")
	goName, ok := schema.Extensions.GetString("x-go-name")
	if !ok {
		goName = defName
	}

	if pkgPath != "" {
		if st, ok := sm[pkgPath+"."+goName]; ok {
			return st
		}
	}

	// fallback: try all packages
	for key, st := range sm {
		if strings.HasSuffix(key, "."+goName) {
			return st
		}
	}
	return nil
}

func setXOrderFromStruct(st *types.Struct, schema *spec.Schema) {
	if len(schema.Properties) == 0 {
		return
	}

	order := 0
	for i := range st.NumFields() {
		fld := st.Field(i)
		if fld.Embedded() || !fld.Exported() {
			continue
		}

		tag := st.Tag(i)
		propName := jsonPropertyName(fld.Name(), tag)
		if propName == "" || propName == "-" {
			continue
		}

		prop, exists := schema.Properties[propName]
		if !exists {
			continue
		}

		if prop.Extensions == nil {
			prop.Extensions = make(spec.Extensions)
		}
		prop.Extensions["x-order"] = float64(order)
		schema.Properties[propName] = prop
		order++
	}
}

func jsonPropertyName(fieldName, tag string) string {
	jsonTag := reflect.StructTag(tag).Get("json")
	if jsonTag == "" || jsonTag == "-" {
		if jsonTag == "-" {
			return "-"
		}
		return fieldName
	}

	name, _, _ := strings.Cut(jsonTag, ",")
	if name == "" {
		return fieldName
	}
	if name == "-" {
		return "-"
	}
	return name
}
