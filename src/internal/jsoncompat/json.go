package jsoncompat

import (
	json "encoding/json"
	jsonv2 "encoding/json/v2"
	"io"
)

func UnmarshalRead(in io.Reader, out any) error {
	return jsonv2.UnmarshalRead(in, out, json.DefaultOptionsV1())
}
