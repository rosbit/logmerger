package logmerger

import (
	"log"
	"fmt"
	"os"
	"time"
	"io"
)

const (
	_DEFAULT_DURATION = 1000  // in milliseconds
	_TIME_FORMAT      = "20060102_150405"
	_REUSE_SUFFIX     = "_reuse"
)

type LogMerger struct {
	sleepDuration time.Duration
	exit          bool
}

type FnFileHandler func(fileName string)

func NewLogMerger(sleepDurationInMs int) *LogMerger {
	if sleepDurationInMs <= 0 {
		sleepDurationInMs = _DEFAULT_DURATION
	}
	return &LogMerger{sleepDuration:time.Duration(sleepDurationInMs)*time.Millisecond, exit:false}
}

/**
 * The main process: monitor logFile, handle it with fileHandler, and merge it to a daily file.
 * The process will loop forever until Stop() called.
 */
func (lm *LogMerger) Run(logFile string, fileHandler FnFileHandler) {
	lm.run(logFile, fileHandler, false)
}

/**
 * The main process: monitor logFile, handle it with fileHandler, and merge it to a file with suffix "_reuse".
 * The process will loop forever until Stop() called.
 */
func (lm *LogMerger) RunWithoutMerging(logFile string, fileHandler FnFileHandler) {
	lm.run(logFile, fileHandler, true)
}

func (lm *LogMerger) Stop() {
	lm.exit = true
	log.Printf("[logmerger] stop signal received\n")
}

func (lm *LogMerger) Test(logFile string, fileHandler FnFileHandler) {
	log.Printf("[logmerger] Processing %s for testing\n", logFile)
	fileHandler(logFile)
}

func (lm *LogMerger) run(logFile string, fileHandler FnFileHandler, dontMerge bool) {
	reuseLogFile := fmt.Sprintf("%s%s", logFile, _REUSE_SUFFIX)
	var lf string
	for !lm.exit {
		if _, err := os.Stat(reuseLogFile); err == nil {
			lf = reuseLogFile
		} else if _, err := os.Stat(logFile); err == nil {
			lf = logFile
		} else {
			time.Sleep(lm.sleepDuration)
			continue
		}
		lm.processLogFile(lf, fileHandler, dontMerge)
		time.Sleep(lm.sleepDuration)
	}
	log.Printf("[logmerger] I will exit\n")
}

func (lm *LogMerger) processLogFile(logFile string, fileHandler FnFileHandler, dontMerge bool) {
	t := time.Now()
	now := t.Format(_TIME_FORMAT)
	inFile, logFile := lm.renameLogFile(dontMerge, logFile, now)
	fileHandler(inFile)

	var mergedLogFile string
	if dontMerge {
		// merged to a tmp file to be re-used later
		mergedLogFile = fmt.Sprintf("%s%s", logFile, _REUSE_SUFFIX)
	} else {
		mergedLogFile = fmt.Sprintf("%s_%s", logFile, now[:len(now)-7])
	}
	lm.mergeLog(inFile, mergedLogFile)
}

func (lm *LogMerger) renameLogFile(dontMerge bool, logFile string, now string) (string, string) {
	lenSuffix := len(_REUSE_SUFFIX)
	if !dontMerge {
		tmpSuffixPos := len(logFile)-lenSuffix
		if logFile[tmpSuffixPos:] == _REUSE_SUFFIX {
			// a tmp file encountered
			oldLogFile := logFile[0:tmpSuffixPos]
			inFile := fmt.Sprintf("%s_%s", oldLogFile, now)
			os.Rename(logFile, inFile)
			return inFile, oldLogFile
		}
	}

	inFile := fmt.Sprintf("%s_%s", logFile, now)
	os.Rename(logFile, inFile)
	return inFile, logFile
}

func (lm *LogMerger) mergeLog(inFile string, mergedLogFile string) {
	if _, err := os.Stat(mergedLogFile); err == nil {
		fpDest, _ := os.OpenFile(mergedLogFile, os.O_WRONLY|os.O_APPEND, 0644)
		fpSrc, _  := os.Open(inFile)
		io.Copy(fpDest, fpSrc)
		fpDest.Close()
		fpSrc.Close()
		os.Remove(inFile)
	} else {
		os.Rename(inFile, mergedLogFile)
	}
}
