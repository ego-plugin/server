package erestful

import (
	"github.com/ego-plugin/server/erestful/staticswagger"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"net/http"
	"net/url"
	"strings"
)

// NewSwaggerService API文档服务
func NewSwaggerService(container *restful.Container) *restful.WebService {
	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		APIPath:                       "/swagger.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	container.ServeMux.Handle("/swagger/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, "/swagger"); len(p) < len(r.URL.Path) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = "swagger" + p
			http.FileServer(staticswagger.FS(false)).ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	}))
	return restfulspec.NewOpenAPIService(config)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.BasePath = ""
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "API",
			Description: "API Development Document",
			Contact: &spec.ContactInfo{
				Name:  "ego",
				Email: "system18188@gmail.com",
				URL:   "https://swagger.io/",
			},
			License: &spec.License{
				Name: "MIT",
				URL:  "http://mit.org",
			},
			Version: "2.0.0",
		},
	}
	// swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
	//			"jwt": spec.APIKeyAuth("Authorization", "header"),
	//		}
}

