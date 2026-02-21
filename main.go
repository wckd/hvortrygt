package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 || v > 65535 {
			log.Fatalf("invalid PORT: %q", p)
		}
		*port = v
	}

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	cache := NewCache()
	defer cache.Close()
	handler := newMux(staticFS, cache)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
