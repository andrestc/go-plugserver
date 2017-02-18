package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"
)

func main() {
	mux := NewServeMux()
	log.Println("Starting server on port :8080")
	go watchHandlers(mux, "./handlers")
	http.ListenAndServe(":8080", mux)
}

func watchHandlers(mux *ServeMux, pluginsDir string) {
	for {
		select {
		case <-time.After(5 * time.Second):
			log.Println("Refreshing handlers...")
			loadHandlers(mux, pluginsDir)
			log.Println("Done refreshing handlers...")
		}
	}
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
			continue
		}
		fun, ok := handler.(func() (string, func(w http.ResponseWriter, r *http.Request)))
		if !ok {
			log.Printf("invalid Handler function on plugin %q, has type %T", f.Name(), handler)
			continue
		}
		pattern, handleFunc := fun()
		log.Printf("Registering handler for pattern %q from file %q", pattern, f.Name())
		mux.HandleFunc(pattern, handleFunc)
	}
}
