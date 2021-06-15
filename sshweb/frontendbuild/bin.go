package frontendbuild

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed dist
var server embed.FS

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func embedFolder(fsEmbed embed.FS, targetPath string) ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}

func MwServeFrontendFiles(r *gin.Engine) {
	r.Use(Serve("/", "/api/", embedFolder(server, "dist")))
}

// Serve Static returns a middleware handler that serves static files in the given directory.
func Serve(urlPrefix, apiPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	httpFileServer := http.FileServer(fs)
	if urlPrefix != "" {
		httpFileServer = http.StripPrefix(urlPrefix, httpFileServer)
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, apiPrefix) {
			//if the request api start with /api/
			//will not serve static frontend files
			return
		}
		if fs.Exists(urlPrefix, path) {
			httpFileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
