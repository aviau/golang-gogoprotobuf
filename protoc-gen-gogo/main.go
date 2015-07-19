// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2010 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// protoc-gen-go is a plugin for the Google protocol buffer compiler to generate
// Go code.  Run it by building this program and putting it in your path with
// the name
// 	protoc-gen-gogo
// That word 'gogo' at the end becomes part of the option string set for the
// protocol compiler, so once the protocol compiler (protoc) is installed
// you can run
// 	protoc --gogo_out=output_directory input_directory/file.proto
// to generate Go bindings for the protocol defined by file.proto.
// With that input, the output will be written to
// 	output_directory/file.pb.go
//
// The generated code is documented in the package comment for
// the library.
//
// See the README and documentation for protocol buffers to learn more:
// 	https://developers.google.com/protocol-buffers/
package main

import (
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"

	_ "github.com/gogo/protobuf/plugin/defaultcheck"
	_ "github.com/gogo/protobuf/plugin/description"
	_ "github.com/gogo/protobuf/plugin/embedcheck"
	_ "github.com/gogo/protobuf/plugin/enumstringer"
	_ "github.com/gogo/protobuf/plugin/equal"
	_ "github.com/gogo/protobuf/plugin/face"
	_ "github.com/gogo/protobuf/plugin/gostring"
	_ "github.com/gogo/protobuf/plugin/marshalto"
	_ "github.com/gogo/protobuf/plugin/populate"
	_ "github.com/gogo/protobuf/plugin/size"
	_ "github.com/gogo/protobuf/plugin/stringer"
	_ "github.com/gogo/protobuf/plugin/union"
	_ "github.com/gogo/protobuf/plugin/unmarshal"

	"github.com/gogo/protobuf/plugin/testgen"

	"go/format"
	"strings"
)

func main() {
	// Begin by allocating a generator. The request and response structures are stored there
	// so we can do error handling easily - the response structure contains the field to
	// report failure.
	g := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	g.CommandLineParameters(g.Request.GetParameter())

	// Create a wrapped version of the Descriptors and EnumDescriptors that
	// point to the file that defines them.
	g.WrapTypes()

	g.SetPackageNames()
	g.BuildTypeNameMap()

	g.GenerateAllFiles()

	gtest := generator.New()

	if err := proto.Unmarshal(data, gtest.Request); err != nil {
		gtest.Error(err, "parsing input proto")
	}

	if len(gtest.Request.FileToGenerate) == 0 {
		gtest.Fail("no files to generate")
	}

	gtest.CommandLineParameters(gtest.Request.GetParameter())

	// Create a wrapped version of the Descriptors and EnumDescriptors that
	// point to the file that defines them.
	gtest.WrapTypes()

	gtest.SetPackageNames()
	gtest.BuildTypeNameMap()

	gtest.GeneratePlugin(testgen.NewPlugin())

	for i := 0; i < len(gtest.Response.File); i++ {
		if strings.Contains(*gtest.Response.File[i].Content, `//These tests are generated by github.com/gogo/protobuf/plugin/testgen`) {
			gtest.Response.File[i].Name = proto.String(strings.Replace(*gtest.Response.File[i].Name, ".pb.go", "pb_test.go", -1))
			g.Response.File = append(g.Response.File, gtest.Response.File[i])
		}
	}

	for i := 0; i < len(g.Response.File); i++ {
		formatted, err := format.Source([]byte(g.Response.File[i].GetContent()))
		if err != nil {
			g.Error(err, "go format error")
		}
		fmts := string(formatted)
		g.Response.File[i].Content = &fmts
	}

	// Send back the results.
	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}
