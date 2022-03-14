// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	ical "github.com/arran4/golang-ical"
)

//go:embed static/index.html
var indexBytes []byte

//go:embed static/favicon/*
var favicon embed.FS

const (
	coachHeader           = "X-Coach"
	gzipEncoding          = "gzip"
	htmlType              = "text/html"
	jsonType              = "application/json"
	calendarType          = "text/calendar"
	cacheControl          = "public, max-age=14400"
	contentTypeHeader     = "Content-Type"
	contentEncodingHeader = "Content-Encoding"
	cacheControlHeader    = "Cache-Control"

	root      = "/"
	teamsJson = "/teams.json"
	teamsIcs  = "/teams.ics"
	fav       = "/favicon/"

	portEnv  = "PORT"
	coachEnv = "TEAMS_COACH"

	portDefault = "8080"
)

type (
	game struct {
		Home     bool
		Team     string
		Opponent string
		Time     time.Time
		Location string
	}

	team struct {
		Name string
		Link string
	}

	coach struct {
		Name  string
		Path  string
		Teams []team
	}
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println(r)
			next.ServeHTTP(w, r)
			log.Println(w)
		},
	)
}

func main() {
	log.Print("starting server...")

	http.HandleFunc(root, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, htmlType)
		w.Header().Set(contentEncodingHeader, gzipEncoding)
		w.Header().Set(cacheControlHeader, cacheControl)
		w.WriteHeader(http.StatusOK)
		c := getCoach(r.Header[coachHeader])
		renderIndex(w, c)
	})

	http.HandleFunc(teamsJson, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, jsonType)
		w.Header().Set(contentEncodingHeader, gzipEncoding)
		w.Header().Set(cacheControlHeader, cacheControl)
		w.WriteHeader(http.StatusOK)
		c := getCoach(r.Header[coachHeader])
		getTeams(w, c)
	})

	http.HandleFunc(teamsIcs, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, calendarType)
		w.Header().Set(contentEncodingHeader, gzipEncoding)
		w.Header().Set(cacheControlHeader, cacheControl)
		w.WriteHeader(http.StatusOK)
		c := getCoach(r.Header[coachHeader])
		getTeamsCalendar(w, c)
	})

	fSys, err := fs.Sub(favicon, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle(fav, http.FileServer(http.FS(fSys)))

	port := os.Getenv(portEnv)
	if port == "" {
		port = portDefault
		log.Printf("defaulting to port %s", port)
	}

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, Logger(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}

func newGame(event *ical.VEvent, team team) game {
	nodes := strings.Split(event.GetProperty(ical.ComponentPropertyDescription).Value, "\\n")

	start, err := event.GetStartAt()
	if err != nil {
		log.Fatal(err)
	}

	game := game{
		Home:     strings.EqualFold(nodes[0], team.Name),
		Team:     team.Name,
		Time:     start,
		Location: event.GetProperty(ical.ComponentPropertyLocation).Value,
	}

	if game.Home {
		game.Opponent = nodes[2]
	} else {
		game.Opponent = nodes[0]
	}

	return game
}

func getCoach(h []string) coach {
	var env string
	if len(h) > 0 {
		env = h[0]
	} else {
		env = os.Getenv(coachEnv)
	}

	if len(env) == 0 {
		log.Fatal("unable to find coach token")
	}

	dec := json.NewDecoder(
		base64.NewDecoder(
			base64.StdEncoding,
			bytes.NewBufferString(env),
		),
	)

	var c coach
	err := dec.Decode(&c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func getTeams(w io.Writer, c coach) {
	teams := map[string][]game{}

	for _, t := range c.Teams {
		resp, err := http.Get(t.Link)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		cal, err := ical.ParseCalendar(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range cal.Events() {
			game := newGame(entry, t)
			day := game.Time.Truncate(24 * time.Hour).Format("Jan 2, 2006")
			teams[day] = append(teams[day], game)
		}
	}

	gz := gzip.NewWriter(w)
	defer gz.Close()

	err := json.NewEncoder(gz).Encode(teams)

	if err != nil {
		log.Fatal(err)
	}
}

func renderIndex(w io.Writer, c coach) {

	tmpl, err := template.New("index").Parse(string(indexBytes))
	if err != nil {
		log.Fatal(err)
	}

	gz := gzip.NewWriter(w)
	defer gz.Close()

	err = tmpl.Execute(gz, c)
	if err != nil {
		log.Fatal(err)
	}
}

func getTeamsCalendar(w io.Writer, c coach) {
	ics := ical.NewCalendarFor(c.Name)
	ics.SetMethod(ical.MethodRequest)
	for _, t := range c.Teams {
		resp, err := http.Get(t.Link)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		cal, err := ical.ParseCalendar(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range cal.Events() {
			ics.AddVEvent(entry)
		}
	}

	gz := gzip.NewWriter(w)
	defer gz.Close()

	err := ics.SerializeTo(gz)
	if err != nil {
		log.Fatal(err)
	}
}
