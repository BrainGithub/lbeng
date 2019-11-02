package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"lbeng/pkg/file"
	"lbeng/pkg/setting"

	"github.com/sirupsen/logrus"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	loggerus   *logrus.Logger
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

var logType = 2 //0: standard, 1: logrus

// Setup initialize the log instance
func Setup() {
	log.Print("[info] logging.Init")
	filePath := getLogFilePath()
	fileName := getLogFileName()
	log.Printf("[info] log file:%s/%s", filePath, fileName)
	f, err := file.MustOpen(fileName, filePath)
	if err != nil {
		log.Fatalf("logging.Init err: %v", err)
	}

	{
		logger = log.New(f, DefaultPrefix, log.LstdFlags|log.Lmicroseconds)
	}

	{
		fName := fileName + ".rus"
		log.Printf("[info] log file:%s/%s", filePath, fName)
		// file, _ := file.MustOpen(fName, filePath)
		loggerus = logrus.New()
		loggerus.SetOutput(os.Stdout)
		loggerus.SetLevel(logrus.DebugLevel)
		loggerus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func GetLogrus() *logrus.Logger {
	return loggerus
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

// FmtError output logs at error level
func FmtError(str string, v ...interface{}) {
	setPrefix(ERROR)
	logger.Println(fmt.Sprintf(str, v...))
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

// getLogFilePath get the log file save path
func getLogFilePath() string {
	return fmt.Sprintf("%s", setting.AppSetting.LogSavePath)
}

// getLogFileName get the save name of the log file
func getLogFileName() string {
	return fmt.Sprintf("%s.%s",
		setting.AppSetting.LogSaveName,
		setting.AppSetting.LogFileExt,
	)
}
