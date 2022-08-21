package static

import "embed"

//go:embed coach.gotmpl
var coach string

//go:embed index.gotmpl
var index string

//go:embed favicon/*
var favicon embed.FS

//go:embed app.js
var appJS []byte

func Coach() string {
	return coach
}

func Favicon() embed.FS {
	return favicon
}

func Index() string {
	return index
}

func AppJS() []byte {
	return appJS
}
