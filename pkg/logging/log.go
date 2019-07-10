package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"lbeng/pkg/file"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

// Setup initialize the log instance
func Setup() {
	var err error
	log.Print("[info] logging.Init")
	filePath := getLogFilePath()
	fileName := getLogFileName()
	log.Printf("[info] log file:%s/%s", filePath, fileName)
	F, err = file.MustOpen(fileName, filePath)
	if err != nil {
		log.Fatalf("logging.Init err: %v", err)
	}
	logger = log.New(F, DefaultPrefix, log.LstdFlags)
}

// Debug output logs at debug level
func Debug(v ...interface{}) {
	setPrefix(DEBUG)
	logger.Println(v)
}

//FmtInfo
func FmtInfo(str string, v ...interface{}) {
	setPrefix(INFO)
	logger.Println(fmt.Sprintf(str, v...))
}

// Info output logs at info level
func Info(v ...interface{}) {
	setPrefix(INFO)
	logger.Println(v)
}

// Warn output logs at warn level
func Warn(v ...interface{}) {
	setPrefix(WARNING)
	logger.Println(v)
}

// Error output logs at error level
func Error(v ...interface{}) {
	setPrefix(ERROR)
	logger.Println(v)
}

// Fatal output logs at fatal level
func Fatal(v ...interface{}) {
	setPrefix(FATAL)
	logger.Fatalln(v)
}

// setPrefix set the prefix of the log output
func setPrefix(level Level) {
	pc, file, line, ok := runtime.Caller(DefaultCallerDepth)
	if ok {
		fun := filepath.Ext(runtime.FuncForPC(pc).Name())
		logPrefix = fmt.Sprintf("[%s][%s%s:%d]", levelFlags[level], filepath.Base(file), fun, line)
	} else {
		logPrefix = fmt.Sprintf("[%s]", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}
