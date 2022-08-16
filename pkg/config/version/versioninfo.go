package version

import "github.com/veandco/go-sdl2/sdl"

var (
	NAME   = "Flare"
	ENGINE = Version{1, 12, 12}
	MIN    = Version{0, 0, 0}
	MAX    = Version{65535, 65535, 65535}
)

func CreateVersionStringFull() string {
	// example output: Flare 1.0 (Linux)
	return NAME + " " + ENGINE.GetString() + " (" + sdl.GetPlatform() + ")"
}
