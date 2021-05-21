package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// EventSourceJSONPb is a Marshaler which marshals/unmarshals into/from JSON
// with the "google.golang.org/protobuf/encoding/protojson" marshaler for text/event-stream.
type EventSourceJSONPb struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

// ContentType always returns "application/json".
func (*EventSourceJSONPb) ContentType(_ interface{}) string {
	return "text/event-stream"
}

// Marshal marshals "v" into JSON.
func (j *EventSourceJSONPb) Marshal(v interface{}) ([]byte, error) {
	if _, ok := v.(proto.Message); !ok {
		return j.marshalNonProtoField(v)
	}

	var buf bytes.Buffer
	if err := j.marshalTo(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (j *EventSourceJSONPb) marshalTo(w io.Writer, v interface{}) error {
	p, ok := v.(proto.Message)
	if !ok {
		buf, err := j.marshalNonProtoField(v)
		if err != nil {
			return err
		}
		_, err = w.Write(buf)
		return err
	}
	b, err := j.MarshalOptions.Marshal(p)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

// marshalNonProto marshals a non-message field of a protobuf message.
// This function does not correctly marshal arbitrary data structures into JSON,
// it is only capable of marshaling non-message field values of protobuf,
// i.e. primitive types, enums; pointers to primitives or enums; maps from
// integer/string types to primitives/enums/pointers to messages.
func (j *EventSourceJSONPb) marshalNonProtoField(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return []byte("null"), nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice {
		if rv.IsNil() {
			if j.EmitUnpopulated {
				return []byte("[]"), nil
			}
			return []byte("null"), nil
		}

		if rv.Type().Elem().Implements(protoMessageType) {
			var buf bytes.Buffer
			err := buf.WriteByte('[')
			if err != nil {
				return nil, err
			}
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					err = buf.WriteByte(',')
					if err != nil {
						return nil, err
					}
				}
				if err = j.marshalTo(&buf, rv.Index(i).Interface().(proto.Message)); err != nil {
					return nil, err
				}
			}
			err = buf.WriteByte(']')
			if err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}

	if rv.Kind() == reflect.Map {
		m := make(map[string]*json.RawMessage)
		for _, k := range rv.MapKeys() {
			buf, err := j.Marshal(rv.MapIndex(k).Interface())
			if err != nil {
				return nil, err
			}
			m[fmt.Sprintf("%v", k.Interface())] = (*json.RawMessage)(&buf)
		}
		if j.Indent != "" {
			return json.MarshalIndent(m, "", j.Indent)
		}
		return json.Marshal(m)
	}
	if enum, ok := rv.Interface().(protoEnum); ok && !j.UseEnumNumbers {
		return json.Marshal(enum.String())
	}
	return json.Marshal(rv.Interface())
}

// Unmarshal unmarshals JSON "data" into "v"
func (j *EventSourceJSONPb) Unmarshal(data []byte, v interface{}) error {
	return unmarshalJSONPb(data, j.UnmarshalOptions, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (j *EventSourceJSONPb) NewDecoder(r io.Reader) Decoder {
	d := json.NewDecoder(r)
	return DecoderWrapper{
		Decoder:          d,
		UnmarshalOptions: j.UnmarshalOptions,
	}
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *EventSourceJSONPb) NewEncoder(w io.Writer) Encoder {
	return EncoderFunc(func(v interface{}) error {
		if err := j.marshalTo(w, v); err != nil {
			return err
		}
		// mimic json.Encoder by adding a newline (makes output
		// easier to read when it contains multiple encoded items)
		_, err := w.Write(j.Delimiter())
		return err
	})
}

// Delimiter for newline encoded JSON streams.
func (j *EventSourceJSONPb) Delimiter() []byte {
	return []byte("\n")
}
