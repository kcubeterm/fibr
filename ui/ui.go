package ui

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils"
)

// Message rendered to user
type Message struct {
	Level   string
	Content string
}

var base map[string]interface{}
var tpl *template.Template

// Init initialize ui
func Init(baseTpl *template.Template, publicURL string, staticURL string, authURL string, version string, root string) {
	tpl = baseTpl

	base = map[string]interface{}{
		`Config`: map[string]interface{}{
			`PublicURL`: publicURL,
			`StaticURL`: staticURL,
			`AuthURL`:   authURL,
			`Version`:   version,
			`Root`:      root,
		},
		`Seo`: map[string]interface{}{
			`Title`:       `fibr`,
			`Description`: fmt.Sprintf(`FIle BRowser on the server`),
			`URL`:         `/`,
			`Img`:         path.Join(staticURL, `/favicon/android-chrome-512x512.png`),
			`ImgHeight`:   512,
			`ImgWidth`:    512,
		},
	}
}

func cloneContent(content map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{})
	for key, value := range content {
		clone[key] = value
	}

	return clone
}

// Login render login page
func Login(w http.ResponseWriter, message *Message) {
	loginContent := cloneContent(base)
	if message != nil {
		loginContent[`Message`] = message
	}

	if err := httputils.WriteHTMLTemplate(tpl.Lookup(`login`), w, loginContent); err != nil {
		httputils.InternalServerError(w, err)
	}
}

// Sitemap render sitemap.xml
func Sitemap(w http.ResponseWriter) {
	if err := httputils.WriteHTMLTemplate(tpl.Lookup(`sitemap`), w, base); err != nil {
		httputils.InternalServerError(w, err)
	}
}

// Directory render Dir content
func Directory(w http.ResponseWriter, path string, info os.FileInfo, files []os.FileInfo, message *Message) {
	pathParts := strings.Split(strings.Trim(path, `/`), `/`)
	if pathParts[0] == `` {
		pathParts = nil
	}

	pageContent := cloneContent(base)
	if message != nil {
		pageContent[`Message`] = message
	}

	seo := base[`Seo`].(map[string]interface{})

	pageContent[`PathParts`] = pathParts
	pageContent[`Current`] = info
	pageContent[`Files`] = files
	pageContent[`Seo`] = map[string]interface{}{
		`Title`:       fmt.Sprintf(`fibr - %s`, path),
		`Description`: fmt.Sprintf(`FIle BRowser of directory %s on the server`, path),
		`URL`:         path,
		`Img`:         seo[`Img`],
		`ImgHeight`:   seo[`ImgHeight`],
		`ImgWidth`:    seo[`ImgWidth`],
	}

	if err := httputils.WriteHTMLTemplate(tpl.Lookup(`files`), w, pageContent); err != nil {
		httputils.InternalServerError(w, err)
	}
}
