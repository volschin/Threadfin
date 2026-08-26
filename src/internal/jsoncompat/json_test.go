package jsoncompat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

type compatibilityPayload struct {
	Name   string
	Items  []string
	Lookup map[string]int
}

func TestUnmarshalReadMatchesJSONV1(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"duplicate names", []byte(`{"Name":"first","Name":"last"}`)},
		{"invalid UTF-8", []byte("{\"Name\":\"\xff\"}")},
		{"case insensitive", []byte(`{"name":"Threadfin"}`)},
		{"unknown member", []byte(`{"Name":"Threadfin","Other":1}`)},
		{"trailing whitespace", []byte(" {\"Name\":\"Threadfin\"} \n")},
		{"trailing value", []byte(`{"Name":"one"}{"Name":"two"}`)},
		{"trailing garbage", []byte(`{"Name":"one"}x`)},
		{"empty", nil},
		{"whitespace", []byte(" \n\t")},
		{"nil map and slice", []byte(`{"Items":null,"Lookup":null}`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var want compatibilityPayload
			wantErr := json.Unmarshal(tc.data, &want)
			var got compatibilityPayload
			gotErr := UnmarshalRead(bytes.NewReader(tc.data), &got)
			if (gotErr != nil) != (wantErr != nil) {
				t.Fatalf("errors: got=%v v1=%v", gotErr, wantErr)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("values: got=%#v v1=%#v", got, want)
			}
		})
	}
}

var errJSONReader = errors.New("JSON reader failed")

type failingJSONReader struct{}

func (failingJSONReader) Read([]byte) (int, error) { return 0, errJSONReader }

func TestUnmarshalReadReturnsReaderFailure(t *testing.T) {
	var got compatibilityPayload
	reader := io.MultiReader(strings.NewReader(`{"Name":`), failingJSONReader{})
	if err := UnmarshalRead(reader, &got); !errors.Is(err, errJSONReader) {
		t.Fatalf("error = %v", err)
	}
}
