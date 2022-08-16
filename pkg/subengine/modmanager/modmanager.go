package modmanager

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"monster/pkg/common"
	"monster/pkg/common/define/modmanager"
	"monster/pkg/config/version"
	"monster/pkg/filesystem/logfile"
	"monster/pkg/utils"
	"monster/pkg/utils/parsing"
	"monster/pkg/utils/tools"
	"os"
)

// ==================== mod ===================
type Mod struct {
	name              string
	description       string
	game              string
	version           version.Version
	engineMinVersion  version.Version
	engineMaxVersion  version.Version
	descriptionLocale map[string]string
	depends           []string
	dependsMin        []version.Version
	dependsMax        []version.Version
}

func NewMod() *Mod {
	m := &Mod{}
	m.Init()

	return m
}

func (this *Mod) Init() {
	this.version = version.Version{0, 0, 0}
	this.engineMinVersion = version.MIN
	this.engineMaxVersion = version.MAX
	this.descriptionLocale = map[string]string{}
}

func (this *Mod) GetVersion() version.Version {
	return this.version
}

func (this *Mod) GetEngineMinVersion() version.Version {
	return this.engineMinVersion
}

func (this *Mod) GetEngineMaxVersion() version.Version {
	return this.engineMaxVersion
}

func (this *Mod) GetGame() string {
	return this.game
}

func (this *Mod) GetName() string {
	return this.name
}

func (this *Mod) GetDepends() []string {
	lenDepends := len(this.depends)

	tmpDepends := make([]string, lenDepends)
	for i, depend := range this.depends {
		tmpDepends[i] = depend

	}
	return tmpDepends
}

func (this *Mod) GetDependsMin() []version.Version {
	lenDepends := len(this.dependsMin)

	tmpDepends := make([]version.Version, lenDepends)
	for i, depend := range this.dependsMin {
		tmpDepends[i] = depend

	}
	return tmpDepends
}

func (this *Mod) GetDependsMax() []version.Version {
	lenDepends := len(this.dependsMax)

	tmpDepends := make([]version.Version, lenDepends)
	for i, depend := range this.dependsMax {
		tmpDepends[i] = depend

	}
	return tmpDepends
}

func (this *Mod) GetLocalDescription(key string) string {
	val, ok := this.descriptionLocale[key]
	if ok {
		return val
	} else {
		return this.description
	}
}

// ==================== mod manager ===================
type ModManager struct {
	locCache    map[string]string
	modPaths    []string     // 查找的根目录 modpaths: [CustomPathData, PathUser, PathData]
	modDirs     []string     // dirs in modPath/mods/  mods文件里的全部目录
	modList     []common.Mod // 已经加载的mod
	cmdLineMods []string
}

func New(platform common.Platform, settings common.Settings, cmdLineMods []string) *ModManager {
	mm := &ModManager{
		locCache: map[string]string{},
		modPaths: []string{},
		modDirs:  []string{},
		modList:  []common.Mod{},
	}

	mm.init(platform, settings, cmdLineMods)

	return mm
}

func (this *ModManager) init(platform common.Platform, settings common.Settings, cmdLineMods []string) common.ModManager {

	var err error

	this.cmdLineMods = cmdLineMods

	this.setPaths(platform, settings)

	modDirsOther := []string{}
	modDirsOther, err = utils.GetDirList(settings.GetPathData()+"mods", modDirsOther)
	if err != nil {
		panic(err)
	}
	modDirsOther, err = utils.GetDirList(settings.GetPathUser()+"mods", modDirsOther)
	if err != nil {
		panic(err)
	}

	for _, val := range modDirsOther {
		if !tools.FindStr(this.modDirs, val) {
			this.modDirs = append(this.modDirs, val)
		}
	}

	if err = this.loadModList(settings); err != nil {
		panic(err)
	}

	if err = this.ApplyDepends(); err != nil {
		panic(err)
	}

	activeModsStr := "Active mods: "
	for index, mod := range this.modList {
		activeModsStr += mod.GetName()
		if version.Compare(version.MIN, mod.GetVersion()) != 0 {
			activeModsStr += " (" + mod.GetVersion().GetString() + ")"
		}

		if index < len(this.modList)-1 {
			activeModsStr += ", "
		}
	}

	logfile.LogInfo(activeModsStr)

	return this
}

// 列出已加载mod下的全部指定文件 (mod里结构为目录和txt文件)
func (this *ModManager) List(path string) ([]string, error) {
	var ret []string

	// [default, ...]
	for _, mod := range this.modList {

		// modpaths: [CustomPathData, PathUser, PathData]
		// 反向遍历 从  PathData 开始找到对应的文件
		for j := len(this.modPaths); j > 0; j-- {
			testPath := this.modPaths[j-1] + "mods/" + mod.GetName() + "/" + path

			if isExist, err := utils.PathExists(testPath); err != nil {
				return nil, err
			} else {
				if isExist {
					if isDir, err := utils.IsDirectory(testPath); err != nil {
						return nil, err
					} else if isDir {
						ret, err = utils.GetFileList(testPath, "txt", ret)
						if err != nil {
							return nil, err
						}
					} else {
						ret = append(ret, testPath)
					}
				}
			}
		}
	}

	//fmt.Println(path, ret)
	return ret, nil
}

// modpaths: [CustomPathData, PathUser, PathData]
func (this *ModManager) setPaths(platform common.Platform, settings common.Settings) {
	s := settings

	var uniqPathData = (s.GetPathUser() != s.GetPathData())

	if s.GetCustomPathData() != "" {
		this.modPaths = append(this.modPaths, s.GetCustomPathData())
		uniqPathData = false
	}

	this.modPaths = append(this.modPaths, s.GetPathUser())
	if uniqPathData {
		this.modPaths = append(this.modPaths, s.GetPathData())
	}

}

// 加载单个mod
func (this *ModManager) LoadMod(name string) (common.Mod, error) {
	mod := NewMod()
	mod.name = name

	found := false

	// default mod 没有 settings.txt，作为被覆盖公共基础资源
	if name == modmanager.FALLBACK_MOD {
		found = true
		return mod, nil
	}

	for _, modPath := range this.modPaths {

		path := modPath + "mods/" + name + "/settings.txt"
		f, err := os.Open(path)
		if err != nil && utils.IsNotExist(err) {
			continue
		} else if err != nil {
			return mod, err
		}

		found = true
		// just one path
		defer f.Close()
		scanner := bufio.NewScanner(f)
		bytesComment := []byte("#")
		for scanner.Scan() {
			if len(scanner.Bytes()) == 0 || bytes.HasPrefix(bytes.TrimSpace(scanner.Bytes()), bytesComment) {
				continue
			}

			key, val := parsing.GetKeyPair(scanner.Bytes())
			switch key {
			case "description":
				mod.description = val
			case "description_locale":
				localeStr, str := parsing.PopFirstString(val, "")
				if localeStr != "" {
					mod.descriptionLocale[localeStr], _ = parsing.PopFirstString(str, "")
				}
			case "version":
				mod.version.SetFromString(val)
			case "requires":
				//(e.g. fantasycore:0.1:2.0)
				str := val + ","
				dep := ""
				for {
					if dep, str = parsing.PopFirstString(str, ""); dep == "" {
						break
					}

					depFull := dep + "::"
					dep, depFull = parsing.PopFirstString(depFull, ":")
					mod.depends = append(mod.depends, dep)
					depMin := version.Version{0, 0, 0}
					depMax := version.Version{0, 0, 0}
					dep, depFull = parsing.PopFirstString(depFull, ":")
					depMin.SetFromString(dep)
					dep, depFull = parsing.PopFirstString(depFull, ":")
					depMax.SetFromString(dep)

					if version.Compare(depMin, version.MIN) != 0 && version.Compare(depMax, version.MIN) != 0 && version.Compare(depMin, depMax) > 0 {

						depMax = depMin
					}

					mod.dependsMin = append(mod.dependsMin, depMin)

					if version.Compare(depMax, version.MIN) == 0 {
						mod.dependsMax = append(mod.dependsMax, version.MAX)
					} else {
						mod.dependsMax = append(mod.dependsMax, depMax)
					}
				}

			case "game":
				mod.game = val
			case "engine_version_min":
				mod.engineMinVersion.SetFromString(val)
			case "engine_version_max":
				mod.engineMaxVersion.SetFromString(val)
			}
		}

		if err := scanner.Err(); err != nil {
			return mod, err
		}
		break
	}

	if !found {
		return mod, common.Err_modmanager_load_mod_failed
	}

	if version.Compare(mod.engineMinVersion, version.MIN) != 0 && version.Compare(mod.engineMinVersion, mod.engineMaxVersion) > 1 {
		mod.engineMaxVersion = mod.engineMinVersion
	}

	return mod, nil
}

/**
 * The mod list is in either:
 * 1. [PATH_CONF]/mods.txt
 * 2. [PATH_DATA]/mods/mods.txt
 * The mods.txt file shows priority/load order for mods
 *
 * File format:
 * One mod folder name per line
 * Later mods override previous mods
 */

// 从配置文件mods.txt加载先前配置的mod
func (this *ModManager) loadModList(settings common.Settings) error {
	s := settings

	// Add the fallback mod by default
	// Note: if a default mod is not found in mod_dirs, the game will exit
	foundAnyMod := tools.FindStr(this.modDirs, modmanager.FALLBACK_MOD)
	if !foundAnyMod {
		return common.Err_no_fallbackmod
	}

	mod, err := this.LoadMod(modmanager.FALLBACK_MOD)
	if err != nil {
		return err
	}
	this.modList = append(this.modList, mod)

	if len(this.cmdLineMods) == 0 {
		place1 := s.GetPathConf() + "mods.txt"
		place2 := s.GetPathData() + "mods/mods.txt"

		f, err := os.Open(place1)
		if err != nil && utils.IsNotExist(err) {
			if f, err = os.Open(place2); err != nil && utils.IsNotExist(err) {
				logfile.LogError("ModManager: Error during loadModList() -- couldn't open mods.txt, to be located at:%s\n%s\n", place1, place2)
				return err
			} else if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		defer f.Close()
		scanner := bufio.NewScanner(f)
		bytesComment := []byte("#")
		for scanner.Scan() {
			if len(scanner.Bytes()) == 0 || bytes.HasPrefix(bytes.TrimSpace(scanner.Bytes()), bytesComment) {
				continue
			}

			line := scanner.Text()
			if line != modmanager.FALLBACK_MOD {
				if tools.FindStr(this.modDirs, line) {
					mod, err := this.LoadMod(line)
					if err != nil {
						return err
					}

					this.modList = append(this.modList, mod)
					foundAnyMod = true
				} else {
					logfile.LogError("ModManager: Mod \"%s\" not found, skipping", line)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		for _, line := range this.cmdLineMods {
			if line != modmanager.FALLBACK_MOD {
				if tools.FindStr(this.modDirs, line) {
					mod, err := this.LoadMod(line)
					if err != nil {
						return err
					}

					this.modList = append(this.modList, mod)
					foundAnyMod = true
				} else {
					logfile.LogError("ModManager: Mod \"%s\" not found, skipping", line)
				}
			}
		}
	}

	return nil
}

// 返回mod范围内的某个文件，不存在则返回不存在错误
func (this *ModManager) Locate(settings common.Settings, filename string) (string, error) {
	if loc, ok := this.locCache[filename]; ok {
		return loc, nil
	}

	for i := len(this.modList); i > 0; i-- {
		for _, path := range this.modPaths {
			testPath := path + "mods/" + this.modList[i-1].GetName() + "/" + filename
			if ok, err := utils.FileExists(testPath); err != nil {
				return "", err
			} else if ok {
				this.locCache[filename] = testPath
				return testPath, nil
			}
		}
	}

	testPath := settings.GetPathData() + filename
	if ok, err := utils.FileExists(testPath); err != nil {
		return "", err
	} else if ok {
		return testPath, nil
	}

	return "", fs.ErrNotExist
}

// 加载mod的依赖mod
func (this *ModManager) ApplyDepends() error {
	var newMods []common.Mod
	newModsMap := map[string]struct{}{}
	var game string
	finished := true

	if len(this.modList) != 0 {
		game = this.modList[len(this.modList)-1].GetGame() // the last push back one
	}

	for _, mod := range this.modList {
		// skip the mod if the game doesn't match
		if game != modmanager.FALLBACK_GAME && mod.GetGame() != modmanager.FALLBACK_GAME && mod.GetGame() != game && mod.GetName() != modmanager.FALLBACK_MOD {
			logfile.LogError("ModManager: Tried to enable \"%s\", but failed. Game does not match \"%s\".", mod.GetName(), game)
			continue
		}

		// skip the mod if it's incompatible with this engine version
		if version.Compare(mod.GetEngineMinVersion(), version.ENGINE) > 0 || version.Compare(version.ENGINE, mod.GetEngineMaxVersion()) > 0 {
			logfile.LogError("ModManager: Tried to enable \"%s\", but failed. Not compatible with engine version %s.", mod.GetName(), version.ENGINE.GetString())
			continue
		}

		// skip the mod if it's already in the new_mods list
		if _, ok := newModsMap[mod.GetName()]; ok {
			continue
		}

		dependsMet := true
		for index, dep := range mod.GetDepends() {
			foundDepend := false

			// try to add the dependecy to the new_mods list
			if _, ok := newModsMap[dep]; ok {
				foundDepend = true
			}

			if !foundDepend {
				// if we don't already have this dependency, try to load it from the list of available mods
				if tools.FindStr(this.modDirs, dep) {
					newDepend, err := this.LoadMod(dep)
					if err != nil {
						return err
					}

					if game != modmanager.FALLBACK_GAME && newDepend.GetGame() != modmanager.FALLBACK_GAME && newDepend.GetGame() != game {
						logfile.LogError("ModManager: Tried to enable dependency \"%s\" for \"%s\", but failed. Game does not match \"%s\".", newDepend.GetName(), mod.GetName(), game)
						dependsMet = false
					} else if version.Compare(newDepend.GetEngineMinVersion(), version.ENGINE) > 0 || version.Compare(version.ENGINE, newDepend.GetEngineMaxVersion()) > 0 {
						logfile.LogError("ModManager: Tried to enable dependency \"%s\" for \"%s\", but failed. Not compatible with engine version %s.", newDepend.GetName(), mod.GetName(), version.ENGINE.GetString())

						dependsMet = false
					} else if version.Compare(newDepend.GetVersion(), mod.GetDependsMin()[index]) < 0 || version.Compare(newDepend.GetVersion(), mod.GetDependsMax()[index]) > 0 {
						logfile.LogError("ModManager: Tried to enable dependency \"%s\" for \"%s\", but failed. Version \"%s\" is required, but only version \"%s\" is available.", newDepend.GetName(), mod.GetName(), version.CreateVersionReqString(mod.GetDependsMin()[index], mod.GetDependsMax()[index]), newDepend.GetVersion().GetString())
						dependsMet = false
					} else if _, ok := newModsMap[newDepend.GetName()]; !ok {
						logfile.LogError("ModManager: Mod \"%s\" requires the \"%s\" mod. Enabling \"%s\" now.", mod.GetName(), dep, dep)
						newMods = append(newMods, newDepend)
						newModsMap[newDepend.GetName()] = struct{}{}
						dependsMet = true
						finished = false
					}

				} else {
					logfile.LogError("ModManager: Could not find mod \"%s\", which is required by mod \"%s\". Disabling \"%s\" now.", dep, mod.GetName(), mod.GetName())
					dependsMet = false
				}
			}

			if !dependsMet {
				break
			}
		}

		// if the mod's depends is all ok then add the mod self
		if dependsMet {
			if _, ok := newModsMap[mod.GetName()]; !ok {
				newMods = append(newMods, mod)
				newModsMap[mod.GetName()] = struct{}{}
			}
		}
	}

	this.modList = newMods
	if !finished {
		if err := this.ApplyDepends(); err != nil {
			return err
		}
	}

	return nil
}

// 获得已经加载的mod，copy出来
func (this *ModManager) GetModList() []common.Mod {
	lenModList := len(this.modList)

	newModList := make([]common.Mod, lenModList)

	for i, mod := range this.modList {
		// copy one
		newModList[i] = mod
	}

	return newModList
}

//  获取/mods文件里的全部目录
func (this *ModManager) GetModDirs() []string {
	lenModDirs := len(this.modDirs)

	newModDirs := make([]string, lenModDirs)

	for i, dir := range this.modDirs {
		// copy one
		newModDirs[i] = dir
	}

	return newModDirs
}

func (this *ModManager) ClearModList() {
	this.modList = nil
}

func (this *ModManager) AddToModList(newMod common.Mod) {
	for _, mod := range this.modList {
		if mod.GetName() == newMod.GetName() {
			return
		}
	}

	this.modList = append(this.modList, newMod)
}

func (this *ModManager) SaveMods(modules common.Modules) error {
	settings := modules.Settings()

	f, err := os.OpenFile(settings.GetPathConf()+"mods.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "## flare-engine mods list file ##")
	fmt.Fprintln(f, "# Mods lower on the list will overwrite data in the entries higher on the list")
	fmt.Fprintln(f)

	for _, mod := range this.modList {
		if mod.GetName() != modmanager.FALLBACK_MOD {
			fmt.Fprintf(f, "%s\n", mod.GetName())
		}
	}

	return nil
}
