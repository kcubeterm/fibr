package crud

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/fibr/pkg/provider"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func createOrOpenFile(filename string, info os.FileInfo) (io.WriteCloser, error) {
	if info == nil {
		return os.Create(filename)
	}
	return os.Open(filename)
}

// CreateDir creates given path directory to filesystem
func (a *App) CreateDir(w http.ResponseWriter, r *http.Request, request *provider.Request) {
	if !request.CanEdit {
		a.renderer.Error(w, http.StatusForbidden, ErrNotAuthorized)
		return
	}

	var filename string

	formName := r.FormValue(`name`)
	if formName != `` {
		filename, _ = a.getFileinfo(request, []byte(formName))
	}

	if filename == `` {
		if !strings.HasSuffix(request.Path, `/`) {
			a.renderer.Error(w, http.StatusForbidden, ErrNotAuthorized)
			return
		}

		filename, _ = a.getFileinfo(request, nil)
	}

	if strings.Contains(filename, `..`) {
		a.renderer.Error(w, http.StatusForbidden, ErrNotAuthorized)
		return
	}

	if err := os.MkdirAll(filename, 0700); err != nil {
		a.renderer.Error(w, http.StatusInternalServerError, fmt.Errorf(`Error while creating directory: %v`, err))
		return
	}

	a.GetDir(w, request, path.Dir(filename), r.URL.Query().Get(`d`), &provider.Message{Level: `success`, Content: fmt.Sprintf(`Directory %s successfully created`, path.Base(filename))})
}

func (a *App) saveUploadedFile(request *provider.Request, uploadedFile io.ReadCloser, uploadedFileHeader *multipart.FileHeader) (string, error) {
	filename, info := a.getFileinfo(request, []byte(uploadedFileHeader.Filename))
	hostFile, err := createOrOpenFile(filename, info)
	if hostFile != nil {
		defer func() {
			if err := hostFile.Close(); err != nil {
				log.Printf(`Error while closing writted file: %v`, err)
			}
		}()
	}

	if err != nil {
		return ``, fmt.Errorf(`Error while creating or opening file: %v`, err)
	} else if _, err = io.Copy(hostFile, uploadedFile); err != nil {
		return ``, fmt.Errorf(`Error while writing file: %v`, err)
	} else if provider.ImageExtensions[path.Ext(uploadedFileHeader.Filename)] {
		go a.generateImageThumbnail(strings.TrimPrefix(filename, a.rootDirectory))
	}

	return filename, nil
}

// SaveFiles saves form files to filesystem
func (a *App) SaveFiles(w http.ResponseWriter, r *http.Request, request *provider.Request) {
	if !request.CanEdit {
		a.renderer.Error(w, http.StatusForbidden, ErrNotAuthorized)
	}

	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		a.renderer.Error(w, http.StatusBadRequest, fmt.Errorf(`Error while parsing form: %v`, err))
		return
	}

	if r.MultipartForm.File == nil || len(r.MultipartForm.File[`files[]`]) == 0 {
		a.renderer.Error(w, http.StatusBadRequest, errors.New(`No file provided for save`))
		return
	}

	var outputDir string
	filenames := make([]string, len(r.MultipartForm.File[`files[]`]))

	for index, file := range r.MultipartForm.File[`files[]`] {
		uploadedFile, err := file.Open()
		if uploadedFile != nil {
			defer func() {
				if err := uploadedFile.Close(); err != nil {
					log.Printf(`Error while closing uploaded file: %v`, err)
				}
			}()
		}
		if err != nil {
			a.renderer.Error(w, http.StatusBadRequest, fmt.Errorf(`Error while getting file from form: %v`, err))
			return
		}

		filename, err := a.saveUploadedFile(request, uploadedFile, file)
		if err != nil {
			a.renderer.Error(w, http.StatusInternalServerError, err)
			return
		}

		if outputDir == `` {
			outputDir = path.Dir(filename)
		}

		filenames[index] = file.Filename
	}

	message := fmt.Sprintf(`File %s successfully uploaded`, filenames[0])
	if len(filenames) > 1 {
		message = fmt.Sprintf(`Files %s successfully uploaded`, strings.Join(filenames, `, `))
	}

	a.GetDir(w, request, outputDir, r.URL.Query().Get(`d`), &provider.Message{Level: `success`, Content: message})
}
