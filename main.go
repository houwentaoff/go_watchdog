package main

import (
	"flag"
	"go_watchdog/common"
	"go_watchdog/mail"
	"go_watchdog/watchdog"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

var (
	serviceName   = flag.String("s", "", "service name to watch")
	logFile       = flag.String("l", "log", "log fle name")
	checkInterval = flag.String("c", "60s", "Check isterval [duration string]")
	startInterval = flag.String("i", "10s", "Start interval [duration string]")
	attempts      = flag.Int("a", 4, "Number of attempts when starting service")
	mailFile      = flag.String("m", "mail.conf", "File name with mail config")
)

func main() {
	flag.Parse()

	conf, err := common.LoadConfig(*mailFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	mailSender, err := mail.NewEmailSender(conf.Mail)
	if err != nil {
		log.Fatal(err)
		return
	}
	service, err := watchdog.NewService(*serviceName, *checkInterval, *startInterval, *attempts)
	if err != nil {
		log.Fatal(err)
	}

	context := daemon.Context{
		LogFileName: *logFile,
		LogFilePerm: 0640,
	}

	child, err := context.Reborn()
	if err != nil {
		log.Fatal(err)
	}
	if child != nil { //nil is returned in child process
		return
	}
	defer context.Release()

	stop := make(chan bool)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		_ = <-sigc
		stop <- true
	}()

	service.Watch(mailSender, stop)
}
