// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

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
			defer func(t time.Time) {
				log.Printf("%s request time %v", r.URL.Path, time.Since(t))
			}(time.Now())

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

	http.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(static.AppJS())
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		templateServe(w, "index", static.Index(), pkg.Coaches())
	})

	for _, c := range pkg.Coaches() {
		http.HandleFunc(fmt.Sprintf("/%s", c.Path), func(w http.ResponseWriter, r *http.Request) {
			templateServe(w, "coach", static.Coach(), c)
		})

		http.HandleFunc(fmt.Sprintf("/%s/%s", c.Path, teamsJson), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			t, err := c.GetTeams()
			if err != nil {
				writeError(w, err)
			} else {
				err = json.NewEncoder(w).Encode(t)
				writeError(w, err)
			}
		})

		http.HandleFunc(fmt.Sprintf("/%s/%s", c.Path, teamsIcs), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/calendar")

			t, err := c.GetTeamsCalendar()
			if err != nil {
				writeError(w, err)
			} else {
				err = t.SerializeTo(w)
				writeError(w, err)
			}
		})
	}

	http.Handle("/favicon/", http.FileServer(http.FS(static.Favicon())))

	log.Println("listening on port 8080")
	log.Fatalln(http.ListenAndServe(":8080",
		logger(gzipper(setCacheHeader(http.DefaultServeMux)))))
}
