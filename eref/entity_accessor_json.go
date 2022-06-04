package eref

import (
	"github.com/ego-plugin/binding"
	"github.com/emicklei/go-restful/v3"
)

// NewEntityAccessorJson 自己动解析自定义类型
func NewEntityAccessorJson() restful.EntityReaderWriter {
	return entityAccessorJson{}
}

// entityOctetAccess is a go-restful EntityReaderWriter for Octet encoding
type entityAccessorJson struct{}

// Read unmarshalls the value from byte slice and using 自动 to unmarshal
func (e entityAccessorJson) Read(req *restful.Request, v interface{}) error {
	valid := binding.Default(req.Request.Method, req.HeaderParameter(restful.HEADER_ContentType))
	return valid.Bind(req.Request, v, binding.LANG_EN)
}

// Write marshals the value to byte slice and set the Content-Type Header.
func (e entityAccessorJson) Write(resp *restful.Response, status int, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		// do not write a nil representation
		return nil
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteHeader(status)
	return restful.NewEncoder(resp).Encode(v)

}
