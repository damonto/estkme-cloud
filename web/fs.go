package web

import "embed"

//go:embed dist/*
var Root embed.FS
