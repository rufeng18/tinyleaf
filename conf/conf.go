package conf

import "time"

var (
	LenStackBuf    = 4096
	VerifyInterval = 30 * time.Second

	// log
	LogLevel string
	LogPath  string

	// console
	ConsolePort   int
	ConsolePrompt string = "mmobay# "
	ProfilePath   string

	// cluster
	ListenAddr      string
	ConnAddrs       []string
	PendingWriteNum int
)
