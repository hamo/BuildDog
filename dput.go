package main

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	globalConf  = "/etc/dput.cf"
	peruserConf = filepath.Join(os.Getenv("HOME"), ".dput.cf")
)

func (t *task) dput() error {
	if err := runDput(t.WorkingDir, t.PPA, t.Output); err != nil {
		return err
	}
	return nil
}

func runDput(workingDir string, PPA string, output io.Writer) error {
	if err := checkPPA(PPA); err != nil {
		return err
	}

	var args = []string{"-d"}

	args = append(args, "-f")

	args = append(args, PPA)

	// FIXME: Generate .changes name instead of iterate working
	// dir
	changesFile := findChangesFile(workingDir)
	if changesFile == "" {
		return errors.New("Can not find .changes file")
	}
	// target := filepath.Join(workingDir, changesFile)
	// args = append(args, target)
	args = append(args, changesFile)

	cmd := exec.Command("dput", args...)
	cmd.Dir = workingDir
	cmd.Stdout = output
	cmd.Stderr = output

	return cmd.Run()
}

func checkPPA(PPA string) error {
	if strings.HasPrefix(PPA, "ppa:") {
		return nil
	}

	p3a := getPrivatePPA()
	for _, p := range p3a {
		if PPA == p {
			return nil
		}
	}

	return errors.New("unknown PPA")
}

func getPrivatePPA() (res []string) {
	res = make([]string, 10)

	// First, read global dput conf
	gf, err := os.Open(globalConf)
	if err != nil {
		goto peruser
	}
	defer gf.Close()

	res = append(res, getPrivatePPAFromConf(gf)...)

peruser:
	pf, err := os.Open(peruserConf)
	if err != nil {
		return res
	}
	defer pf.Close()

	res = append(res, getPrivatePPAFromConf(pf)...)

	return res
}

func getPrivatePPAFromConf(f *os.File) (res []string) {
	res = make([]string, 5)
	reader := bufio.NewReader(f)

	re := regexp.MustCompilePOSIX("^\\[(.+)\\]$")

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		if match := re.FindStringSubmatch(line); len(match) == 2 {
			res = append(res, match[1])
		}

		if err == io.EOF {
			break
		}
	}
	return res
}

func findChangesFile(workingDir string) string {
	fis, err := ioutil.ReadDir(workingDir)
	if err != nil {
		return ""
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(fi.Name(), ".changes") {
			return fi.Name()
		}
	}

	return ""
}
