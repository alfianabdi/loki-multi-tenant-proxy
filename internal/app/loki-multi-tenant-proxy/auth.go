package proxy

import (
	"context"
	"crypto/subtle"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/angelbarrera92/loki-multi-tenant-proxy/internal/pkg"
)

type key int

const (
	// OrgIDKey Key used to pass loki tenant id though the middleware context
	OrgIDKey key = iota
	realm        = "Loki multi-tenant proxy"
)

// BasicAuth can be used as a middleware chain to authenticate users before proxying a request
func BasicAuth(handler http.HandlerFunc, authConfig *pkg.Authn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		orgID := r.Header.Get("X-Scope-OrgID")
		authorized := isAuthorized(user, pass, orgID, authConfig)
		if !ok || !authorized {
			writeUnauthorisedResponse(w)
			return
		}
		ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
		handler(w, r.WithContext(ctx))
	}
}

func isAuthorized(user string, pass string, orgID string, authConfig *pkg.Authn) bool {
	for _, v := range authConfig.Users {
		hashErr := bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(pass))
		if subtle.ConstantTimeCompare([]byte(user), []byte(v.Username)) == 1 && (subtle.ConstantTimeCompare([]byte(pass), []byte(v.Password)) == 1 || hashErr == nil) && subtle.ConstantTimeCompare([]byte(orgID), []byte(v.OrgID)) == 1 {
			return true
		}
	}
	return false
}

func writeUnauthorisedResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}
