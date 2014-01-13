package main

import (
	"io"
	"os/exec"
	"path/filepath"
)

type debuilder struct {
	SourceDir string

	UpstreamPackage bool

	Output io.Writer
}

func newDebuilder(workingDir string, upstreamPackage bool, output io.Writer) (res *debuilder) {
	res = new(debuilder)
	res.SourceDir = filepath.Join(workingDir, "src")
	res.UpstreamPackage = upstreamPackage
	res.Output = output

	return
}

func (d *debuilder) build() error {
	args := []string{}

	// Do not check orig file since we want to do it ourselves
	args = append(args, "--no-tgz-check")

	// args = append(args, "-d")

	args = append(args, "-k" + flSignKey)

	args = append(args, "-S")

	if d.UpstreamPackage {
		args = append(args, "-sd")
	} else {
		args = append(args, "-sa")
	}

	cmd := exec.Command("debuild", args...)
	cmd.Stdout = d.Output
	cmd.Stderr = d.Output

	cmd.Dir = d.SourceDir

	return cmd.Run()
}
