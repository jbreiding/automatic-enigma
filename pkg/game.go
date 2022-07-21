package pkg

import (
	"fmt"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"
)

type game struct {
	Home     bool
	Team     string
	Opponent string
	Time     time.Time
	Location string
	Venue    string
}

func NewGame(event *ical.VEvent, teamName string) (game, error) {
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
		Venue:    getVenue(event),
	}

	if !game.Home {
		game.Opponent = team1
	}

	return game, nil
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

func getVenue(event *ical.VEvent) string {
	description := event.GetProperty(ical.ComponentPropertyDescription).Value

	return strings.SplitN(strings.SplitN(description, "=", 2)[1], "\\n", 2)[0]
}
