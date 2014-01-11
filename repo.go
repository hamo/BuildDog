package main

import (
	"strings"
	"errors"
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
		t.RepoWorker = newBzrRepo(t.Repo, t.Rev, t.WorkingDir, t.Output)
	}
	return nil
}
		











