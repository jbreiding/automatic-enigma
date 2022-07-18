package static

import "embed"

//go:embed coach.html
var coachBytes []byte

//go:embed index.html
var indexBytes []byte

//go:embed favicon/*
var favicon embed.FS

func Coach() string {
	return string(coachBytes)
}

func Favicon() embed.FS {
	return favicon
}

func Index() string {
	return string(indexBytes)
}
