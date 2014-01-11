package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type packageInfo struct {
	Command string

	Name         string
	MajorVersion string
	MinorVersion string
	Series       string

	UpstreamPackage bool
}

func (t *task) analyze() error {
	t.PackageInfo = new(packageInfo)

	srcDir := filepath.Join(t.WorkingDir, "src")
	debianDir := filepath.Join(srcDir, "debian")

	// Now, we only support standard debian source format
	t.PackageInfo.Command = "debuild"
	fi, err := os.Stat(debianDir)
	switch {
	case err != nil:
		fallthrough
	case err == nil && !fi.IsDir():
		return err
	default:
		analyzeChangelog(debianDir, t.PackageInfo)
	}

	return nil
}

func analyzeChangelog(debianDir string, pi *packageInfo) error {
	changelogFile := filepath.Join(debianDir, "changelog")
	f, err := os.Open(changelogFile)
	defer f.Close()

	if err != nil {
		return err
	}

	r := bufio.NewReader(f)

	fl, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	fields := strings.Fields(fl)
	if len(fields) != 4 {
		return errors.New("changelog file format error")
	}

	pi.Name = fields[0]

	v := strings.TrimSuffix(strings.TrimPrefix(fields[1], "("), ")")
	vs := strings.SplitN(v, "-", 2)

	pi.MajorVersion = vs[0]
	if len(vs) == 2 {
		pi.MinorVersion = vs[1]
	}

	pi.Series = strings.TrimSuffix(fields[2], ";")

	// Now check upstream package
	// 1. If format is 3.0 (quilt), then this is one upstream package
	formatFile := filepath.Join(debianDir, "source", "format")
	format, err := os.Open(formatFile)
	defer format.Close()
	if err == nil {
		fr := bufio.NewReader(format)
		debFormat, err := fr.ReadString('\n')
		if err == nil || err == io.EOF {
			if strings.Contains(debFormat, "quilt") {
				pi.UpstreamPackage = true
			}
		}
	}

	// FIXME: Add more upstream testings

	return nil

}
