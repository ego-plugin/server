package eref

import (
	"github.com/ego-plugin/binding"
	"github.com/emicklei/go-restful/v3"
	"github.com/vmihailenco/msgpack/v5"
)

const MIME_MSGPACK = "application/x-msgpack"

// NewEntityAccessorMsgPack returns a new EntityReaderWriter for accessing MessagePack content.
// This package is not initialized with such an accessor using the MIME_MSGPACK contentType.
func NewEntityAccessorMsgPack() restful.EntityReaderWriter {
	return entityMsgPackAccess{}
}

// entityOctetAccess is a EntityReaderWriter for Octet encoding
type entityMsgPackAccess struct{}

// Read unmarshalls the value from byte slice and using msgpack to unmarshal
func (e entityMsgPackAccess) Read(req *restful.Request, v interface{}) error {
	// return msgpack.NewDecoder(req.Request.Body).Decode(v)
	valid := binding.Default(req.Request.Method, req.HeaderParameter(restful.HEADER_ContentType))
	return valid.Bind(req.Request, v, binding.LANG_EN)
}

// Write marshals the value to byte slice and set the Content-Type Header.
func (e entityMsgPackAccess) Write(resp *restful.Response, status int, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		// do not write a nil representation
		return nil
	}

	resp.Header().Set(restful.HEADER_ContentType, MIME_MSGPACK)
	resp.WriteHeader(status)
	return msgpack.NewEncoder(resp).Encode(v)
}
