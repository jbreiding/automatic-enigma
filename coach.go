package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"
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
		Name        string
		Flight      string
		Application string
	}

	coach struct {
		Name  string
		Path  string
		Teams []team
	}
)

const (
	linkFmt    = "http://wys-2022pc.sportsaffinity.com/tour/public/info/ischedule.aspx?flightguid=%s&tournamentguid=%s&tourappguid=%s&gametimezone=pacific"
	tournament = "9DCC3624-BBFE-4F71-A8CA-0452DB26E0FF"
)

var coaches = map[string]coach{
	"coach-nigel": {
		Name: "Coach Nigel",
		Path: "coach-nigel",
		Teams: []team{
			{
				Name:        "XF B13 RCL 5",
				Flight:      "9EFDD764-318A-4E93-8F98-D785F3BE3C57",
				Application: "B344602F-0179-4E89-B1DE-9AD9643C2E51",
			},
			{
				Name:        "XF B13 RCL 6",
				Flight:      "167495CC-794E-443F-96C9-623BD9F27A06",
				Application: "E009ECC2-9F8D-497F-9083-16C28172AE3D",
			},
			{
				Name:        "XF B11 RCL 5",
				Flight:      "30C1EC64-7BAC-45AB-BC01-D5713B224BD4",
				Application: "E50870BC-674B-4992-87E5-49E10A0BB54E",
			},
		},
	},
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
		resp, err := http.Get(getTeamLink(t.Flight, t.Application))
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
		resp, err := http.Get(getTeamLink(t.Flight, t.Application))
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

func getTeamLink(f, a string) string {
	return fmt.Sprintf(linkFmt, f, tournament, a)
}
