package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	protogen.Options{}.Run(func(p *protogen.Plugin) error {
		p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		return generate(p)
	})
}
