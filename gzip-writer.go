package main

import (
	"compress/gzip"
	"log"
	"net/http"
)

type (
	gzipWriter struct {
		http.ResponseWriter
		gz *gzip.Writer
	}
)

var (
	_ http.ResponseWriter = (*gzipWriter)(nil)
	_ http.Flusher        = (*gzipWriter)(nil)
)

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		gz:             gzip.NewWriter(w),
	}
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	return g.gz.Write(b)
}

func (g *gzipWriter) close() {
	err := g.gz.Close()
	if err != nil {
		log.Println(err)
	}
}

func (g *gzipWriter) Flush() {
	err := g.gz.Flush()
	if err != nil {
		log.Println(err)
	}
}
