package crud

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/ViBiOh/fibr/pkg/provider"
)

// Create creates given path directory to filesystem
func (a *App) Create(w http.ResponseWriter, r *http.Request, request *provider.Request) {
	if !request.CanEdit {
		a.renderer.Error(w, http.StatusForbidden, ErrNotAuthorized)
		return
	}

	pathname, err := getFilepath(r, request)
	if err != nil {
		a.renderer.Error(w, http.StatusForbidden, ErrNotAuthorized)
		return
	}

	pathname, err = provider.SanitizeName(pathname)
	if err != nil {
		a.renderer.Error(w, http.StatusBadRequest, fmt.Errorf(`error while sanitizing name: %v`, err))
		return
	}

	if err := a.storage.Create(pathname); err != nil {
		a.renderer.Error(w, http.StatusInternalServerError, fmt.Errorf(`error while creating directory: %v`, err))
		return
	}

	http.Redirect(w, r, fmt.Sprintf(`%s/?message=%s&messageLevel=success`, pathname, url.QueryEscape(fmt.Sprintf(`Directory %s successfully created`, path.Base(pathname)))), http.StatusMovedPermanently)
}
