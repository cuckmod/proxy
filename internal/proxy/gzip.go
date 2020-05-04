package proxy

import (
	"bytes"
	"compress/gzip"
	"net/http"
)

func EncodeGzip(data []byte) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	gw := gzip.NewWriter(buffer)

	_, err := gw.Write(data)
	if err != nil {
		return nil, err
	}
	err = gw.Flush()
	if err != nil {
		return nil, err
	}
	err = gw.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecodeGZIP(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	buffer := bytes.Buffer{}
	_, err = buffer.ReadFrom(gr)
	if err != nil {
		return nil, err
	}
	err = gr.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil

}

func isArchived(data []byte) bool {
	return http.DetectContentType(data) == "application/x-gzip"
}
