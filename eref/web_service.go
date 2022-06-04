package eref

import (
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

// WebService holds a collection of Route values that bind a Http Method + URL Path to a function.
type WebService struct {
	v []*restful.WebService
}

func NewWebService() *WebService {
	return &WebService{
		v: make([]*restful.WebService, 0),
	}
}

func (w *WebService) Append(route ...*restful.WebService) {
	w.v = append(w.v, route...)
}

func (w *WebService) AppendWebService(ws *WebService) {
	w.v = append(w.v, ws.Value()...)
}

func (w *WebService) Value() []*restful.WebService {
	return w.v
}

func (w *WebService) Build() {
	for _, v := range w.v {
		restful.DefaultContainer.Add(v)
	}
}

func WebHandle(pattern string, handler http.HandlerFunc) {
	restful.DefaultContainer.ServeMux.Handle(pattern, handler)
}
