package lockfile

import (
	"bufio"
	"bytes"
	"fmt"
	"monster/pkg/common"
	"monster/pkg/utils"
	"os"
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	defaultLockFile = new()
)

type LockFile struct {
	lockIndex int
}

func new() *LockFile {
	return &LockFile{
		lockIndex: 0,
	}
}

func (this *LockFile) lockFileRead(platform common.Platform, settings common.Settings) error {
	p := platform
	s := settings

	if !p.GetHasLockfile() {
		return nil
	}

	lockFilePath := s.GetPathConf() + "flare_lock"
	f, err := os.Open(lockFilePath)
	if err != nil {
		return err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	bytesComment := []byte("#")
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 || bytes.HasPrefix(bytes.TrimSpace(scanner.Bytes()), bytesComment) {
			continue
		}
		this.lockIndex, err = strconv.Atoi(scanner.Text())
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if this.lockIndex < 0 {
		this.lockIndex = 0
	}
	return nil
}

func (this *LockFile) lockFileWrite(platform common.Platform, settings common.Settings, increment int) error {
	p := platform
	s := settings
	var err error

	if !p.GetHasLockfile() {
		return nil
	}

	lockFilePath := s.GetPathConf() + "flare_lock"
	if increment < 0 {
		if this.lockIndex == 0 {
			return nil
		}

		if err = this.lockFileRead(p, s); err != nil && !utils.IsNotExist(err) {
			return err
		}
	}

	f, err := os.OpenFile(lockFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	this.lockIndex += increment
	fmt.Fprintf(f, "# Flare lock file. Counts instances of Flare\n")
	fmt.Fprintln(f, strconv.Itoa(this.lockIndex))

	return nil
}

func (this *LockFile) lockFileCheck(platform common.Platform, settings common.Settings) error {
	p := platform
	s := settings
	var err error

	if p.GetHasLockfile() {
		return nil
	}

	this.lockIndex = 0
	if err = this.lockFileRead(p, s); err != nil {
		return err
	}

	if this.lockIndex > 0 {
		buttons := []sdl.MessageBoxButtonData{
			{sdl.MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT | sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, 0, "Quit"},
			{0, 1, "Conintue"},
			{0, 2, "Reset"},
			{0, 3, "Safe Video"},
		}
		messageBoxData := sdl.MessageBoxData{
			sdl.MESSAGEBOX_INFORMATION,
			nil,
			"Flare",
			"Flare is unable to launch properly. This may be because it did not exit properly, or because there is another instance running.\n\nIf Flare crashed, it is recommended to try 'Safe Video' mode. This will try launching Flare with the minimum video settings.\n\nIf Flare is already running, you may:\n- 'Quit' Flare (safe, recommended)\n- 'Continue' to launch another copy of Flare.\n- 'Reset' the counter which tracks the number of copies of Flare that are currently running.\n  If this dialog is shown every time you launch Flare, this option should fix it.",
			buttons,
			nil,
		}
		buttonid, err := sdl.ShowMessageBox(&messageBoxData)
		if err != nil {
			return err
		}

		switch buttonid {
		case 0:
			if err = this.lockFileWrite(p, s, 1); err != nil {
				return err
			}
			return common.Err_normal_exit
		case 2:
			this.lockIndex = 0
		case 3:
			this.lockIndex = 0
			settings.SetSafeVideo(true)
		}
	}

	if err = this.lockFileWrite(p, s, 1); err != nil {
		return err
	}

	return nil
}

func (this *LockFile) close(platform common.Platform, settings common.Settings) error {
	return this.lockFileWrite(platform, settings, -1)
}

func Close(platform common.Platform, settings common.Settings) error {
	return defaultLockFile.close(platform, settings)
}

func LockFileWrite(platform common.Platform, settings common.Settings, increment int) error {
	return defaultLockFile.lockFileWrite(platform, settings, increment)
}

func LockFileCheck(platform common.Platform, settings common.Settings) error {
	return defaultLockFile.lockFileCheck(platform, settings)
}
