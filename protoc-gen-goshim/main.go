package main

import (
	"flag"
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	var flags flag.FlagSet
	old := flags.String("old", "", "old path")
	new := flags.String("new", "", "new path")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if f.Generate {
				if *old == "" || *new == "" {
					return fmt.Errorf("must specify old and new")
				}
				genShim(gen, f, *old, *new)
			}
		}
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL | pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
		return nil
	})
}

func genShim(gen *protogen.Plugin, file *protogen.File, old string, new string) {
	importPath := strings.ReplaceAll(string(file.GoImportPath), old, new)

	filename := file.GeneratedFilenamePrefix + ".shim.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-goshim. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P(`import slim "`, importPath, `"`)

	for _, m := range file.Messages {
		doMessage(g, m)
	}

	for _, m := range file.Enums {
		doEnum(g, m)
	}
}

func doMessage(g *protogen.GeneratedFile, m *protogen.Message) {
	g.P(fmt.Sprintf("type %[1]s = slim.%[1]s", m.GoIdent.GoName))
	for _, e := range m.Enums {
		doEnum(g, e)
	}

	for _, o := range m.Oneofs {
		// protobuf 3 optional like oneof
		if !o.Desc.IsSynthetic() {
			for _, f := range o.Fields {
				g.P(fmt.Sprintf("type %[1]s = slim.%[1]s", f.GoIdent.GoName))
			}
		}
	}
	for _, m := range m.Messages {
		doMessage(g, m)
	}
}

func doEnum(g *protogen.GeneratedFile, m *protogen.Enum) {
	g.P(fmt.Sprintf("type %[1]s = slim.%[1]s", m.GoIdent.GoName))
	g.P("const (")
	for _, v := range m.Values {
		g.P(fmt.Sprintf("%[1]s = slim.%[1]s", v.GoIdent.GoName))
	}
	g.P(")")
}
