package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"github.com/rprtr258/fun/exp/zun"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/impulse-http/local-backend/internal/database"
)

func connect(
	ctx context.Context,
	target string,
) (grpcurl.DescriptorSource, *grpc.ClientConn, error) {
	cc, err := grpcurl.BlockingDial(ctx, "tcp", target, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "dial")
	}

	refClient := grpcreflect.NewClientAuto(ctx, cc)
	refClient.AllowMissingFileDescriptors()
	return grpcurl.DescriptorSourceFromServer(ctx, refClient), cc, nil
}

func splitService(serviceName string) (pkg, short string) {
	dotI := strings.LastIndexByte(serviceName, '.')
	return serviceName[:dotI], serviceName[dotI+1:]
}

// JSONSchema represents the structure of a JSON schema
type JSONSchema struct {
	Type string `json:"type"`
	// type == "object"
	Properties map[string]JSONSchema `json:"properties,omitempty"`
	// type == "object" && "oneOf" is present
	OneOf []JSONSchema `json:"oneOf,omitempty"`
	// type == "array"
	Items *JSONSchema `json:"items,omitempty"`
}

func convertObjectToJSONSchema(msg protoreflect.MessageDescriptor) (JSONSchema, error) {
	fields := msg.Fields()
	properties := make(map[string]JSONSchema, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldSchema, err := convertFieldToJSONSchema(field)
		if err != nil {
			return JSONSchema{}, err
		}

		if oneof := field.ContainingOneof(); oneof != nil {
			oneofName := string(oneof.Name())
			if _, ok := properties[oneofName]; !ok {
				properties[oneofName] = JSONSchema{
					Type:  "object",
					OneOf: []JSONSchema{},
				}
			}
			m := properties[oneofName]
			m.OneOf = append(m.OneOf, JSONSchema{
				Type:       "object",
				Properties: map[string]JSONSchema{string(field.Name()): fieldSchema},
			})
			properties[oneofName] = m
		} else {
			properties[string(field.Name())] = fieldSchema
		}
	}

	return JSONSchema{
		Type:       "object",
		Properties: properties,
		Items:      nil,
	}, nil
}

func convertFieldToJSONSchema(field protoreflect.FieldDescriptor) (JSONSchema, error) {
	var fieldSchema JSONSchema
	switch field.Kind() {
	case protoreflect.BoolKind:
		fieldSchema.Type = "boolean"
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		fieldSchema.Type = "integer"
	case protoreflect.FloatKind,
		protoreflect.DoubleKind:
		fieldSchema.Type = "number"
	case protoreflect.StringKind:
		fieldSchema.Type = "string"
	case protoreflect.BytesKind:
		fieldSchema.Type = "string"
		// TODO: support bytes
	case protoreflect.MessageKind:
		var err error
		fieldSchema, err = convertObjectToJSONSchema(field.Message())
		if err != nil {
			return JSONSchema{}, err
		}
	case protoreflect.EnumKind:
		fieldSchema.Type = "string"
		vals := field.Enum().Values()
		for i := 0; i < vals.Len(); i++ {
			val := vals.Get(i)
			fmt.Println("ENUM", i, val.Name())
		}
	default:
		return JSONSchema{}, fmt.Errorf("unsupported field kind: %v", field.Kind())
	}

	if field.IsList() {
		return JSONSchema{
			Type:  "array",
			Items: &fieldSchema,
		}, nil
	}

	return fieldSchema, nil
}

func ConvertMessageToJSONSchema(msg protoreflect.Message) (JSONSchema, error) {
	return convertObjectToJSONSchema(msg.Descriptor())
}

func newFake(js JSONSchema) any {
	switch js.Type {
	case "object":
		m := make(map[string]any, len(js.Properties))
		for k, v := range js.Properties {
			if v.Type == "object" && v.OneOf != nil {
				idx := rand.Intn(len(v.OneOf))

				// TODO: (((GO)))VNO
				var kkk string
				var vvv JSONSchema
				for k, v := range v.OneOf[idx].Properties {
					kkk = k
					vvv = v
				}

				m[kkk] = newFake(vvv)
			} else {
				m[k] = newFake(v)
			}
		}
		return m
	case "array":
		a := []any{}
		for _, v := range js.Items.Properties {
			a = append(a, newFake(v))
		}
		return a
	case "number":
		return float64(5.2)
	case "integer":
		return int64(52)
	case "string":
		return "ABOBA"
	default:
		return nil
	}
}

type grpcServiceMethods struct {
	Service string   `json:"service"`
	Methods []string `json:"methods"`
}

func (s *Service) HandlerGRPCMethods(ctx context.Context, req struct {
	Target string `json:"target"`
}) ([]grpcServiceMethods, error) {
	reflSource, cc, err := connect(ctx, req.Target)
	if err != nil {
		return nil, errors.Wrap(err, "connect")
	}
	defer cc.Close()

	services, err := grpcurl.ListServices(reflSource)
	if err != nil {
		return nil, errors.Wrap(err, "list services")
	}

	res := make([]grpcServiceMethods, 0, len(services))
	for _, service := range services {
		pkg, serviceName := splitService(service)
		methods, err := grpcurl.ListMethods(reflSource, service)
		if err != nil {
			return nil, errors.Wrapf(err, "list service %s methods", service)
		}

		prefix := pkg + "." + serviceName + "."
		zun.Map(methods, func(method string) string {
			return strings.TrimPrefix(method, prefix)
		}, methods...)
		res = append(res, grpcServiceMethods{
			service,
			methods,
		})
	}
	return res, nil
}

func (s *Service) HandlerGRPCQueryFake(ctx context.Context, req struct {
	Target string `json:"target"`
	Method string `json:"method"` // NOTE: fully qualified
}) (string, error) {
	reflSource, cc, err := connect(ctx, req.Target)
	if err != nil {
		return "", errors.Wrap(err, "connect")
	}
	defer cc.Close()

	dsc, err := reflSource.FindSymbol(req.Method)
	if err != nil {
		return "", errors.Wrap(err, "find method")
	}

	methodDesc := dsc.(*desc.MethodDescriptor)
	// fmt.Println("    IN", strings.TrimPrefix(inputType.GetFullyQualifiedName(), pkg+"."))
	// fmt.Println("    OUT", strings.TrimPrefix(methodDesc.GetOutputType().GetFullyQualifiedName(), pkg+"."))
	m := dynamicpb.NewMessage(methodDesc.GetInputType().UnwrapMessage())

	schema, err := ConvertMessageToJSONSchema(m)
	// schema["$schema"]= "http://json-schema.org/schema#"
	if err != nil {
		return "", errors.Wrap(err, "convert message to json schema")
	}

	// jsonSchema, err := json.MarshalIndent(schema, "", "  ")
	// check(err)

	b, err := json.MarshalIndent(newFake(schema), "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "marshal fake")
	}
	return string(b), nil
}

func (s *Service) sendGRPC(ctx context.Context, req database.GRPCRequest) (database.GRPCResponse, error) {
	reflSource, cc, err := connect(ctx, req.Target)
	if err != nil {
		return database.GRPCResponse{}, errors.Wrap(err, "connect")
	}
	defer cc.Close()

	// bb := invoke(ctx, reflSource, cc, "helloworld.Greeter.SayHello", string(b))
	formatter := grpcurl.NewJSONFormatter(true, nil)
	var bb bytes.Buffer
	r := bytes.NewReader([]byte(req.Payload))
	if err := grpcurl.InvokeRPC(
		ctx, reflSource, cc, req.Method, nil,
		&grpcurl.DefaultEventHandler{
			Out:            &bb,
			Formatter:      formatter,
			VerbosityLevel: 0,
		},
		grpcurl.NewJSONRequestParserWithUnmarshaler(r, jsonpb.Unmarshaler{}).Next,
	); err != nil {
		return database.GRPCResponse{}, errors.Wrap(err, "invoke rpc")
	}

	return database.GRPCResponse{
		bb.String(),
		0, // TODO: pass response code
	}, nil
}

func (s *Service) HandlerGRPCQueryValidate(ctx context.Context, req struct {
	Target  string `json:"target"`
	Method  string `json:"method"` // NOTE: fully qualified
	Payload string `json:"payload"`
}) (struct{}, error) {
	return struct{}{}, errors.New("not implemented")
}
