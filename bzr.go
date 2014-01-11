package main

import (
	"path/filepath"
	"io"
	"os/exec"
)

type bzrRepo struct {
	Repo string
	Rev  string
	
	SourceDir string

	Output io.Writer
}

func newBzrRepo(repo, rev, workingDir string, output io.Writer) *bzrRepo {
	r := new(bzrRepo)
	r.Repo = repo
	r.Rev = rev 
	r.SourceDir = filepath.Join(workingDir, "src")
	r.Output = output
	return r
}

func (r *bzrRepo) checkout() error {
	args :=[]string{"branch", "-v"}
	if r.Rev !=  "" {
		args = append(args, "-r", r.Rev)
	}
	args = append(args, r.Repo)
	args = append(args, r.SourceDir)
	
	cmd := exec.Command("bzr", args...)
	cmd.Stdout = r.Output
	cmd.Stderr = r.Output
	
	return cmd.Run()
}

		
		
	








