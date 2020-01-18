// Code generated by protoc-gen-go-json. DO NOT EDIT.
// source: entity.proto

package data

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
)

// MarshalJSON implements json.Marshaler
func (msg *EntityRealtimeData) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := (&jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		OrigName:     false,
	}).Marshal(&buf, msg)
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (msg *EntityRealtimeData) UnmarshalJSON(b []byte) error {
	return jsonpb.Unmarshal(bytes.NewReader(b), msg)
}

// MarshalJSON implements json.Marshaler
func (msg *EntityData) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := (&jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		OrigName:     false,
	}).Marshal(&buf, msg)
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (msg *EntityData) UnmarshalJSON(b []byte) error {
	return jsonpb.Unmarshal(bytes.NewReader(b), msg)
}