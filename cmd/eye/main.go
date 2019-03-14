/*-
 * Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
 * Copyright (c) 2018, 1&1 Internet SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main // import "github.com/solnx/eye/cmd/eye"

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	rt "runtime"
	"runtime/pprof"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/client9/reopen"
	"github.com/droundy/goopt"
	"github.com/mjolnir42/erebos"
	"github.com/solnx/eye/internal/eye"
	"github.com/solnx/eye/internal/eye.mock"
	"github.com/solnx/eye/internal/eye.rest"
)

// eyeVersion is the version string set by make
var eyeVersion string

func init() {
	logrus.SetOutput(os.Stderr)
	erebos.SetLogrusOptions()
	rest.EyeVersion = eyeVersion
	log.SetOutput(os.Stderr)
}

var cpuprofile = goopt.String([]string{`--cpuprofile`}, ``, `write cpu profile to file`)
var memprofile = goopt.String([]string{`--memprofile`}, ``, `write memory profile to file`)

func main() {
	os.Exit(daemon())
}

func daemon() int {
	var err error
	var configurationFile string
	var lfhApp *reopen.FileWriter

	goopt.Version = eyeVersion
	goopt.Suite = `eye`
	goopt.Summary = `Configuration Lookup service`
	goopt.Author = `Jörg Pernfuß`
	goopt.Description = func() string {
		return "Eye stores threshold configuration profiles in a " +
			"Format suitable for Cyclone metric monitoring."
	}

	cliConfPath := goopt.String([]string{`-c`, `--config`}, `/srv/eye/conf/eye.conf`, `Configuration file`)
	goopt.Parse(nil)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		rt.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
	run := runtime{}
	run.logFileMap = &eye.LogHandleMap{}
	run.logFileMap.Init()

	run.conf = &erebos.Config{}

	// read configuration file
	if configurationFile, err = filepath.Abs(*cliConfPath); err != nil {
		logrus.Fatal(err)
	}
	if configurationFile, err = filepath.EvalSymlinks(configurationFile); err != nil {
		logrus.Fatal(err)
	}
	if err = run.conf.FromFile(configurationFile); err != nil {
		logrus.Fatal(err)
	}
	panicLog, err := os.OpenFile(filepath.Join(run.conf.Log.Path, `panic.log`), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		logrus.Fatal(err)
	}
	redirectStderr(panicLog)

	// open application logfile
	run.appLog = logrus.New()
	if lfhApp, err = reopen.NewFileWriter(
		filepath.Join(run.conf.Log.Path, `eye.log`),
	); err != nil {
		logrus.Fatal(`Unable to open application log: `, err)
	}
	run.appLog.Out = lfhApp
	run.logFileMap.Add(`application`, lfhApp)

	// print startup header in all logfiles
	logrus.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)
	run.appLog.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)
	switch strings.ToLower(run.conf.Log.LogLevel) {
	case `trace`:
		run.appLog.SetLevel(logrus.TraceLevel)
	case `debug`:
		run.appLog.SetLevel(logrus.DebugLevel)
	case `info`:
		run.appLog.SetLevel(logrus.InfoLevel)
	case `warning`:
		run.appLog.SetLevel(logrus.WarnLevel)
	case `error`:
		run.appLog.SetLevel(logrus.ErrorLevel)
	case `fatal`:
		run.appLog.SetLevel(logrus.FatalLevel)
	case `panic`:
		run.appLog.SetLevel(logrus.PanicLevel)
	default:
		run.appLog.SetLevel(logrus.InfoLevel)
	}

	// signal handler will reopen all logfiles on USR2
	sigChanLogRotate := make(chan os.Signal, 1)
	signal.Notify(sigChanLogRotate, syscall.SIGUSR2)
	go run.logrotate(sigChanLogRotate)
	run.appLog.Println(`Listening for logrotate requests on SIGUSR2`)

	// initialize database
	run.connectDatabase()
	go run.pingDatabase()

	// handler map shared between eye.Eye and rest.Rest
	hm := eye.HandlerMap{}
	hm.Init()
	// start application
	app := eye.New(&hm, run.conn, run.conf, run.appLog)
	app.Start()

	// start REST API
	rst := rest.New(mock.AlwaysAuthorize, &hm, run.conf, run.conn, run.appLog)
	if rst == nil {
		run.appLog.Fatal(fmt.Errorf("could not initialize rest endpoints"))
	}
	go rst.Run()
	sigChanKill := make(chan os.Signal, 1)
	signal.Notify(sigChanKill, syscall.SIGTERM, syscall.SIGINT)
	<-sigChanKill
	return 0
}

// redirectStderr to the file passed in
func redirectStderr(f *os.File) {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
