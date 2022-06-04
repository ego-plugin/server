package eref

import (
	"github.com/ego-plugin/binding"
	"github.com/emicklei/go-restful/v3"
)

// NewEntityJsonAccess 自己动解析自定义类型
func NewEntityJsonAccess() restful.EntityReaderWriter {
	return entityJsonAccess{}
}

// entityOctetAccess is a go-restful EntityReaderWriter for Octet encoding
type entityJsonAccess struct{}

// Read unmarshalls the value from byte slice and using 自动 to unmarshal
func (e entityJsonAccess) Read(req *restful.Request, v interface{}) error {
	valid := binding.Default(req.Request.Method, req.HeaderParameter(restful.HEADER_ContentType))
	return valid.Bind(req.Request, v, binding.LANG_EN)
}

// Write marshals the value to byte slice and set the Content-Type Header.
func (e entityJsonAccess) Write(resp *restful.Response, status int, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		// do not write a nil representation
		return nil
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteHeader(status)
	return restful.NewEncoder(resp).Encode(v)

}
