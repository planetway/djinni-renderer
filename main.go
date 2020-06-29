package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/SafetyCulture/djinni-parser/pkg/ast"
	"github.com/SafetyCulture/djinni-parser/pkg/parser"
)

func usage() {
	log.Printf("usage: %s path/to/file.djinni path/to/template", os.Args[0])
	os.Exit(-1)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	src := os.Args[1]
	f, err := parser.ParseFile(src, nil)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	tmplname := os.Args[2]
	b, err := ioutil.ReadFile(tmplname)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	t := template.Must(template.New("").Funcs(funcs()).Parse(string(b)))
	if err := t.Execute(os.Stdout, f); err != nil {
		log.Printf("execute: %s", err)
	}
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"DeclType": func(value interface{}) string {
			switch value.(type) {
			case *ast.Record:
				return "record"
			case *ast.Enum:
				return "enum"
			case *ast.Interface:
				return "interface"
			default:
				return ""
			}
		},
		"TypeDeclsMatchString": func(typeDecls []ast.TypeDecl, match string) []ast.TypeDecl {
			var ret []ast.TypeDecl
			for _, v := range typeDecls {
				if strings.Contains(v.Ident.Name, match) {
					ret = append(ret, v)
				}
			}
			return ret
		},
		"RecordFields": func(value interface{}) []ast.Field {
			record := value.(*ast.Record)
			return record.Fields
		},
		"RecordConstants": func(value interface{}) []ast.Const {
			record := value.(*ast.Record)
			return record.Consts
		},
		"EnumOptions": func(value interface{}) []ast.EnumOption {
			e := value.(*ast.Enum)
			return e.Options
		},
		"FullType": fullType,
		"CppType":  cppType,
		"ConstString": func(cnst ast.Const) string {
			if cnst.Type.Ident.Name == "string" {
				return fmt.Sprintf(`"%s"`, cnst.Value.(string))
			} else {
				// TODO
				return "# to be implemented"
			}
		},
		"IsCustomType": func(typeExpr ast.TypeExpr) bool {
			return ast.IsCustomType(typeExpr)
		},
		"TitleCase": strings.Title,
		"ToUpper":   strings.ToUpper,
		"Dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"Except": func(array []ast.TypeDecl, exceptions ...string) []ast.TypeDecl {
			ret := []ast.TypeDecl{}
			for _, v := range array {
				found := false
				for _, vv := range exceptions {
					if v.Ident.Name == vv {
						found = true
						break
					}
				}
				if !found {
					ret = append(ret, v)
				}
			}
			return ret
		},
	}
}

func fullType(typeExpr ast.TypeExpr) string {
	name := typeExpr.Ident.Name
	if name == "map" {
		return fmt.Sprintf("map<%s, %s>", typeExpr.Args[0].Ident.Name, typeExpr.Args[1].Ident.Name)
	} else if name == "set" || name == "list" || name == "optional" {
		// generic types can be recursive, ex: optional<list<string>>
		return fmt.Sprintf("%s<%s>", name, fullType(typeExpr.Args[0]))
	} else {
		return name
	}
}

func cppType(typeExpr ast.TypeExpr, customTypePrefix string) string {
	name := typeExpr.Ident.Name
	switch name {
	case "map":
		return fmt.Sprintf("std::unordered_map<%s, %s>", cppType(typeExpr.Args[0], customTypePrefix), cppType(typeExpr.Args[1], customTypePrefix))
	case "set":
		return fmt.Sprintf("std::unordered_set<%s>", cppType(typeExpr.Args[0], customTypePrefix))
	case "list":
		return fmt.Sprintf("std::vector<%s>", cppType(typeExpr.Args[0], customTypePrefix))
	case "optional":
		return fmt.Sprintf("std::optional<%s>", cppType(typeExpr.Args[0], customTypePrefix))
	case "string":
		return "std::string"
	case "i32":
		return "int32_t"
	case "i64":
		return "int64_t"
	case "bool":
		return "bool"
	case "binary":
		return "std::vector<uint8_t>"
	case "date":
		return "std::chrono::system_clock::time_point"
	default:
		return fmt.Sprintf("%s%s", customTypePrefix, strings.Title(name))
	}
}
