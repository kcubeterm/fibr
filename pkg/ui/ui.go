package ui

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/fibr/pkg/provider"
	"github.com/ViBiOh/fibr/pkg/utils"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/templates"
	"github.com/ViBiOh/httputils/pkg/tools"
)

// App for rendering UI
type App struct {
	rootDirectory string
	debug         bool
	config        *provider.Config
	tpl           *template.Template
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`publicURL`: flag.String(tools.ToCamel(fmt.Sprintf(`%sPublicURL`, prefix)), `https://fibr.vibioh.fr`, `[fibr] Public URL`),
		`version`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sVersion`, prefix)), ``, `[fibr] Version (used mainly as a cache-buster)`),
	}
}

// NewApp create ui from given config
func NewApp(config map[string]*string, rootDirectory string) *App {
	tpl := template.New(`fibr`)

	tpl.Funcs(template.FuncMap{
		`filename`: func(file os.FileInfo) string {
			return file.Name()
		},
		`urlescape`: func(url string) string {
			return strings.Replace(url, ` `, `%20`, -1)
		},
		`sha1`: func(file os.FileInfo) string {
			return tools.Sha1(file.Name())
		},
		`asyncImage`: func(file os.FileInfo, version string) map[string]interface{} {
			return map[string]interface{}{
				`File`:        file,
				`Fingerprint`: template.JS(tools.Sha1(file.Name())),
				`Version`:     version,
			}
		},
		`rebuildPaths`: func(parts []string, index int) string {
			return path.Join(parts[:index+1]...)
		},
		`iconFromExtension`: func(file os.FileInfo) string {
			extension := strings.ToLower(path.Ext(file.Name()))

			switch {
			case provider.ArchiveExtensions[extension]:
				return `file-archive`
			case provider.AudioExtensions[extension]:
				return `file-audio`
			case provider.CodeExtensions[extension]:
				return `file-code`
			case provider.ExcelExtensions[extension]:
				return `file-excel`
			case provider.ImageExtensions[extension]:
				return `file-image`
			case provider.PdfExtensions[extension]:
				return `file-pdf`
			case provider.VideoExtensions[extension]:
				return `file-video`
			case provider.WordExtensions[extension]:
				return `file-word`
			default:
				return `file`
			}
		},
		`isImage`: func(file os.FileInfo) bool {
			return provider.ImageExtensions[path.Ext(file.Name())]
		},
		`hasThumbnail`: func(request *provider.Request, file os.FileInfo) bool {
			_, info := provider.GetFileinfoFromRoot(path.Join(rootDirectory, provider.MetadataDirectoryName), request, []byte(file.Name()))
			return info != nil
		},
	})

	fibrTemplates, err := utils.ListFilesByExt(`./templates/`, `.gohtml`)
	if err != nil {
		log.Fatalf(`Error while getting templates: %v`, err)
	}

	return &App{
		rootDirectory: rootDirectory,
		debug:         os.Getenv(`DEBUG`) == `true`,
		tpl:           template.Must(tpl.ParseFiles(fibrTemplates...)),
		config: &provider.Config{
			RootName:  path.Base(rootDirectory),
			PublicURL: *config[`publicURL`],
			Version:   *config[`version`],
			Seo: &provider.Seo{
				Title:       `fibr`,
				Description: fmt.Sprintf(`FIle BRowser`),
				Img:         fmt.Sprintf(`/favicon/android-chrome-512x512.png?v=%s`, *config[`version`]),
				ImgHeight:   512,
				ImgWidth:    512,
			},
		},
	}
}

// Error render error page with given status
func (a *App) Error(w http.ResponseWriter, status int, err error) {
	page := &provider.Page{
		Config: a.config,
		Error: &provider.Error{
			Status: status,
		},
	}

	if err := templates.WriteHTMLTemplate(a.tpl.Lookup(`error`), w, page, status); err != nil {
		httperror.InternalServerError(w, err)
	}

	log.Printf(`[error] %v`, err)
}

// Sitemap render sitemap.xml
func (a *App) Sitemap(w http.ResponseWriter) {
	if err := templates.WriteXMLTemplate(a.tpl.Lookup(`sitemap`), w, provider.Page{Config: a.config}, http.StatusOK); err != nil {
		httperror.InternalServerError(w, err)
	}
}

// Directory render directory listing
func (a *App) Directory(w http.ResponseWriter, request *provider.Request, content map[string]interface{}, layout string, message *provider.Message) {
	page := &provider.Page{
		Config:  a.config,
		Request: request,
		Message: message,
		Layout:  layout,
		Content: content,
	}

	if request.IsDebug && a.debug {
		if err := httpjson.ResponseJSON(w, http.StatusOK, page, true); err != nil {
			a.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.Header().Set(`content-language`, `fr`)
	if err := templates.WriteHTMLTemplate(a.tpl.Lookup(`files`), w, page, http.StatusOK); err != nil {
		a.Error(w, http.StatusInternalServerError, err)
	}
}
