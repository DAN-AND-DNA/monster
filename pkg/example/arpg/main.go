package main

import (
	"monster/pkg/allocs"
	"monster/pkg/common"
	"monster/pkg/config/version"
	"monster/pkg/filesystem/lockfile"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/game/state"
	"os"
	"os/signal"
	"syscall"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	done    = false
	donec   = make(chan os.Signal, 1)
	modules = NewModules()
)

func init() {
	signal.Notify(donec, syscall.SIGTERM)
	signal.Notify(donec, syscall.SIGINT)

	go func() {
		_ = <-donec
		allocs.PrintAllStack()
		done = true
	}()
}

func main() {
	s := modules.NewSettings()
	p := modules.NewPlatform()

	var cmdLineArgs CmdLineArgs
	var err error

	if err = p.SetPaths(s); err != nil {
		panic(err)
	}

	if err = lockfile.LockFileCheck(p, s); err != nil {
		if err == common.Err_normal_exit {
			logfile.LogError("normal exit")
			return
		}
		panic(err)
	}
	defer lockfile.Close(p, s)

	if err = logfile.CreateLogfile(s); err != nil {
		panic(err)
	}

	logfile.LogInfo(version.CreateVersionStringFull())

	// log common paths
	logfile.LogInfo("main: PATH_CONF = '%s'", s.GetPathConf())
	logfile.LogInfo("main: PATH_USER = '%s'", s.GetPathUser())
	logfile.LogInfo("main: PATH_DATA = '%s'", s.GetPathData())

	// sdl inits
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_JOYSTICK); err != nil {
		logfile.LogError("main: Could not initialize SDL: %s", sdl.GetError())
		logfile.LogErrorDialog("main: Could not initialize SDL: %s", sdl.GetError())
		panic(err)
	}
	defer sdl.Quit()

	mods := modules.NewMods(p, s, cmdLineArgs.ModList)

	err = s.LoadSettings(mods)
	if err != nil {
		panic(err)
	}
	s.LogSettings()

	/*
		saveLoad := saveload.New()
		_ = saveLoad
	*/

	msg := modules.NewMsg()
	font := modules.NewFont(s, mods)
	defer font.Close()
	anim := modules.NewAnim()
	defer anim.Close()

	eset := modules.NewEset()
	err = eset.Load(s, mods, msg, font)
	if err != nil {
		panic(err)
	}

	inpt := modules.NewInpt(p, s, eset, mods, msg)
	defer inpt.Close()

	render := modules.NewRender(s, eset)
	defer render.Close()

	if err := render.CreateContext(s, eset, msg, mods); err != nil {
		panic(err)
	}

	render.ReloadGraphics()

	tooltipm := modules.NewTooltipm(s, mods, render)
	defer tooltipm.Close()

	icons := modules.NewIcons(s, eset, render, mods)
	defer icons.Close()

	gswitcher := state.NewSwitcher(modules) // 场景切换控制器
	defer gswitcher.Close(modules)

soft_reset:
	if !done {
		err = mainLoop(gswitcher)
		if err != nil {
			panic(err)
		}
	}

	if s.GetSoftReset() {
		logfile.LogInfo("main: Restarting Flare...")
		s.SetSoftReset(false)
		done = false
		goto soft_reset
	}

}

// 获得真实时间
func getSecondsElapsed(prevTicks, nowTicks uint64) float32 {
	// 计数 / 每秒计数 = 时间
	return (float32)(nowTicks-prevTicks) / (float32)(sdl.GetPerformanceFrequency())
}

func mainLoop(gswitcher common.GameSwitcher) error {
	render := modules.Render()
	settings := modules.Settings()
	inpt := modules.Inpt()

	maxFps := settings.Get("max_fps").(int)
	secondsPerFrame := float32(1) / (float32)(maxFps)
	prevTicks := sdl.GetPerformanceCounter()
	logicTicks := sdl.GetPerformanceCounter()
	lastFPS := (float32)(-1)
	var err error

	for !done {
		loops := 0
		nowTicks := sdl.GetPerformanceCounter()

		// 处理输入，处理帧逻辑
		for {
			if nowTicks < logicTicks || loops >= maxFps {
				break
			}

			// 正在加载资源，属于加载帧直接跳出
			if gswitcher.IsLoadingFrame() {
				// --1
				logicTicks = nowTicks
				break
			}

			// 处理输入
			sdl.PumpEvents()
			err = inpt.Handle(modules)
			if err != nil {
				return err
			}

			// 最小化的时候不处理
			if inpt.GetWindowMinimized() && !inpt.GetWindowRestored() && !inpt.GetDone() {
				break
			}

			// 场景逻辑
			err := gswitcher.Logic(modules) // 新游戏状态 2 ++1
			if err != nil {
				return err
			}
			inpt.ResetScroll()

			done = (gswitcher.GetDone() || inpt.GetDone())

			logicTicks += (uint64)(secondsPerFrame * (float32)(sdl.GetPerformanceFrequency()))
			loops++

			if inpt.GetWindowMinimized() && inpt.GetWindowRestored() {
				nowTicks = sdl.GetPerformanceCounter()
				logicTicks = nowTicks
				inpt.SetWindowMinimized(false)
				inpt.SetWindowRestored(false)
				break
			}
		}

		// 渲染
		if !inpt.GetWindowMinimized() {

			err = render.BlankScreen()
			if err != nil {
				return err
			}

			err = gswitcher.Render(modules)
			if err != nil {
				return err
			}

			if lastFPS != -1 {
				err = gswitcher.ShowFPS(modules, lastFPS)
				if err != nil {
					return err
				}
			}

			render.CommitFrame(inpt)

			var fpsDelay float32
			tmpe := getSecondsElapsed(prevTicks, sdl.GetPerformanceCounter())
			if tmpe < secondsPerFrame {
				fpsDelay = secondsPerFrame
			} else {
				fpsDelay = tmpe
			}

			if fpsDelay != 0 {
				lastFPS = (1000 / fpsDelay) / 1000
			} else {
				lastFPS = -1
			}
		}

		tmpe := getSecondsElapsed(prevTicks, sdl.GetPerformanceCounter())
		if tmpe < secondsPerFrame {
			delayMs := (secondsPerFrame-tmpe)*1000 - 1
			if delayMs > 0 {
				sdl.Delay(uint32(delayMs))
			}

			for {
				if getSecondsElapsed(prevTicks, sdl.GetPerformanceCounter()) >= secondsPerFrame {
					break
				}
			}
		}

		prevTicks = sdl.GetPerformanceCounter()
	}

	return nil
}

type CmdLineArgs struct {
	RenderDevideName string
	ModList          []string
}
