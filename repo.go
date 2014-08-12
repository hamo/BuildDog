package main

import (
	"errors"
	"strings"
)

type repoWorker interface {
	checkout() error
}

func (t *task) parseRepo() error {
	if t.Repo == "" {
		return errors.New("Repo needed")
	}

	switch {
	case strings.HasPrefix(t.Repo, "lp:"):
		fallthrough
	case strings.HasPrefix(t.Repo, "bzr+ssh://"):
		t.RepoWorker = newBzrRepo(t.Repo, t.Rev, t.WorkingDir, t.Output)
	default:
		return errors.New("Repo format error")
	}
	return nil
}
