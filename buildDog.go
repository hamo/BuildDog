package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var version int = 1

var (
	flPort    int
	flDir     string
	flSignKey string
)

var workingDir string

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.IntVar(&flPort, "port", 8888, "server port")
	flag.StringVar(&flDir, "dir", os.TempDir(), "building dir")
	flag.StringVar(&flSignKey, "key", "", "Sign GPG key")
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

	flag.Parse()

	if flSignKey == "" {
		flag.Usage()
	}

	workingDir = filepath.Join(flDir, "BuildDog", strconv.Itoa(version))

	initWorkingDir()
	initTaskPool()

	serv := newAPIServer()
	panic(serv.ListenAndServe())
}
