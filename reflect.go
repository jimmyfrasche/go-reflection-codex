package main

// "reflect"

import (
	"go/ast" //just for Type interface
	"go/build"
	"go/doc" //convient way to smush the ast and just grab names
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var topics = MustReflectDefs()

func MustReflectDefs() map[string][]byte {
	n, err := ReflectDefs()
	if err != nil {
		panic(err)
	}
	return n
}

func gofiles(fi os.FileInfo) bool {
	return strings.HasSuffix(fi.Name(), ".go")
}

func ReflectDefs() (names map[string][]byte, err error) {
	b, err := build.Import("reflect", "", 0)
	if err != nil {
		return nil, err
	}

	ts := token.NewFileSet()

	p, err := parser.ParseDir(ts, b.Dir, gofiles, 0)
	if err != nil {
		return nil, err
	}

	pkg := doc.New(p["reflect"], "reflect", doc.AllMethods)

	acc := []string{}
	acc = values(pkg.Consts, acc)
	acc = values(pkg.Vars, acc)
	acc = funcs(pkg.Funcs, "reflect.", acc)
	acc = types(pkg.Types, acc)

	names = map[string][]byte{}
	for _, name := range acc {
		names[name] = []byte(name)
	}
	return names, nil
}

func values(vals []*doc.Value, acc []string) []string {
	for _, v := range vals {
		for _, name := range v.Names {
			acc = append(acc, "reflect."+name)
		}
	}
	return acc
}

func funcs(funcs []*doc.Func, prefix string, acc []string) []string {
	for _, f := range funcs {
		acc = append(acc, prefix+f.Name)
	}
	return acc
}

func types(types []*doc.Type, acc []string) []string {
	for _, t := range types {
		if strings.HasSuffix(t.Name, "Error") {
			continue
		}
		acc = append(acc, "reflect."+t.Name)

		if t.Name == "Type" {
			//drill down to interface type
			s := t.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.InterfaceType)
			for _, m := range s.Methods.List {
				acc = append(acc, "Type."+m.Names[0].Name)
			}
		}
		acc = values(t.Consts, acc)
		acc = values(t.Vars, acc)
		acc = funcs(t.Funcs, "reflect.", acc)
		acc = funcs(t.Methods, t.Name+".", acc)
	}
	return acc
}
