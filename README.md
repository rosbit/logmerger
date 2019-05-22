# logmerger

A utility package which can be used to process log file asynchronously.

## Archtecture
```
                    +---------------+                    +---------------------+ <--- log-files such as:
  log-generator --->+   log-file    +---> logmerger ---> + log-files-with-date |      log-file_20190522
                    +---------------+        |           +----------+----------+      log-file_20190523
                                             V                      ^                 log-file_2019....
                                +------------+-------------+        |
                                | current-log-file-content +--------+
                                +------------+-------------+
                                             |
                                             V
                                    one-log-file-handler
```

## Usage

```go
package main

import (
	lm "github.com/rosbit/logmerger"
	"fmt"
)

func main() {
	m := lm.NewLogMerger(100)  // time interval to check the existence of log-file
	m.Run("log-file-name.here", sampleLogFileHandler)  // use go m.Run(...) if you don't want to block
}

func sampleLogFileHandler(fileName string) {
	fmt.Printf("process file %s as you need here\n", fileName)
	// make sure the log-file not be opened
}
```

## Status

The package is fully tested.
