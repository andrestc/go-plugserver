package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func main() {
	mux := NewServeMux()
	log.Println("Starting server on port :8080")
	go loadHandlers(mux, "./handlers")
	http.ListenAndServe(":8080", mux)
}

func loadHandlers(mux *ServeMux, pluginsDir string) {
	dir, err := os.Open(pluginsDir)
	if err != nil {
		panic(fmt.Sprintf("unable to open plugins dir: %s", err))
	}
	files, err := dir.Readdir(0)
	if err != nil {
		log.Printf("failed to read plugins dir content: %s", err)
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".so") {
			continue
		}
		p, err := plugin.Open(filepath.Join(dir.Name(), f.Name()))
		if err != nil {
			log.Printf("failed to read plugin file: %s", err)
			continue
		}
		handler, err := p.Lookup("Handler")
		if err != nil {
			log.Printf("failed to lookup handler in plugin %s: %s", f.Name(), err)
		}
		pattern, handleFunc := handler.(func() (string, func(w http.ResponseWriter, r *http.Request)))()
		mux.HandleFunc(pattern, handleFunc)
	}
}
