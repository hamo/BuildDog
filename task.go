package main

import (
	"bytes"
	"container/list"
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
	ID      uint64
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
	lock sync.Mutex

	NextId      uint64
	ErrorList   *list.List
	RunningList *list.List

	Pool map[uint64]*task
}

var tasks taskPool

func initTaskPool() {
	tasks.Pool = make(map[uint64]*task)
	tasks.ErrorList = list.New()
	tasks.RunningList = list.New()
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

	t.ID = id
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

func (t *task) setErrorStatus(eID *list.Element, err error) {
	t.Status = StatusError
	t.Error = err.Error()

	tasks.lock.Lock()
	tasks.RunningList.Remove(eID)
	tasks.ErrorList.PushBack(t.ID)
	tasks.lock.Unlock()
}

func (t *task) process() {
	if t.Status != StatusNew {
		panic("err")
	}

	tasks.lock.Lock()
	eID := tasks.RunningList.PushBack(t.ID)
	tasks.lock.Unlock()

	// 1. Check out source code
	t.Output.WriteString(GenOutputSep("Enter " + ProcessRepo))
	t.Status = StatusRunning
	t.Process = ProcessRepo

	if err := t.parseRepo(); err != nil {
		t.setErrorStatus(eID, err)
		return
	}
	if err := t.RepoWorker.checkout(); err != nil {
		t.setErrorStatus(eID, err)
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessRepo))

	// 2. Analyze debian dir
	t.Output.WriteString(GenOutputSep("Enter " + ProcessAnalyze))
	t.Process = ProcessAnalyze
	if err := t.analyze(); err != nil {
		t.setErrorStatus(eID, err)
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessAnalyze))

	// 3. Build
	t.Output.WriteString(GenOutputSep("Enter " + ProcessBuild))
	t.Process = ProcessBuild
	t.parseBuilder()
	if err := t.Builder.build(); err != nil {
		t.setErrorStatus(eID, err)
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessBuild))

	// 4. Sign and dput the changes file
	t.Output.WriteString(GenOutputSep("Enter " + ProcessDput))
	t.Process = ProcessDput
	if err := t.dput(); err != nil {
		t.setErrorStatus(eID, err)
		return
	}
	t.Output.WriteString(GenOutputSep("Leave " + ProcessDput))

	t.Status = StatusFinish

	tasks.lock.Lock()
	tasks.RunningList.Remove(eID)
	tasks.lock.Unlock()

	// print buildlog to stdout
	// style: "[BUILDLOG] time ID user repo[#rev] ppa"
	if t.Rev != "" {
		fmt.Printf("[BUILDLOG] %s %d %s %s#%s %s\n", time.Now().Format(time.RFC3339), t.ID, t.Creator, t.Repo, t.Rev, t.PPA)
	} else {
		fmt.Printf("[BUILDLOG] %s %d %s %s %s\n", time.Now().Format(time.RFC3339), t.ID, t.Creator, t.Repo, t.PPA)
	}

	os.RemoveAll(t.WorkingDir)

}
