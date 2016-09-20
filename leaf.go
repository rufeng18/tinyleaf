package tinyleaf

import (
	//"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rufeng18/tinyleaf/cluster"
	"github.com/rufeng18/tinyleaf/conf"
	"github.com/rufeng18/tinyleaf/console"
	"github.com/rufeng18/tinyleaf/log"
	"github.com/rufeng18/tinyleaf/module"
)

var logger *log.Logger

func Run(mods ...module.Module) {
	// logger
	if conf.LogLevel != "" {
		var err error
		logger, err = log.New(conf.LogLevel, conf.LogPath)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("mmoBay server %v starting up", version)

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// cluster
	cluster.Init()

	// console
	console.Init()

	// control the singal
	ch := make(chan bool, 1)
	installSignal(ch)
	<-ch

	log.Release("mmoBay server closing down ")
	console.Destroy()
	cluster.Destroy()
	module.Destroy()

}

func reload() {
	log.Release("mmoBay server reload config ")
	module.Reload()
	logger.SetLoggerLevel(conf.LogLevel)
}

func installSignal(ctrl chan bool) {

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGPIPE, syscall.SIGHUP)
	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGHUP:
				reload()
			case syscall.SIGPIPE:
			default:
				ctrl <- true
			}
			log.Release("mmoBay server recv signal: %v ", sig)
		}
	}()
}
