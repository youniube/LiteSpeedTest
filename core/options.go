package core

// Options describes runtime startup options for a local LiteSpeedTest instance.
type Options struct {
	LocalHost      string
	LocalPort      int
	Link           string
	Ping           int
	Engine         string
	SingboxBin     string
	SingboxWorkDir string
	KeepTempFile   bool
}
