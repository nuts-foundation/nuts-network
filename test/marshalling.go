package test

import (
	"bytes"
	"encoding/json"
	"io"
)

type JSONReader struct {
	Input interface{}
	reader io.Reader
}

func (j *JSONReader) Read(p []byte) (n int, err error) {
	if j.reader == nil {
		data, err := json.Marshal(j.Input)
		if err != nil {
			return 0, err
		}
		j.reader = bytes.NewReader(data)
	}
	return j.reader.Read(p)
}

