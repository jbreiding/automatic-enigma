package pkg

import (
	"fmt"
	"net/http"
	"time"

	ical "github.com/arran4/golang-ical"
)

type (
	coach struct {
		Name  string
		Path  string
		Teams []team
	}
)

const (
	linkFmt    = "http://wys.affinitysoccer.com/tour/public/info/ischedule.aspx?flightguid=%s&tournamentguid=%s&tourappguid=%s&gametimezone=pacific"
	tournament = "5952FE5C-4709-4C13-8E39-13803690B102"
)

var coaches = map[string]coach{
	"coach-johnny": {
		Name: "Coach Johnny",
		Path: "coach-johnny",
		Teams: []team{
			NewTeam("XF B13 RCL 4", "5E120FFE-F973-49EB-A8C3-D171AADD1A98", "D199B903-6CDE-43EB-A48E-8DAE6C144BC1"),
			NewTeam("XF B13 RCL 5", "5E120FFE-F973-49EB-A8C3-D171AADD1A98", "26D457F5-F62A-4042-8D55-BBE09AA96331"),
			NewTeam("XF B11 RCL 5", "418CBC2A-7AF6-48F4-988C-7A07566F9AA4", "1F193B87-8A58-4D53-A436-BADE1D81ED88"),
		},
	},
}

func Coaches() map[string]coach {
	return coaches
}

func (c coach) GetTeams() (map[string][]game, error) {
	teams := map[string][]game{}

	for _, t := range c.Teams {
		resp, err := http.Get(getTeamLink(t.Flight(), t.Application()))
		if err != nil {
			return teams, err
		}
		defer resp.Body.Close()

		if cal, err := ical.ParseCalendar(resp.Body); err != nil {
			return teams, err
		} else {
			for _, entry := range cal.Events() {
				game, err := NewGame(entry, t.Name())
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

func (c coach) GetTeamsCalendar() (*ical.Calendar, error) {
	ics := ical.NewCalendarFor(c.Name)

	for _, t := range c.Teams {
		resp, err := http.Get(getTeamLink(t.Flight(), t.Application()))
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
