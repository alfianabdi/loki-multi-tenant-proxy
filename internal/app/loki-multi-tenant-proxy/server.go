package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/angelbarrera92/loki-multi-tenant-proxy/internal/pkg"
	"github.com/gorilla/mux"
	"github.com/urfave/cli"
)

// Serve serves
func Serve(c *cli.Context) error {

	lokiServerDistributorURL, _ := url.Parse(c.String("loki-server-distributor"))
	lokiServerQuerierURL, _ := url.Parse(c.String("loki-server-querier"))
	lokiServerQueryFrontendURL, _ := url.Parse(c.String("loki-server-queryfrontend"))

	serveAt := fmt.Sprintf(":%d", c.Int("port"))
	authConfigLocation := c.String("auth-config")
	authConfig, _ := pkg.ParseConfig(&authConfigLocation)

	rtr := mux.NewRouter()

	// Distributor API
	rtr.HandleFunc("/api/prom/push", createHandler(lokiServerDistributorURL, authConfig))
	rtr.HandleFunc("/loki/api/v1/push", createHandler(lokiServerDistributorURL, authConfig))
	// Querier API
	rtr.HandleFunc("/api/prom/tail", createHandler(lokiServerQuerierURL, authConfig))
	rtr.HandleFunc("/loki/api/v1/tail", createHandler(lokiServerQuerierURL, authConfig))
	// Query Frontend API

	rtr.PathPrefix("/api/prom/").Handler(createHandler(lokiServerQueryFrontendURL, authConfig))
	rtr.PathPrefix("/loki/api/").Handler(createHandler(lokiServerQueryFrontendURL, authConfig))

	http.Handle("/", rtr)
	if err := http.ListenAndServe(serveAt, nil); err != nil {
		log.Fatalf("Loki multi tenant proxy can not start %v", err)
		return err
	}
	return nil
}

func createHandler(lokiServerURL *url.URL, authConfig *pkg.Authn) http.HandlerFunc {
	reverseProxy := httputil.NewSingleHostReverseProxy(lokiServerURL)
	return LogRequest(BasicAuth(ReverseLoki(reverseProxy, lokiServerURL), authConfig))
}
