package httpreq

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
)

const (
	HeaderContentType         = "Content-Type"
	ContentTypeText           = "text/plain; charset=utf-8"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeJSON           = "application/json"
	ContentTypeXML            = "application/xml"
)

type Marshaller interface {
	Marshal(value interface{}) ([]byte, error)
}

var marshallers = map[string]Marshaller{
	ContentTypeText:           textPlainMarshaller{},
	ContentTypeFormURLEncoded: formURLEncodedMarshaller{},
	ContentTypeJSON:           jsonMarshaller{},
	ContentTypeXML:            xmlMarshaller{},
}

func RegisterMarshaller(contentType string, marshaller Marshaller) {
	marshallers[contentType] = marshaller
}

// Build-in Marshaller implementation
type textPlainMarshaller struct {
}

func (m textPlainMarshaller) Marshal(value interface{}) (bytes []byte, err error) {
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		err = fmt.Errorf("text/plain: unsupported type: %t", value)
	}
	return
}

type formURLEncodedMarshaller struct {
}

func (m formURLEncodedMarshaller) Marshal(value interface{}) (bytes []byte, err error) {
	switch v := value.(type) {
	case url.Values:
		bytes = []byte(v.Encode())
	case map[string][]string:
		bytes = []byte((url.Values)(v).Encode())
	default:
		err = fmt.Errorf("x-www-form-urlencoded: unsupported type: %t", value)
	}
	return
}

type jsonMarshaller struct {
}

func (m jsonMarshaller) Marshal(value interface{}) (bytes []byte, err error) {
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		bytes, err = json.Marshal(value)
	}
	return
}

type xmlMarshaller struct {
}

func (m xmlMarshaller) Marshal(value interface{}) (bytes []byte, err error) {
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		bytes, err = xml.Marshal(value)
	}
	return
}
