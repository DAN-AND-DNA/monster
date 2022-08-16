package common

import "errors"

const (
	Code_ok    = 0
	Code_error = 1000
)

var (
	Err_normal_exit                  = errors.New("normal exit")
	Err_no_fallbackmod               = errors.New("no fallbackmod")
	Err_sdl_init_failed              = errors.New("sdl init failed")
	Err_no_such_file_or_dir          = errors.New("no such file or dir")
	Err_bad_modManager               = errors.New("bad mod manager")
	Err_modmanager_load_mod_failed   = errors.New("mod manager load mod failed")
	Err_bad_shared_resources         = errors.New("bad shared resources")
	Err_bad_key_in_settings          = errors.New("bad key in settings")
	Err_bad_val_in_enginesettings    = errors.New("bad val in engine settings")
	Err_bad_key_in_enginesettings    = errors.New("bad key in engine settings")
	Err_bad_key_in_fontengine        = errors.New("bad key in font engine")
	Err_bad_args_in_sdlhardwareimage = errors.New("bad args in sdl hardware image")
	Err_bad_key_in_gameswitcher      = errors.New("bad key in game switcher")
)
