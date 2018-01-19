package crud

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/fibr/provider"
	"github.com/ViBiOh/fibr/utils"
)

func getFileForm(w http.ResponseWriter, r *http.Request) (io.ReadCloser, *multipart.FileHeader, error) {
	uploadedFile, uploadedFileHeader, err := r.FormFile(`file`)
	if err != nil {
		return uploadedFile, uploadedFileHeader, fmt.Errorf(`Error while reading file form: %v`, err)
	}

	return uploadedFile, uploadedFileHeader, nil
}

func createOrOpenFile(filename string, info os.FileInfo) (io.WriteCloser, error) {
	if info == nil {
		return os.Create(filename)
	}
	return os.Open(filename)
}

// CreateDir creates given path directory to filesystem
func (a *App) CreateDir(w http.ResponseWriter, r *http.Request, config *provider.RequestConfig) {
	if !config.CanEdit {
		a.renderer.Error(w, http.StatusForbidden, errors.New(`You're not authorized to do this ⛔`))
	}

	if strings.HasSuffix(config.Path, `/`) {
		filename, _ := utils.GetPathInfo(config.Root, config.Path)

		if err := os.MkdirAll(filename, 0700); err != nil {
			a.renderer.Error(w, http.StatusInternalServerError, fmt.Errorf(`Error while creating directory: %v`, err))
		} else {
			w.WriteHeader(http.StatusCreated)
		}
	} else {
		a.renderer.Error(w, http.StatusForbidden, errors.New(`You're not authorized to do this`))
	}
}

// SaveFile saves form file to filesystem
func (a *App) SaveFile(w http.ResponseWriter, r *http.Request, config *provider.RequestConfig) {
	if !config.CanEdit {
		a.renderer.Error(w, http.StatusForbidden, errors.New(`You're not authorized to do this ⛔`))
	}

	uploadedFile, uploadedFileHeader, err := getFileForm(w, r)
	if uploadedFile != nil {
		defer uploadedFile.Close()
	}
	if err != nil {
		a.renderer.Error(w, http.StatusBadRequest, fmt.Errorf(`Error while getting file from form: %v`, err))
		return
	}

	filename, info := utils.GetPathInfo(config.Root, config.Path, uploadedFileHeader.Filename)
	hostFile, err := createOrOpenFile(filename, info)
	if hostFile != nil {
		defer hostFile.Close()
	}
	if err != nil {
		a.renderer.Error(w, http.StatusInternalServerError, fmt.Errorf(`Error while creating or opening file: %v`, err))
	} else if _, err = io.Copy(hostFile, uploadedFile); err != nil {
		a.renderer.Error(w, http.StatusInternalServerError, fmt.Errorf(`Error while writing file: %v`, err))
	} else {
		a.GetDir(w, config, path.Dir(filename), r.URL.Query().Get(`d`), &provider.Message{Level: `success`, Content: fmt.Sprintf(`File %s successfully uploaded`, uploadedFileHeader.Filename)})
	}
}