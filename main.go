package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

func main() {

	// testing bytes of json string since it's the return value of https://github.com/zendesk/golang-locust/blob/master/pkg/shared/step/kafka/runner.go#L204-L217
	// booksBytes := []byte(`[{"author":"Brandon Sanderson","rating":4.98,"title":"Words of Radiance"},{"author":"Maria Ressa","title":"How to Standup to a Dictator"}]`)
	booksBytes := []byte(`[{"title":"Words of Radiance","author":"Brandon Sanderson","rating":4.98},{"title":"How to Standup to a Dictator","author":"Maria Ressa"}]`)

	readDynamically(booksBytes)
	ConvertInterfaceToAny(booksBytes)
}

// testing https://stackoverflow.com/questions/65242456/convert-protobuf-serialized-messages-to-json-without-precompiling-go-code
func readDynamically(in []byte) {
	registry, err := createProtoRegistry(".", "books.proto")
	if err != nil {
		panic(err)
	}

	desc, err := registry.FindFileByPath("books.proto")
	if err != nil {
		panic(err)
	}
	fd := desc.Messages()
	addressBook := fd.ByName("Books")

	msg := dynamicpb.NewMessage(addressBook)
	err = protojson.Unmarshal(in, msg)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := protojson.Marshal(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))

}

func createProtoRegistry(srcDir string, filename string) (*protoregistry.Files, error) {
	// Create descriptors using the protoc binary.
	// Imported dependencies are included so that the descriptors are self-contained.
	tmpFile := filename + "-tmp.pb"
	cmd := exec.Command("protoc",
		"--include_imports",
		"--descriptor_set_out="+tmpFile,
		"-I"+srcDir,
		path.Join(srcDir, filename))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	// defer os.Remove(tmpFile)

	marshalledDescriptorSet, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return nil, err
	}
	descriptorSet := descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(marshalledDescriptorSet, &descriptorSet)
	if err != nil {
		return nil, err
	}

	files, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// attempt 2: https://ravina01997.medium.com/converting-interface-to-any-proto-and-vice-versa-in-golang-27badc3e23f1
func ConvertInterfaceToAny(v interface{}) (*any.Any, error) {
	anyValue := &any.Any{}
	bytes, _ := json.Marshal(v)
	bytesValue := &wrappers.BytesValue{
		Value: bytes,
	}
	err := anypb.MarshalFrom(anyValue, bytesValue, proto.MarshalOptions{})
	fmt.Println(anyValue.String())
	return anyValue, err
}
