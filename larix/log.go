package larix

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Log struct {
	//Log file name
	File string

	//error file , default
	errorFile string

	// if log rotate, current support for hour
	Rotate bool

	// basic fields, will out in every log
	BasicFields map[string]interface{}

	//log level
	Level int

	//current rotate sign, we use year montn day hour
	CurrRotateSign int

	// ensure atomic log handler update
	mu sync.Mutex

	// normal log file handler
	nfd *os.File

	// wrong log file handler
	wfd *os.File

	// normal Log Handler
	nlog *log.Logger

	// wrong log Handler
	wlog *log.Logger
}

//control fields
// file : string , if nil, we'll set to stdout
// level : int  , if nil, we'll set to debug
// rotate : bool  , if nil, we'll set to true
// like:
/*
	conf = map[string]interface{}{
		"file" : "./logs/test",
		"rotate": true,
		"level": 0,
	}
*/
// conf define for ini conf to map
// now not use
type LogConf struct {
	File   string `ini:"file"`
	Rotate bool   `ini:"rotate"`
	Level  int    `ini:"level"`
}

//log level
const (
	LOG_DEBUG int = iota
	LOG_TRACE
	LOG_NOTICE
	LOG_WARN
	LOG_FATAL
	LOG_OVER
)

var logString map[int]string = map[int]string{
	LOG_DEBUG:  "debug",
	LOG_TRACE:  "trace",
	LOG_NOTICE: "notice",
	LOG_WARN:   "warning",
	LOG_FATAL:  "fatal",
}

//log level string

//Log handler
var logHdr *Log = nil

//interface
func LogInit(conf map[string]interface{}) error {
	if logHdr != nil {
		LogFatal("re init log handler, invalid")
		return nil
	}

	logHdr = &Log{}

	//why use exists ,not use err; for get memeber succ ,exists set to true, not false,
	tmpFile, exists := conf["log_path"]
	if exists {
		file_path, exists := tmpFile.(string)
		if !exists {
			panic("log file path type error, we need string\n")
		}
		//we'll not check file path here
		logHdr.File = file_path
	} else {
		//now we'll not support to write to stdout or stderr
		panic("log file path not found, please set by file field\n")
	}

	tmpRotate, exists := conf["log_rotate"]
	if exists {
		logHdr.Rotate, exists = tmpRotate.(bool)
		if !exists {
			io.WriteString(os.Stdout, "log rotate get failed, we need bool type, and now set to false\n")
			logHdr.Rotate = false
		}
	} else {
		logHdr.Rotate = false
	}

	tmpLevel, exists := conf["log_level"]
	if exists {
		logHdr.Level, exists = tmpLevel.(int)
		if !exists {
			io.WriteString(os.Stdout, "log level get failed, we need int type, and now set to debug\n")
			logHdr.Level = LOG_DEBUG
		}
		if logHdr.Level >= LOG_OVER {
			io.WriteString(os.Stdout, "log level invalid, now set to debug\n")
			logHdr.Level = LOG_DEBUG
		}
	} else {
		logHdr.Level = LOG_DEBUG
	}

	//todo: basic fields support
	var nlfile string
	var wlfile string

	if logHdr.Rotate {
		logHdr.CurrRotateSign = getRotateSign()
		nlfile = fmt.Sprintf("%s.%d", logHdr.File, logHdr.CurrRotateSign)
		wlfile = fmt.Sprintf("%s.wf.%d", logHdr.File, logHdr.CurrRotateSign)
	} else {
		logHdr.CurrRotateSign = 0
		nlfile = logHdr.File
		wlfile = fmt.Sprintf("%s.wf", logHdr.File)
	}

	//open file and add log handler
	var errf error
	logHdr.nfd, errf = os.OpenFile(nlfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeSetuid|os.ModeSetgid|0660)
	if errf != nil {
		panic(fmt.Sprintf("open log file %s failed, %s\n", nlfile, errf.Error()))
	}

	logHdr.wfd, errf = os.OpenFile(wlfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeSetuid|os.ModeSetgid|0660)
	if errf != nil {
		logHdr.nfd.Close()
		panic(fmt.Sprintf("open wflog file %s failed, %s\n", wlfile, errf.Error()))
	}

	// normal Log Handler
	logHdr.nlog = log.New(logHdr.nfd, "", log.Ldate|log.Ltime|log.Lshortfile)

	// wrong log Handler
	logHdr.wlog = log.New(logHdr.wfd, "", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func getRotateSign() int {
	t_case := time.Now()
	return t_case.Year()*1000000 + int(t_case.Month())*10000 + t_case.Day()*100 + t_case.Hour()
}

func (l *Log) rotate() error {
	if !l.Rotate {
		return nil
	}

	curr_rotate_sign := getRotateSign()
	if l.CurrRotateSign == curr_rotate_sign {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.nfd != nil {
		l.nfd.Close()
	}
	if l.wfd != nil {
		l.wfd.Close()
	}

	nlfile := fmt.Sprintf("%s.%d", l.File, curr_rotate_sign)
	wlfile := fmt.Sprintf("%s.wf.%d", l.File, curr_rotate_sign)

	//open file and add log handler
	var err error
	l.nfd, err = os.OpenFile(nlfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeSetuid|os.ModeSetgid|0660)
	if err != nil {
		//todo: to write error msg to stderr
		return err
	}

	l.wfd, err = os.OpenFile(wlfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeSetuid|os.ModeSetgid|0660)
	if err != nil {
		//todo: to write error msg to stderr
		return err
	}

	// normal Log Handler
	l.nlog = log.New(l.nfd, "", log.Ldate|log.Ltime|log.Lshortfile)

	// wrong log Handler
	l.wlog = log.New(l.wfd, "", log.Ldate|log.Ltime|log.Lshortfile)

	//when rotate done, we change the sign ,for we can reentrant
	l.CurrRotateSign = curr_rotate_sign
	return nil
}

func (l *Log) WriteLog(level int, v ...interface{}) {
	//step1: rotate
	if l.rotate() != nil {
		return
	}

	//step2: gen log string
	log_str := l.genLogString(level, v...)

	//step3: write log
	if level < LOG_WARN {
		l.nlog.Print(log_str)
	} else {
		l.wlog.Print(log_str)
	}

	return
}

//todo
func (l *Log) genLogString(level int, v ...interface{}) string {
	if len(v) >= 2 {
		format_str, err := v[0].(string)

		//if v[0] is string
		if err {
			return strings.TrimSpace(fmt.Sprintf("["+logString[level]+"] "+format_str, v[1:]...)) + "\n"
		}
	}

	var res bytes.Buffer
	res.WriteString("[" + logString[level] + "]")
	//compate with multi params
	for _, val := range v {
		res.WriteString(" ")
		//for standard logger just check last character if is "\n",
		//can't deal multi space characters,
		//so we need to deal space characters by ourself
		res.WriteString(strings.TrimSpace(l.dealFields(val)))
	}
	res.WriteString("\n")

	return res.String()
}

func (l *Log) dealFields(v interface{}) string {
	//now we just deal string ,map and
	value := reflect.ValueOf(v)
	var vKind = value.Kind()

	switch {
	case vKind == 0:
		return "nil"

	//bool + int +uint
	case vKind < 12:
		return fmt.Sprintf("%v", value)

	//uintptr
	case vKind == 12:
		return "[uptr]"

	//float + complex
	case vKind < 17:
		return fmt.Sprintf("%v", value)

	//Array
	case vKind == 17:
		return fmt.Sprintf("array[%v]", value)

	//chan\func\Interface
	case vKind < 21:
		return fmt.Sprintf("%v", value)

	//Map
	case vKind == 21:
		var res bytes.Buffer
		keys := value.MapKeys()

		for _, key := range keys {
			res.WriteString(fmt.Sprintf("%v[%v] ", key, value.MapIndex(key)))
		}
		return res.String()

	//ptr
	case vKind == 22:
		return "[ptr]"

	//slice
	case vKind == 23:
		return fmt.Sprintf("slice[%v]", value)

	//string
	case vKind == 24:
		return fmt.Sprintf("%v", value)

	//Struct
	case vKind == 25:
		return fmt.Sprintf("struct[%s]", value.Type())

	//Unsafeptr
	case vKind == 26:
		return "[Unsafeptr]"

	default:
		return "Unknown"
	}

	return "Unknown"
}

func LogDestory() {
	if logHdr == nil {
		return
	}

	if logHdr.nfd != nil {
		logHdr.nfd.Close()
	}
	if logHdr.wfd != nil {
		logHdr.wfd.Close()
	}
	logHdr = nil
}

//todo
func LogAddBasic() {
}

//todo
func LogRmBasic() {
}

func LogDebug(v ...interface{}) {
	if logHdr == nil {
		return
	}

	if LOG_DEBUG < logHdr.Level {
		return
	}

	logHdr.WriteLog(LOG_DEBUG, v...)
}

func LogTrace(v ...interface{}) {
	if logHdr == nil {
		return
	}

	if LOG_TRACE < logHdr.Level {
		return
	}

	logHdr.WriteLog(LOG_TRACE, v...)
}

func LogNotice(v ...interface{}) {
	if logHdr == nil {
		return
	}

	if LOG_NOTICE < logHdr.Level {
		return
	}

	logHdr.WriteLog(LOG_NOTICE, v...)
}

func LogWarn(v ...interface{}) {
	if logHdr == nil {
		return
	}

	if LOG_WARN < logHdr.Level {
		return
	}

	logHdr.WriteLog(LOG_WARN, v...)
}

func LogFatal(v ...interface{}) {
	if logHdr == nil {
		return
	}

	if LOG_FATAL < logHdr.Level {
		return
	}

	logHdr.WriteLog(LOG_FATAL, v...)
}
