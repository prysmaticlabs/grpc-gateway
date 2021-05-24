package runtime

import (
	"io"

	"google.golang.org/protobuf/encoding/protojson"
)

// EventSourceJSONPb is a Marshaler which marshals/unmarshals into/from JSON
// with the "google.golang.org/protobuf/encoding/protojson" marshaler for text/event-stream.
type EventSourceJSONPb struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
	JSONPb
}

// ContentType for text/event-stream.
func (*EventSourceJSONPb) ContentType(_ interface{}) string {
	return "text/event-stream"
}

// Marshal marshals "v" into JSON.
func (j *EventSourceJSONPb) Marshal(v interface{}) ([]byte, error) {
	return j.JSONPb.Marshal(v)
}

// Unmarshal unmarshals JSON "data" into "v"
func (j *EventSourceJSONPb) Unmarshal(data []byte, v interface{}) error {
	return j.JSONPb.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (j *EventSourceJSONPb) NewDecoder(r io.Reader) Decoder {
	return j.JSONPb.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *EventSourceJSONPb) NewEncoder(w io.Writer) Encoder {
	return j.JSONPb.NewEncoder(w)
}

// Delimiter for text/event-stream.
func (j *EventSourceJSONPb) Delimiter() []byte {
	return []byte("\n")
}
