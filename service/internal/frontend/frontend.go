package frontend

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesread/golure/pkg/dirs"
)

func findWebuiDir() string {
	directoriesToSearch := []string{
		"../frontend/dist/",
		"/app/frontend/dist/",
	}

	dir, err := dirs.GetFirstExistingDirectory("webui", directoriesToSearch)

	if err != nil {
		panic("Failed to find webui directory (run `make -C frontend prod` to build the UI): " + err.Error())
	}

	return dir
}

func GetNewHandler() http.Handler {
	root := findWebuiDir()
	fileServer := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/assets/") && !isStaticFile(filepath.Join(root, r.URL.Path)) {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func isStaticFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
