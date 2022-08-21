// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"

	"http-server/pkg"
	"http-server/static"
)

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
			g := pkg.NewGzipWriter(w)
			defer g.Close()
			next.ServeHTTP(g, r)
		})
}

func writeError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func templateServe(w http.ResponseWriter, n, t string, c interface{}) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New(n).Parse(t))
	err := tmpl.Execute(w, c)
	writeError(w, err)
}

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		switch {
		case path == "app.js":
			w.Header().Set("Content-Type", "application/javascript")
			w.Write(static.AppJS())
		case len(path) > 0 && !strings.Contains(path, "/"):
			if c, ok := pkg.Coaches()[path]; ok {
				templateServe(w, "coach", static.Coach(), c)
			} else {
				http.NotFound(w, r)
			}
		case len(path) > 0 && strings.Count(path, "/") == 1:
			if c, ok := pkg.Coaches()[strings.Split(path, "/")[0]]; ok {
				switch {
				case strings.HasSuffix(path, teamsJson):
					w.Header().Set("Content-Type", "application/json")

					t, err := c.GetTeams()
					if err != nil {
						writeError(w, err)
					} else {
						err = json.NewEncoder(w).Encode(t)
						writeError(w, err)
					}
				case strings.HasSuffix(path, teamsIcs):
					w.Header().Set("Content-Type", "text/calendar")

					t, err := c.GetTeamsCalendar()
					if err != nil {
						writeError(w, err)
					} else {
						err = t.SerializeTo(w)
					}
				default:
					http.NotFound(w, r)
				}
			}
		default:
			templateServe(w, "index", static.Index(), pkg.Coaches())
		}
	})

	http.Handle("/favicon/", http.FileServer(http.FS(static.Favicon())))

	log.Println("listening on port 8080")
	log.Fatalln(http.ListenAndServe(":8080",
		logger(gzipper(setCacheHeader(http.DefaultServeMux)))))
}
