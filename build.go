package main

type builder interface {
	build() error
}

func (t *task) parseBuilder() error {
	switch t.PackageInfo.Command {
	case "debuild":
		t.Builder = newDebuilder(t.WorkingDir, t.PackageInfo.UpstreamPackage, t.Output)
	}
	return nil
}
