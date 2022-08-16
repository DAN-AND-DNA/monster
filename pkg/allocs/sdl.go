package allocs

import (
	"fmt"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func (this *Allocs) SdlCreateWindow(title string, x, y, w, h int32, flags uint32) (*sdl.Window, error) {
	ptr, err := sdl.CreateWindow(title, x, y, w, h, flags)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_WINDOW, id)

	return ptr, nil
}

func (this *Allocs) SdlCreateRenderer(window *sdl.Window, index int, flags uint32) (*sdl.Renderer, error) {
	ptr, err := sdl.CreateRenderer(window, index, flags)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_RENDERER, id)

	return ptr, nil
}

func (this *Allocs) SdlCreateTexture(renderer *sdl.Renderer, format uint32, access int, w, h int32) (*sdl.Texture, error) {
	ptr, err := renderer.CreateTexture(format, access, w, h)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_TEXTURE, id)

	return ptr, nil
}

func (this *Allocs) ImgLoad(file string) (*sdl.Surface, error) {
	ptr, err := img.Load(file)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_SURFACE, id)

	return ptr, nil
}

func (this *Allocs) ImgLoadTexture(renderer *sdl.Renderer, file string) (*sdl.Texture, error) {
	ptr, err := img.LoadTexture(renderer, file)
	if err != nil {
		return nil, err
	}

	if ptr == nil {
		return nil, nil
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_TEXTURE, id)

	return ptr, nil
}

func (this *Allocs) FontRenderUTF8Blended(font *ttf.Font, text string, color sdl.Color) (*sdl.Surface, error) {
	ptr, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_SURFACE, id)

	return ptr, nil
}

func (this *Allocs) FontRenderUTF8Solid(font *ttf.Font, text string, color sdl.Color) (*sdl.Surface, error) {
	ptr, err := font.RenderUTF8Solid(text, color)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_SURFACE, id)

	return ptr, nil
}

func (this *Allocs) SdlCreateRGBSurface(flags uint32, width, height, depth int32, Rmask, Gmask, Bmask, Amask uint32) (*sdl.Surface, error) {
	ptr, err := sdl.CreateRGBSurface(flags, width, height, depth, Rmask, Gmask, Bmask, Amask)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_SURFACE, id)

	return ptr, nil
}

func (this *Allocs) SdlCreateTextureFromSurface(renderer *sdl.Renderer, surface *sdl.Surface) (*sdl.Texture, error) {
	ptr, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_TEXTURE, id)

	return ptr, err
}

func (this *Allocs) TtfOpenFont(file string, size int) (*ttf.Font, error) {
	ptr, err := ttf.OpenFont(file, size)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%p", ptr)
	//id := *(*uint64)(unsafe.Pointer(ptr))

	this.register(SDL_FONT, id)

	return ptr, err
}

// 对外api
func SdlCreateWindow(title string, x, y, w, h int32, flags uint32) (*sdl.Window, error) {
	return defaultAllocs.SdlCreateWindow(title, x, y, w, h, flags)
}

func SdlCreateRenderer(window *sdl.Window, index int, flags uint32) (*sdl.Renderer, error) {
	return defaultAllocs.SdlCreateRenderer(window, index, flags)
}

func SdlCreateTexture(renderer *sdl.Renderer, format uint32, access int, w, h int32) (*sdl.Texture, error) {
	return defaultAllocs.SdlCreateTexture(renderer, format, access, w, h)
}

func ImgLoad(file string) (*sdl.Surface, error) {
	return defaultAllocs.ImgLoad(file)
}

func ImgLoadTexture(renderer *sdl.Renderer, file string) (*sdl.Texture, error) {
	return defaultAllocs.ImgLoadTexture(renderer, file)
}

func FontRenderUTF8Blended(font *ttf.Font, text string, color sdl.Color) (*sdl.Surface, error) {
	return defaultAllocs.FontRenderUTF8Blended(font, text, color)
}

func FontRenderUTF8Solid(font *ttf.Font, text string, color sdl.Color) (*sdl.Surface, error) {
	return defaultAllocs.FontRenderUTF8Solid(font, text, color)
}

func SdlCreateRGBSurface(flags uint32, width, height, depth int32, Rmask, Gmask, Bmask, Amask uint32) (*sdl.Surface, error) {
	return defaultAllocs.SdlCreateRGBSurface(flags, width, height, depth, Rmask, Gmask, Bmask, Amask)
}

func SdlCreateTextureFromSurface(renderer *sdl.Renderer, surface *sdl.Surface) (*sdl.Texture, error) {
	return defaultAllocs.SdlCreateTextureFromSurface(renderer, surface)
}

func TtfOpenFont(file string, size int) (*ttf.Font, error) {
	return defaultAllocs.TtfOpenFont(file, size)
}
