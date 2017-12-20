package crud

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/fibr/utils"
	"github.com/ViBiOh/httputils"
)

const maxUploadSize = 32 * 1024 * 2014 // 32 MB

func getFileForm(w http.ResponseWriter, r *http.Request) (io.ReadCloser, *multipart.FileHeader, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

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

// Create given path directory to filesystem
func CreateDir(w http.ResponseWriter, r *http.Request, directory string) {
	if strings.HasSuffix(r.URL.Path, `/`) {
		filename, _ := utils.GetPathInfo(directory, r.URL.Path)

		if err := os.MkdirAll(filename, 0700); err != nil {
			httputils.InternalServerError(w, fmt.Errorf(`Error while creating directory: %v`, err))
		} else {
			w.WriteHeader(http.StatusCreated)
		}
	} else {
		httputils.Forbidden(w)
	}
}

// Save form file to filesystem
func SaveFile(w http.ResponseWriter, r *http.Request, directory string) {
	uploadedFile, uploadedFileHeader, err := getFileForm(w, r)
	if uploadedFile != nil {
		defer uploadedFile.Close()
	}
	if err != nil {
		httputils.BadRequest(w, fmt.Errorf(`Error while getting file from form: %v`, err))
		return
	}

	hostFile, err := createOrOpenFile(utils.GetPathInfo(directory, r.URL.Path, uploadedFileHeader.Filename))
	if hostFile != nil {
		defer hostFile.Close()
	}
	if err != nil {
		httputils.InternalServerError(w, fmt.Errorf(`Error while creating or opening file: %v`, err))
		return
	}

	if _, err = io.Copy(hostFile, uploadedFile); err != nil {
		httputils.InternalServerError(w, fmt.Errorf(`Error while writing file: %v`, err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
