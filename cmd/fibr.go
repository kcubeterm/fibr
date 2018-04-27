package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/auth/pkg/auth"
	authProvider "github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/auth/pkg/provider/basic"
	authService "github.com/ViBiOh/auth/pkg/service"
	"github.com/ViBiOh/fibr/pkg/crud"
	"github.com/ViBiOh/fibr/pkg/provider"
	"github.com/ViBiOh/fibr/pkg/ui"
	"github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/datadog"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"golang.org/x/crypto/bcrypt"
)

var errEmptyAuthorizationHeader = errors.New(`Empty authorization header`)

func checkSharePassword(r *http.Request, share *crud.Share) error {
	header := r.Header.Get(`Authorization`)
	if header == `` {
		return errEmptyAuthorizationHeader
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, `Basic `))
	if err != nil {
		return fmt.Errorf(`Error while decoding basic authentication: %v`, err)
	}

	dataStr := string(data)

	sepIndex := strings.Index(dataStr, `:`)
	if sepIndex < 0 {
		return errors.New(`Error while reading basic authentication`)
	}

	password := dataStr[sepIndex+1:]
	if err := bcrypt.CompareHashAndPassword([]byte(share.Password), []byte(password)); err != nil {
		return errors.New(`Invalid credentials`)
	}

	return nil
}

func handleAnonymousRequest(w http.ResponseWriter, r *http.Request, err error, crudApp *crud.App, uiApp *ui.App) {
	if auth.IsForbiddenErr(err) {
		uiApp.Error(w, http.StatusForbidden, errors.New(`You're not authorized to do this ⛔️`))
	} else if !crudApp.CheckAndServeSEO(w, r) {
		if err == authProvider.ErrMalformedAuth || err == authProvider.ErrUnknownAuthType {
			uiApp.Error(w, http.StatusBadRequest, err)
		} else {
			w.Header().Add(`WWW-Authenticate`, `Basic charset="UTF-8"`)
			uiApp.Error(w, http.StatusUnauthorized, err)
		}
	}
}

func browserHandler(crudApp *crud.App, uiApp *ui.App, authApp *auth.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			uiApp.Error(w, http.StatusMethodNotAllowed, errors.New(`We don't understand what you want from us`))
			return
		}

		if strings.Contains(r.URL.Path, `..`) {
			uiApp.Error(w, http.StatusForbidden, errors.New(`You're not authorized to do this ⛔`))
			return
		}

		_, err := authApp.IsAuthenticated(r)

		config := &provider.RequestConfig{
			URL:      r.URL.Path,
			Path:     r.URL.Path,
			CanEdit:  true,
			CanShare: true,
		}

		if share := crudApp.GetSharedPath(config.Path); share != nil {
			config.Root = share.Path
			config.Path = strings.TrimPrefix(config.Path, fmt.Sprintf(`/%s`, share.ID))
			config.Prefix = share.ID

			if err != nil {
				config.CanEdit = share.Edit
				config.CanShare = false
			}

			if share.Password != `` && err != nil {
				if err := checkSharePassword(r, share); err != nil {
					w.Header().Add(`WWW-Authenticate`, `Basic realm="Password required, username optional" charset="UTF-8"`)
					uiApp.Error(w, http.StatusUnauthorized, err)
					return
				}
			}

			err = nil
		}

		if err != nil {
			handleAnonymousRequest(w, r, err, crudApp, uiApp)
		} else if r.Method == http.MethodGet {
			crudApp.Get(w, r, config, nil)
		} else if r.Method == http.MethodPost {
			crudApp.Post(w, r, config)
		} else if r.Method == http.MethodPut {
			crudApp.CreateDir(w, r, config)
		} else if r.Method == http.MethodPatch {
			crudApp.Rename(w, r, config)
		} else if r.Method == http.MethodDelete {
			crudApp.Delete(w, r, config)
		} else {
			httperror.NotFound(w)
		}
	})
}

func main() {
	owaspConfig := owasp.Flags(``)
	authConfig := auth.Flags(`auth`)
	basicConfig := basic.Flags(`basic`)
	crudConfig := crud.Flags(``)
	uiConfig := ui.Flags(``)
	datadogConfig := datadog.Flags(`datadog`)

	httputils.NewApp(httputils.Flags(``), func() http.Handler {
		authApp := auth.NewApp(authConfig, authService.NewBasicApp(basicConfig))
		uiApp := ui.NewApp(uiConfig, *crudConfig[`directory`].(*string))
		crudApp := crud.NewApp(crudConfig, uiApp)

		serviceHandler := owasp.Handler(owaspConfig, browserHandler(crudApp, uiApp, authApp))
		healthHandler := healthcheck.Handler()

		return datadog.NewApp(datadogConfig).Handler(gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == `/health` {
				healthHandler.ServeHTTP(w, r)
			} else {
				serviceHandler.ServeHTTP(w, r)
			}
		})))
	}, nil).ListenAndServe()
}