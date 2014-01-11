package main

import (
	"os"
	"path/filepath"
	"strconv"
	"runtime"
)

var version int = 1

var workingDir = filepath.Join(os.TempDir(), "BuildDog", strconv.Itoa(version))

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initWorkingDir() {
	_, err := os.Stat(workingDir)
	if err == nil || (err != nil && os.IsExist(err)) {
		// WorkingDir exists
		if err := os.RemoveAll(workingDir); err != nil {
			panic("Remove working dir error")
		}
		goto create
	}
	if err != nil && os.IsNotExist(err) {
		goto create
	} else {
		panic("stat working dir error")
	}

create:
	if err := os.MkdirAll(workingDir, 0755); err != nil {
		panic("mkdir working dir error")
	}

}

func main() {
	initWorkingDir()
	initTaskPool()

	serv := newAPIServer()
	panic(serv.ListenAndServe())
}
