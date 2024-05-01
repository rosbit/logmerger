package logmerger

import (
	"strings"
	p "path"
	"io/fs"
	"log"
	"fmt"
	"os"
	"time"
)

type chanHandler struct {
	logFile string
	fileHandler FnFileHandler
}

/**
 * The main process: monitor rootDir, handle files with ext with fileHandler, and merge them to daily files.
 * The process will loop forever until Stop() called.
 */
func (lm *LogMerger) RunDir(rootDir string, ext string, fileHandler FnFileHandler, maxHandlerCount ...int) {
	lm.runDir(rootDir, ext, fileHandler, maxHandlerCount...)
}

func (lm *LogMerger) runDir(rootDir string, ext string, fileHandler FnFileHandler, maxHandlerCount ...int) {
	ch := lm.startDirFileHandlers(maxHandlerCount...)

	for !lm.exit {
		fileSystem := os.DirFS(rootDir)
		fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(path, ext) {
				return nil
			}
			ch <- &chanHandler{
				logFile: p.Join(rootDir, path),
				fileHandler: fileHandler,
			}
			return nil
		})
		time.Sleep(lm.sleepDuration)
	}
	close(ch)
	log.Printf("[logmerger] I will exit\n")
}

func (lm *LogMerger) startDirFileHandlers(maxHandlerCount ...int) (chan *chanHandler)  {
	maxHandlers := 3
	if len(maxHandlerCount) > 0 && maxHandlerCount[0] > 0 {
		maxHandlers = maxHandlerCount[0]
	}

	ch := make(chan *chanHandler, maxHandlers)
	for i:=0; i<maxHandlers; i++ {
		go func(lm *LogMerger, i int, ch <-chan *chanHandler) {
			for h := range ch {
				lm.runPatternFile(h.logFile, h.fileHandler)
			}
		}(lm, i, ch)
	}
	return ch
}

func (lm *LogMerger) runPatternFile(logFile string, fileHandler FnFileHandler) {
	reuseLogFile := fmt.Sprintf("%s%s", logFile, _REUSE_SUFFIX)
	var lf string
	if _, err := os.Stat(reuseLogFile); err == nil {
		lf = reuseLogFile
	} else if _, err := os.Stat(logFile); err == nil {
		lf = logFile
	} else {
		return
	}
	lm.processLogFile(lf, fileHandler, false)
}

