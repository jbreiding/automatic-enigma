// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed static/coach.html
var coachBytes []byte

//go:embed static/index.html
var indexBytes []byte

//go:embed coaches.yaml
var configBytes []byte

//go:embed static/favicon/*
var favicon embed.FS

const (
	teamsJson = "teams.json"
	teamsIcs  = "teams.ics"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%+v", r)
			next.ServeHTTP(w, r)
			// log.Printf("%+v", w)
		})
}

func setCacheHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=14400")
			next.ServeHTTP(w, r)
		})
}

func gzipper(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Encoding", "gzip")
			g := newGzipWriter(w)
			defer g.close()
			next.ServeHTTP(g, r)
		})
}

func main() {
	loadConfig()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch path := r.URL.Path; {
		case len(path) <= 1:
			w.Header().Set("Content-Type", "text/html")

			tmpl := template.Must(template.New("index").Parse(string(indexBytes)))
			err := tmpl.Execute(w, coaches)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			switch nodes := strings.Split(path[1:], "/"); {
			case len(nodes) == 1:
				if c, ok := coaches[nodes[0]]; ok {
					w.Header().Set("Content-Type", "text/html")

					tmpl := template.Must(template.New("coach").Parse(string(coachBytes)))
					err := tmpl.Execute(w, c)
					if err != nil {
						log.Println(err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				} else {
					http.NotFound(w, r)
				}
			case len(nodes) == 2:
				if c, ok := coaches[nodes[0]]; ok {
					switch file := nodes[1]; {
					case strings.HasSuffix(file, teamsJson):
						w.Header().Set("Content-Type", "application/json")

						t, err := c.getTeams()
						if err != nil {
							log.Println(err)
							http.Error(w, err.Error(), http.StatusInternalServerError)
						} else {
							err = json.NewEncoder(w).Encode(t)
							if err != nil {
								log.Println(err)
								http.Error(w, err.Error(), http.StatusInternalServerError)
							}
						}
					case strings.HasSuffix(file, teamsIcs):
						w.Header().Set("Content-Type", "text/calendar")

						t, err := c.getTeamsCalendar()
						if err != nil {
							log.Println(err)
							http.Error(w, err.Error(), http.StatusInternalServerError)
						} else {
							err = t.SerializeTo(w)
							if err != nil {
								log.Println(err)
								http.Error(w, err.Error(), http.StatusInternalServerError)
							}
						}
					default:
						http.NotFound(w, r)
					}
				}
			default:
				http.NotFound(w, r)
			}
		}
	})

	fSys, err := fs.Sub(favicon, "static")
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/favicon/", http.FileServer(http.FS(fSys)))

	log.Println("listening on port 8080")
	log.Fatalln(http.ListenAndServe(":8080",
		logger(gzipper(setCacheHeader(http.DefaultServeMux)))))
}
