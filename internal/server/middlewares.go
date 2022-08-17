package server

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/tinfoil-knight/gargoyle/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}

func useHeaderModifier(handler http.Handler, headerCfg config.HeaderCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &headerModifier{w, false, &headerCfg}
		handler.ServeHTTP(rw, r)
	})
}

func urlRewriter(handler http.Handler, rewriteCfg config.RewriteCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for from, to := range rewriteCfg {
			if from == r.URL.Path {
				r.URL.Path = to
				break
			}
		}
		handler.ServeHTTP(w, r)
	})
}

func auth(handler http.Handler, auth config.AuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: refactor
		if auth.BasicAuth != nil {
			user, pwd, ok := r.BasicAuth()
			hash, _ := auth.BasicAuth[user]
			if !ok || !check(hash, pwd) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
		if auth.KeyAuth != nil {
			header := auth.KeyAuth.Header
			cfgKey := auth.KeyAuth.Key
			reqKey := r.Header.Get(header)
			if cfgKey != reqKey {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		handler.ServeHTTP(w, r)
	})
}

func check(hash []byte, pwd string) bool {
	if len(hash) == 0 || pwd == "" {
		return false
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte(pwd)); err != nil {
		return false
	}
	return true
}

func applyMiddlewares(handler http.Handler, service config.ServiceCfg) http.Handler {
	if service.Header != nil {
		handler = useHeaderModifier(handler, *service.Header)
	}
	if service.Rewrite != nil {
		handler = urlRewriter(handler, *service.Rewrite)
	}
	if service.Auth != nil {
		handler = auth(handler, *service.Auth)
	}
	return logHTTPRequest(handler)
}
