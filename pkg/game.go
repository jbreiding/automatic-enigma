package pkg

import (
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
	}

	if !game.Home {
		game.Opponent = team1
	}

	return game, nil
}
