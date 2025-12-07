//go:build e2e

package web

import (
	"html/template"
	"log"
	"mime/multipart"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploader(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)
	engine := &GoTemplateEngine{
		T: tpl,
	}
	server := NewHTTPServer(ServerWithTemplateEngine(engine))
	server.Get("/upload", func(ctx *Context) {
		err := ctx.Render("upload.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	fu := &FileUploader{
		//<input type="file" name="myfile" />
		FileField: "myfile",
		DstPathFunc: func(header *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", header.Filename)
		},
	}
	server.Post("/upload", fu.Handle())
	server.Start(":8081")
}

func TestDownloader(t *testing.T) {
	server := NewHTTPServer()
	fu := FileDownloader{
		Dir: filepath.Join("testdata", "download"),
	}
	server.Get("/download", fu.Handle())
	server.Start(":8081")
}

func TestStaticResourceHandler_Handle(t *testing.T) {
	server := NewHTTPServer()
	fu, err := NewStaticResourceHandler(filepath.Join("testdata", "static"))
	require.NoError(t, err)
	// localhost:8081/static/xxx.jpg
	server.Get("/static/:file", fu.Handle)
	server.Start(":8081")
}
