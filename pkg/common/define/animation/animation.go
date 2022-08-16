package animation

const (
	ANIMTYPE_NONE       = iota
	ANIMTYPE_PLAY_ONCE  // just iterates over the images one time. it holds the final image when finished.
	ANIMTYPE_LOOPED     // going over the images again and again.
	ANIMTYPE_BACK_FORTH // iterate from index=0 to maxframe and back again. keeps holding the first image afterwards.
)
