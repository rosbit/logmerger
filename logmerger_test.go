package logmerger

import (
	"testing"
	"fmt"
	"time"
)

func showFile(fileName string) {
	fmt.Printf("file %s got\n", fileName)
}

func stopMergerLater(m *LogMerger) {
	time.Sleep(2 * time.Minute)
	m.Stop()
}

func Test_merger(t *testing.T) {
	m := NewLogMerger(0)
	go stopMergerLater(m)
	fmt.Printf("please give me the file named text.txt In 2 munites\n")
	m.Run("test.txt", showFile)
}
