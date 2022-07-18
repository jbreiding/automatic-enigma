package pkg

type team struct {
	name        string
	flight      string
	application string
}

func NewTeam(n, f, a string) team {
	return team{
		name:        n,
		flight:      f,
		application: a,
	}
}

func (t team) Name() string {
	return t.name
}

func (t team) Flight() string {
	return t.flight
}

func (t team) Application() string {
	return t.application
}
