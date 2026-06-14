package httpservers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/jamesread/data-cleaner/internal/api"
	"github.com/jamesread/data-cleaner/internal/config"
	"github.com/jamesread/data-cleaner/internal/frontend"
	log "github.com/sirupsen/logrus"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func listenAddr() string {
	p := strings.TrimSpace(os.Getenv("PORT"))
	if p == "" {
		if c := config.GetConfig(); c != nil && c.Network != nil {
			if addr := strings.TrimSpace(c.Network.BindProxy); addr != "" {
				return addr
			}
		}
		return ":8080"
	}
	if strings.HasPrefix(p, ":") {
		return p
	}
	if strings.Contains(p, ":") {
		return p
	}
	return ":" + p
}

func Start() {
	mux := http.NewServeMux()

	server := api.NewServer()
	apipath, apihandler := server.ConnectHandler()

	log.Infof("API path: /api/%s", apipath)

	mux.Handle("/api/download/", recoverHandler(http.StripPrefix("/api", http.HandlerFunc(server.DownloadCSV))))
	mux.Handle("/api"+apipath, recoverHandler(http.StripPrefix("/api", apihandler)))
	mux.Handle("/", http.StripPrefix("/", frontend.GetNewHandler()))

	addr := listenAddr()
	log.Infof("HTTP server listening on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func recoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Errorf("panic handling %s: %v\n%s", r.URL.Path, rec, debug.Stack())
				writeConnectError(w, http.StatusInternalServerError, fmt.Sprintf("internal server error: %v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeConnectError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "internal",
		"message": message,
	})
}
