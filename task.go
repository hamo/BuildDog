package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	taskLogSep = strings.Repeat("=", 30)
)

var (
	StatusNew     = "New"
	StatusRunning = "Running"
	StatusError   = "Error"
	StatusFinish  = "Finish"
)

var (
	ProcessRepo    = "Repo"
	ProcessAnalyze = "Analyze"
	ProcessBuild   = "Build"
	ProcessDput    = "Dput"
)

type task struct {
	Creator string

	Repo string
	Rev  string
	PPA  string

	Status  string
	Process string
	Error   string

	WorkingDir string

	RepoWorker repoWorker `json:"-"`

	PackageInfo *packageInfo `json:"-"`
	Builder     builder      `json:"-"`

	Output     *bytes.Buffer `json:"-"`
	CreateTime time.Time     `json:"-"`
}

type taskPool struct {
	lock   sync.Mutex
	NextId uint64
	Pool   map[uint64]*task
}

var tasks taskPool

func initTaskPool() {
	tasks.Pool = make(map[uint64]*task)
}

func newTask(args map[string]string) *task {
	var t task
	t.Creator = args["creator"]
	t.Repo = args["repo"]
	t.PPA = args["ppa"]
	t.Rev = args["rev"]
	t.Output = new(bytes.Buffer)
	t.CreateTime = time.Now()
	t.Status = StatusNew

	return &t
}

func getTaskById(id uint64) *task {
	tasks.lock.Lock()
	if v, ok := tasks.Pool[id]; ok {
		tasks.lock.Unlock()
		return v
	} else {
		tasks.lock.Unlock()
		return nil
	}
}

func (t *task) enqueue() uint64 {
	tasks.lock.Lock()
	id := tasks.NextId
	tasks.Pool[id] = t
	tasks.NextId = tasks.NextId + 1
	tasks.lock.Unlock()

	t.WorkingDir = filepath.Join(workingDir, fmt.Sprintf("%d", id))
	os.MkdirAll(t.WorkingDir, 0755)

	go t.process()

	return id

}

func GenOutputSep(content string) string {
	re := taskLogSep
	re = re + " "
	re = re + content
	re = re + " "
	re = re + taskLogSep
	re = re + "\n"
	return re
}

func (t *task) process() {
	if t.Status != StatusNew {
		panic("err")
	}

	// 1. Check out source code
	t.Output.WriteString(GenOutputSep("Enter " + ProcessRepo))
	t.Status = StatusRunning
	t.Process = ProcessRepo

	t.parseRepo()
	if err := t.RepoWorker.checkout(); err != nil {
		t.Status = StatusError
		t.Error = err.Error()
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessRepo))

	// 2. Analyze debian dir
	t.Output.WriteString(GenOutputSep("Enter " + ProcessAnalyze))
	t.Process = ProcessAnalyze
	if err := t.analyze(); err != nil {
		t.Status = StatusError
		t.Error = err.Error()
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessAnalyze))

	// 3. Build
	t.Output.WriteString(GenOutputSep("Enter " + ProcessBuild))
	t.Process = ProcessBuild
	t.parseBuilder()
	if err := t.Builder.build(); err != nil {
		t.Status = StatusError
		t.Error = err.Error()
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessBuild))

	// 4. Sign and dput the changes file
	t.Output.WriteString(GenOutputSep("Enter " + ProcessDput))
	t.Process = ProcessDput
	if err := t.dput(); err != nil {
		t.Status = StatusError
		t.Error = err.Error()
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessDput))

	t.Status = StatusFinish
	os.RemoveAll(t.WorkingDir)

}
