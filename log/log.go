package log

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// levels
const (
	debugLevel   = 0
	releaseLevel = 1
	warningLevel = 2
	infoLevel    = 3
	errorLevel   = 4
	fatalLevel   = 5
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printWarningLevel = "[warning] "
	printInfoLevel    = "[info   ] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	baseFile   *os.File
	baseName   string
}

func New(strLevel string, pathname string) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "warning":
		level = warningLevel
	case "info":
		level = infoLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File
	if pathname != "" {
		now := time.Now()

		//		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
		//			now.Year(),
		//			now.Month(),
		//			now.Day(),
		//			now.Hour(),
		//			now.Minute(),
		//			now.Second())
		filename := fmt.Sprintf("%d%02d%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour())

		//file, err := os.OpenFile(path.Join(pathname, filename), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}

		baseLogger = log.New(file, "", log.LstdFlags)
		baseFile = file
	} else {
		baseLogger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// new
	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger
	logger.baseFile = baseFile

	return logger, nil
}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}

	logger.baseLogger = nil
	logger.baseFile = nil
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	logger.baseLogger.Printf(format, a...)

	if level == fatalLevel {
		os.Exit(1)
	}
}

func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func (logger *Logger) Warning(format string, a ...interface{}) {
	logger.doPrintf(warningLevel, printWarningLevel, format, a...)
}

func (logger *Logger) Info(format string, a ...interface{}) {
	logger.doPrintf(infoLevel, printInfoLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

func (logger *Logger) SetLoggerLevel(strLevel string) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "warning":
		level = warningLevel
	case "info":
		level = infoLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		level = logger.level
	}
	logger.level = level
}

var gLogger, _ = New("debug", "")

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

func Debug(format string, a ...interface{}) {
	gLogger.Debug(format, a...)
}

func Release(format string, a ...interface{}) {
	gLogger.Release(format, a...)
}

func Warning(format string, a ...interface{}) {
	gLogger.Warning(format, a...)
}

func Info(format string, a ...interface{}) {
	gLogger.Info(format, a...)
}

func Error(format string, a ...interface{}) {
	gLogger.Error(format, a...)
}

func Fatal(format string, a ...interface{}) {
	gLogger.Fatal(format, a...)
}

func Close() {
	gLogger.Close()
}
