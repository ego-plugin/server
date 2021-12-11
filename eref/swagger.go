package eref

import (
	"github.com/ego-plugin/server/eref/staticswagger"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"net/http"
	"net/url"
	"strings"
)

// SwaggerService API文档服务
func SwaggerService() {
	config := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(),
		APIPath:                       "/swagger.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	restful.DefaultContainer.ServeMux.Handle("/swagger/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	restful.Add(restfulspec.NewOpenAPIService(config))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "ApiService",
			Description: "Resource for managing API",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "NETBO",
					Email: "system18188@gmail.com",
					URL:   "http://swagger.org",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "MIT",
					URL:  "http://mit.org",
				},
			},
			Version: "1.0.0",
		},
	}
	//swo.Tags = []spec.Tag{spec.Tag{TagProps: spec.TagProps{
	//	Name:        "API Doc",
	//	Description: "Managing api"}}}
}
