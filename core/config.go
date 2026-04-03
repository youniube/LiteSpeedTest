package core

type Config struct {
	LocalHost      string
	LocalPort      int
	Link           string
	Ping           int
	Engine         string
	SingboxBin     string
	SingboxWorkDir string
	KeepTempFile   bool
}
