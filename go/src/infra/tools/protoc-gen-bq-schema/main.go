package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"infra/libs/bqschema/tabledef"

	"github.com/golang/protobuf/proto"
	//	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	//"github.com/golang/protobuf/protoc-gen-go/generator"
)

type Generator struct {
	Request  *plugin.CodeGeneratorRequest  // The input.
	Response *plugin.CodeGeneratorResponse // The output.
}

func New() *Generator {
	ret := &Generator{
		Request:  &plugin.CodeGeneratorRequest{},
		Response: &plugin.CodeGeneratorResponse{},
	}
	return ret
}

// Error reports a problem, including an error, and exits the program.
func (g *Generator) Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-bq-schema: error:", s)
	os.Exit(1)
}

// Fail reports a problem and exits the program.
func (g *Generator) Fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("protoc-gen-bq-schema: error:", s)
	os.Exit(1)
}

func (g *Generator) WrapTypes() {
	for _, f := range g.Request.ProtoFile {
		for _, mt := range f.MessageType { // *descriptor.DescriptorProto
			if mt.Options == nil {
				continue
			}
			log.Printf("checking MessageType: %v", *mt.Name)
			var datasetID, tableID, tableName, tableDescription *string
			var ok bool

			val, err := proto.GetExtension(mt.Options, tabledef.E_DatasetId)
			if err == nil {
				datasetID, ok = val.(*string)
				if !ok {
					g.Fail("Couldn't cast dataset_id to string.")
				}
			}

			val, err = proto.GetExtension(mt.Options, tabledef.E_TableId)
			if err == nil {
				tableID, ok = val.(*string)
				if !ok {
					g.Fail("Couldn't cast table_id to string.")
				}
			}

			val, err = proto.GetExtension(mt.Options, tabledef.E_TableName)
			if err == nil {
				tableName, ok = val.(*string)
				if !ok {
					g.Fail("Couldn't cast table_name to string.")
				}

			}

			val, err = proto.GetExtension(mt.Options, tabledef.E_TableDescription)
			if err == nil {
				tableDescription, ok = val.(*string)
				if !ok {
					g.Fail("Couldn't cast table_description to string.")
				}
			}

			if datasetID == nil && tableID != nil || datasetID != nil && tableID == nil {
				g.Fail("Must declare both dataset_id and table_id if either is declared.")
			}

			if datasetID != nil { // we have everything we need.
				log.Printf("Should create table: %q, %q, %q, %q", *datasetID, *tableID, *tableName, *tableDescription)
			} else {
				log.Printf("%q has options but not for bq schema.")
			}
		}
	}
}

func main() {
	g := New()

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

	// Do stuff with g...
	g.WrapTypes()

	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}
