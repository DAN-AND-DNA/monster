package platform

import (
	"monster/pkg/common"
	"monster/pkg/common/define/platform"
	"monster/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

type Platform struct {
	hasExitButton       bool
	isMobileDevice      bool
	forceHardwareCursor bool
	hasLockFile         bool
	fullscreenBypass    bool
	configMenuType      int
	configVideo         []bool
	configAudio         []bool
	configInterface     []bool
	configInput         []bool
	configMisc          []bool
	//TODO
}

func New() *Platform {
	p := &Platform{
		hasExitButton:       true,
		isMobileDevice:      false,
		forceHardwareCursor: false,
		hasLockFile:         true,
		fullscreenBypass:    false,
		configMenuType:      platform.CONFIG_MENU_TYPE_DESKTOP,
		configVideo:         make([]bool, platform.VIDEO_COUNT),
		configAudio:         make([]bool, platform.AUDIO_COUNT),
		configInterface:     make([]bool, platform.INTERFACE_COUNT),
		configInput:         make([]bool, platform.INPUT_COUNT),
		configMisc:          make([]bool, platform.MISC_COUNT),
	}

	for i := 0; i < platform.VIDEO_COUNT; i++ {
		p.configVideo[i] = true
	}
	for i := 0; i < platform.AUDIO_COUNT; i++ {
		p.configAudio[i] = true
	}
	for i := 0; i < platform.INTERFACE_COUNT; i++ {
		p.configInterface[i] = true
	}
	for i := 0; i < platform.INPUT_COUNT; i++ {
		p.configInput[i] = true
	}
	for i := 0; i < platform.MISC_COUNT; i++ {
		p.configMisc[i] = true
	}

	return p
}

func (this *Platform) GetHasExitButton() bool {
	return this.hasExitButton
}

func (this *Platform) GetHasLockfile() bool {
	return this.hasLockFile
}

func (this *Platform) SetPaths(settings common.Settings) error {
	// 1. set config path
	tmpEnv := os.Getenv("XDG_CONFIG_HOME")
	if tmpEnv != "" {
		// $XDG_CONFIG_HOME/flare/
		settings.SetPathConf(tmpEnv + "/flare/")
	} else {
		tmpEnv = os.Getenv("HOME")
		if tmpEnv != "" {
			// $HOME/.config/flare/
			settings.SetPathConf(tmpEnv + "/.config/flare/")
		} else {
			// ./config/
			settings.SetPathConf("./config/")
		}
	}

	// 2. create dir
	if err := utils.CreateDir(settings.GetPathConf()); err != nil {
		return err
	}

	// 3. set user path (save games)
	tmpEnv = os.Getenv("XDG_DATA_HOME")
	if tmpEnv != "" {
		// $XDG_DATA_HOME/flare/
		settings.SetPathUser(tmpEnv + "/flare/")
	} else {
		tmpEnv = os.Getenv("HOME")
		if tmpEnv != "" {
			// $HOME/.local/share/flare/
			settings.SetPathUser(tmpEnv + "/.local/share/flare/")
		} else {
			// ./userdata/
			settings.SetPathUser("./userdata/")
		}
	}

	// 4. create dir
	if err := utils.CreateDir(settings.GetPathUser()); err != nil {
		return err
	}

	if err := utils.CreateDir(settings.GetPathUser() + "mods/"); err != nil {
		return err
	}

	if err := utils.CreateDir(settings.GetPathUser() + "saves/"); err != nil {
		return err
	}

	// 5. data folder
	// if the user specified a data path, try to use it
	if settings.GetCustomPathData() != "" {
		if isExists, err := utils.PathExists(settings.GetCustomPathData()); err != nil {
			return err
		} else if isExists {
			settings.SetPathData(settings.GetCustomPathData())
			return nil
		}
	}

	// Check for the local data before trying installed ones.
	if isExists, err := utils.PathExists("./mods"); err != nil {
		return err
	} else if isExists {
		settings.SetPathData("./")
		return nil
	}

	// check $XDG_DATA_DIRS options
	// a list of directories in preferred order separated by :
	tmpEnv = strings.TrimSpace(os.Getenv("XDG_DATA_DIRS"))
	if tmpEnv != "" {
		for _, pathTest := range strings.Split(tmpEnv, ":") {
			if pathTest != "" {
				if isExists, err := utils.PathExists(pathTest); err != nil {
					return err
				} else if isExists {
					settings.SetPathData(pathTest + "/flare/")
					return nil
				}
			}
		}
	}

	for _, tmpPath := range []string{
		"/usr/local/share/flare/",
		"/usr/share/flare/",
		"/usr/local/share/games/flare/",
		"/usr/share/games/flare/"} {
		if isExists, err := utils.PathExists(tmpPath); err == nil && !isExists {
			// just pass
			continue
		} else if err != nil {
			return err
		} else {
			settings.SetPathData(tmpPath)
			return nil
		}
	}

	// finally assume the local folder
	if absPath, err := os.Readlink("/proc/self/exe"); err == nil {
		settings.SetPathData(filepath.Dir(absPath))
	} else {
		settings.SetPathData("./")
	}

	return nil
}

func (this *Platform) GetIsMobileDevice() bool {
	return this.isMobileDevice
}

func (this *Platform) FullscreenBypass() bool {
	return this.fullscreenBypass
}

func (this *Platform) GetConfigMenuType() int {
	return this.configMenuType
}

func (this *Platform) GetConfigVideo() []bool {
	ret := make([]bool, len(this.configVideo))

	for i, val := range this.configVideo {
		ret[i] = val
	}

	return ret
}
func (this *Platform) GetConfigAudio() []bool {
	ret := make([]bool, len(this.configAudio))

	for i, val := range this.configAudio {
		ret[i] = val
	}

	return ret
}
func (this *Platform) GetConfigInterface() []bool {
	ret := make([]bool, len(this.configInterface))

	for i, val := range this.configInterface {
		ret[i] = val
	}

	return ret
}

func (this *Platform) GetConfigInput() []bool {
	ret := make([]bool, len(this.configInput))

	for i, val := range this.configInput {
		ret[i] = val
	}

	return ret
}

func (this *Platform) GetConfigMisc() []bool {
	ret := make([]bool, len(this.configMisc))

	for i, val := range this.configMisc {
		ret[i] = val
	}

	return ret
}

//========================
func (this *Platform) SetFullscreen(bool)  {}
func (this *Platform) FSCommit()           {}
func (this *Platform) SetExitEventFilter() {}
