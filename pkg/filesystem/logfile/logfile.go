package logfile

import (
	"container/list"
	"fmt"
	"monster/pkg/common"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	defaultLogFile = new()
)

type LogFile struct {
	logFileInit bool
	logPath     string
	logMsg      *list.List
}

type Pair struct {
	First  sdl.LogPriority
	Second string
}

func new() *LogFile {
	return &LogFile{
		logFileInit: false,
		logMsg:      list.New(),
	}
}

func (this *LogFile) createLogFile(settings common.Settings) error {
	s := settings
	var err error

	this.logPath = s.GetPathConf() + "/flare_log.txt"
	f, err := os.OpenFile(this.logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		// just clear
		this.logMsg = this.logMsg.Init()
		return err
	}
	defer f.Close()

	// log to file
	fmt.Fprintf(f, "### Flare log file\n\n")
	if this.logMsg.Len() != 0 {

		// write before log
		for e := this.logMsg.Front(); e != nil; e = e.Next() {
			val := e.Value.(*Pair)
			if val.First == sdl.LOG_PRIORITY_INFO {
				fmt.Fprintf(f, "INFO: %s\n", val.Second)
			} else if val.First == sdl.LOG_PRIORITY_ERROR {
				fmt.Fprintf(f, "ERROR: %s\n", val.Second)
			}
		}
		this.logMsg = this.logMsg.Init()
	}
	this.logFileInit = true
	return nil
}

func (this *LogFile) logInfo(format string, args ...interface{}) error {
	tmpArgs := (args[0]).([]interface{})
	sdl.LogInfo(sdl.LOG_CATEGORY_APPLICATION, format, tmpArgs...)

	if !this.logFileInit {
		this.logMsg.PushBack(&Pair{sdl.LOG_PRIORITY_INFO, fmt.Sprintf(format, tmpArgs...)})
		return nil
	}

	f, err := os.OpenFile(this.logPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "INFO: "+format+"\n", tmpArgs...)

	return nil
}

func (this *LogFile) logError(format string, args ...interface{}) error {
	tmpArgs := (args[0]).([]interface{})

	// #FIXME priority has no effect
	// sdl.LogMessage(sdl.LOG_CATEGORY_APPLICATION, sdl.LOG_PRIORITY_ERROR, format, tmpArgs)
	sdl.LogError(sdl.LOG_CATEGORY_APPLICATION, format, tmpArgs...)

	if !this.logFileInit {
		this.logMsg.PushBack(&Pair{sdl.LOG_PRIORITY_ERROR, fmt.Sprintf(format, tmpArgs...)})
		return nil
	}

	f, err := os.OpenFile(this.logPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "ERROR: "+format+"\n", tmpArgs...)

	return nil
}

func (this *LogFile) logErrorDialog(format string, args ...interface{}) {
	tmpArgs := (args[0]).([]interface{})
	sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "FLARE Error", fmt.Sprintf(fmt.Sprintf("%s%s", "FLARE ERROR\n", format), tmpArgs...), nil)
}

// exported package method
func CreateLogfile(settings common.Settings) error {
	return defaultLogFile.createLogFile(settings)
}

func LogInfo(format string, args ...interface{}) error {
	return defaultLogFile.logInfo(format, args)
}

func LogError(format string, args ...interface{}) error {
	return defaultLogFile.logError(format, args)
}

func LogErrorDialog(format string, args ...interface{}) {
	defaultLogFile.logErrorDialog(format, args)
}
