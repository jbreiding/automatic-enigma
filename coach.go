package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	ical "github.com/arran4/golang-ical"
	"sigs.k8s.io/yaml"
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

var (
	coaches     map[string]coach
	coachesOnce sync.Once
)

func loadConfig() {
	coachesOnce.Do(func() {
		if err := yaml.Unmarshal(configBytes, &coaches); err != nil {
			log.Panicln(err)
		}
	})
}

func splitTeams(event *ical.VEvent) (string, string, error) {
	description := event.GetProperty(ical.ComponentPropertyDescription).Value

	header := strings.SplitN(description, "\\n \\n", 2)
	if len(header) != 2 {
		return "", "", fmt.Errorf("unable to find header, event description invalid %s", description)
	}

	teams := strings.SplitN(header[0], "\\n", 3)
	if len(teams) != 3 {
		return "", "", fmt.Errorf("unable to find teams, event description invalid %s", description)
	}

	return teams[0], teams[2], nil
}

func newGame(event *ical.VEvent, teamName string) (game, error) {
	team1, team2, err := splitTeams(event)
	if err != nil {
		return game{}, err
	}

	start, err := event.GetStartAt()
	if err != nil {
		return game{}, err
	}

	game := game{
		Home:     strings.EqualFold(team1, teamName),
		Team:     teamName,
		Time:     start,
		Location: event.GetProperty(ical.ComponentPropertyLocation).Value,
		Opponent: team2,
	}

	if !game.Home {
		game.Opponent = team1
	}

	return game, nil
}

func (c coach) getTeams() (map[string][]game, error) {
	teams := map[string][]game{}

	for _, t := range c.Teams {
		resp, err := http.Get(t.Link)
		if err != nil {
			return teams, err
		}
		defer resp.Body.Close()

		if cal, err := ical.ParseCalendar(resp.Body); err != nil {
			return teams, err
		} else {
			for _, entry := range cal.Events() {
				game, err := newGame(entry, t.Name)
				if err != nil {
					return teams, err
				}

				day := game.Time.Truncate(24 * time.Hour).Format("Jan 2, 2006")
				teams[day] = append(teams[day], game)
			}
		}
	}

	return teams, nil
}

func (c coach) getTeamsCalendar() (*ical.Calendar, error) {
	ics := ical.NewCalendarFor(c.Name)

	for _, t := range c.Teams {
		resp, err := http.Get(t.Link)
		if err != nil {
			return ics, err
		}
		defer resp.Body.Close()

		if cal, err := ical.ParseCalendar(resp.Body); err != nil {
			return ics, err
		} else {
			for _, entry := range cal.Events() {
				ics.AddVEvent(entry)
			}
		}
	}

	return ics, nil
}
