package sdlhardware

import "github.com/veandco/go-sdl2/sdl"

func setRenderDriver() {
	sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl")
}
