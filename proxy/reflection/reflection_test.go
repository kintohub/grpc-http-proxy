package reflection

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
	_ "google.golang.org/grpc/test/grpc_testing"

	perrors "github.com/mercari/grpc-http-proxy/errors"
)

type mockGrpcreflectClient struct {
}

func (m *mockGrpcreflectClient) ResolveService(serviceName string) (*desc.ServiceDescriptor, error) {
	if serviceName == "not.found.NoService" {
		return nil, errors.Errorf("service not found")
	}
	return &desc.ServiceDescriptor{}, nil

}

func TestReflectionClient_ResolveService(t *testing.T) {
	cases := []struct {
		name        string
		serviceName string
		descIsNil   bool
		error       *perrors.Error
	}{
		{
			name:        "found",
			serviceName: "grpc.testing.TestService",
			descIsNil:   false,
			error:       nil,
		},
		{
			name:        "not found",
			serviceName: "not.found.NoService",
			descIsNil:   true,
			error: &perrors.Error{
				Code:    perrors.ServiceNotFound,
				Message: fmt.Sprintf("service %s was not found upstream", "not.found.NoService"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			c := NewReflectionClient(&mockGrpcreflectClient{})
			serviceDesc, err := c.ResolveService(ctx, tc.serviceName)
			if got, want := serviceDesc == nil, tc.descIsNil; got != want {
				t.Fatalf("got %t, want %t", got, want)
			}
			{
				err, ok := err.(*perrors.Error)
				if !ok {
					err = nil
				}
				if got, want := err, tc.error; !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestServiceDescriptor_FindMethodByName(t *testing.T) {
	const serviceName = "grpc.testing.TestService"
	const file = "grpc_testing/test.proto"
	cases := []struct {
		name        string
		serviceName string
		methodName  string
		descIsNil   bool
		error       *perrors.Error
	}{
		{
			name:       "method found",
			methodName: "EmptyCall",
			descIsNil:  false,
			error:      nil,
		},
		{
			name:       "method not found",
			methodName: "ThisMethodDoesNotExist",
			descIsNil:  true,
			error: &perrors.Error{
				Code:    perrors.MethodNotFound,
				Message: fmt.Sprintf("the method %s was not found", "ThisMethodDoesNotExist"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fileDesc := newFileDescriptor(t, file)
			serviceDesc := ServiceDescriptorFromFileDescriptor(fileDesc, serviceName)
			if serviceDesc == nil {
				t.Fatalf("service descriptor is nil")
			}
			methodDesc, err := serviceDesc.FindMethodByName(tc.methodName)
			if got, want := methodDesc == nil, tc.descIsNil; got != want {
				t.Fatalf("got %t, want %t", got, want)
			}
			{
				err, ok := err.(*perrors.Error)
				if !ok {
					err = nil
				}
				if got, want := err, tc.error; !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestServiceDescriptor_GetInputType(t *testing.T) {
	const serviceName = "grpc.testing.TestService"
	const file = "grpc_testing/test.proto"
	cases := []struct {
		name        string
		serviceName string
		methodName  string
		descIsNil   bool
	}{
		{
			name:        "input type found",
			serviceName: "TestService",
			methodName:  "EmptyCall",
			descIsNil:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fileDesc := newFileDescriptor(t, file)
			serviceDesc := ServiceDescriptorFromFileDescriptor(fileDesc, serviceName)
			methodDesc, err := serviceDesc.FindMethodByName(tc.methodName)
			if err != nil {
				t.Fatalf(err.Error())
			}
			inputMsgDesc := methodDesc.GetInputType()
			if got, want := inputMsgDesc == nil, tc.descIsNil; got != want {
				t.Fatalf("got %t, want %t", got, want)
			}
		})
	}
}

func TestServiceDescriptor_GetOutputType(t *testing.T) {
	const serviceName = "grpc.testing.TestService"
	const file = "grpc_testing/test.proto"
	cases := []struct {
		name        string
		serviceName string
		methodName  string
		descIsNil   bool
	}{
		{
			name:        "output type found",
			serviceName: "TestService",
			methodName:  "EmptyCall",
			descIsNil:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fileDesc := newFileDescriptor(t, file)
			serviceDesc := ServiceDescriptorFromFileDescriptor(fileDesc, serviceName)
			methodDesc, err := serviceDesc.FindMethodByName(tc.methodName)
			if err != nil {
				t.Fatalf(err.Error())
			}
			inputMsgDesc := methodDesc.GetOutputType()
			if got, want := inputMsgDesc == nil, tc.descIsNil; got != want {
				t.Fatalf("got %t, want %t", got, want)
			}
		})
	}
}

func TestMessageDescriptor_NewMessage(t *testing.T) {
	const serviceName = "grpc.testing.TestService"
	const methodName = "EmptyCall"
	const file = "grpc_testing/test.proto"
	fileDesc := newFileDescriptor(t, file)
	serviceDesc := ServiceDescriptorFromFileDescriptor(fileDesc, serviceName)
	if serviceDesc == nil {
		t.Fatal("service descriptor is nil")
	}
	methodDesc, err := serviceDesc.FindMethodByName(methodName)
	if err != nil {
		t.Fatalf(err.Error())
	}
	inputMsgDesc := methodDesc.GetInputType()
	inputMsg := inputMsgDesc.NewMessage()
	if got, want := inputMsg == nil, false; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
}

func TestMessage_MarshalJSON(t *testing.T) {
	const serviceName = "grpc.testing.TestService"
	const methodName = "EmptyCall"
	const file = "grpc_testing/test.proto"
	const messageName = "grpc.testing.Payload"
	fileDesc := newFileDescriptor(t, file)
	cases := []struct {
		name string
		json []byte
		error
	}{
		{
			name:  "success",
			json:  []byte("{\"body\":\"aGVsbG8=\"}"),
			error: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageDesc := fileDesc.FindMessage(messageName)
			if messageDesc == nil {
				t.Fatal("messageImpl descriptor is nil")
			}
			message := messageImpl{
				message: dynamic.NewMessage(messageDesc),
			}
			message.message.SetField(message.message.FindFieldDescriptorByName("body"), []byte("hello"))
			j, err := message.MarshalJSON()
			if got, want := j, tc.json; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
			if got, want := err, tc.error; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
	}
}

func TestMessage_UnmarshalJSON(t *testing.T) {
	const serviceName = "grpc.testing.TestService"
	const methodName = "EmptyCall"
	const file = "grpc_testing/test.proto"
	const messageName = "grpc.testing.Payload"
	fileDesc := newFileDescriptor(t, file)
	cases := []struct {
		name string
		json []byte
		error
	}{
		{
			name:  "success",
			json:  []byte("{\"body\":\"aGVsbG8=\"}"),
			error: nil,
		},
		{
			name: "type mismatch",
			json: []byte("{\"body\":\"hello!\""),
			error: &perrors.Error{
				Code:    perrors.MessageTypeMismatch,
				Message: "input JSON does not match messageImpl type",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			messageDesc := fileDesc.FindMessage(messageName)
			if messageDesc == nil {
				t.Fatal("messageImpl descriptor is nil")
			}
			message := messageImpl{
				message: dynamic.NewMessage(messageDesc),
			}
			err := message.UnmarshalJSON(tc.json)

			expectedMessage := dynamic.NewMessage(messageDesc)
			expectedMessage.SetField(expectedMessage.FindFieldDescriptorByName("body"), []byte("hello!"))

			if got, want := err, tc.error; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
	}
}