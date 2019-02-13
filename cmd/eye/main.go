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
	"os"
	"os/signal"
	"path/filepath"
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
}

func main() {
	os.Exit(daemon())
}

func daemon() int {
	var err error
	var configurationFile string
	var lfhGlobal, lfhApp, lfhReq, lfhErr, lfhAudit *reopen.FileWriter

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
	// open global default logger logfile
	if lfhGlobal, err = reopen.NewFileWriter(
		filepath.Join(run.conf.Log.Path, `global.log`),
	); err != nil {
		logrus.Fatal(`Unable to open global log: `, err)
	}
	logrus.SetOutput(lfhGlobal)
	run.logFileMap.Add(`global`, lfhGlobal)

	// open application logfile
	run.appLog = logrus.New()
	if lfhApp, err = reopen.NewFileWriter(
		filepath.Join(run.conf.Log.Path, `eye.log`),
	); err != nil {
		logrus.Fatal(`Unable to open application log: `, err)
	}
	run.appLog.Out = lfhApp
	run.logFileMap.Add(`application`, lfhApp)

	// open error logfile
	run.errLog = logrus.New()
	if lfhErr, err = reopen.NewFileWriter(
		filepath.Join(run.conf.Log.Path, `error.log`),
	); err != nil {
		logrus.Fatal(`Unable to open error log: `, err)
	}
	run.errLog.Out = lfhErr
	run.logFileMap.Add(`error`, lfhErr)

	// open request logfile
	run.reqLog = logrus.New()
	if lfhReq, err = reopen.NewFileWriter(
		filepath.Join(run.conf.Log.Path, `request.log`),
	); err != nil {
		logrus.Fatal(`Unable to open request log: `, err)
	}
	run.reqLog.Out = lfhReq
	run.logFileMap.Add(`request`, lfhReq)

	// open request logfile
	run.auditLog = logrus.New()
	if lfhAudit, err = reopen.NewFileWriter(
		filepath.Join(run.conf.Log.Path, `audit.log`),
	); err != nil {
		logrus.Fatal(`Unable to open audit log: `, err)
	}
	run.reqLog.Out = lfhAudit
	run.logFileMap.Add(`audit`, lfhAudit)

	// print startup header in all logfiles
	logrus.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)
	run.appLog.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)
	run.errLog.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)
	run.reqLog.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)
	run.auditLog.Printf("Starting EYE configuration lookup service, Eye v%s", eyeVersion)

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
	app := eye.New(&hm, run.conn, run.conf, run.appLog, run.reqLog, run.errLog, run.auditLog)
	app.Start()

	// start REST API
	rst := rest.New(mock.AlwaysAuthorize, &hm, run.conf, run.appLog, run.reqLog, run.errLog, run.auditLog)
	go rst.Run()

	sigChanKill := make(chan os.Signal, 1)
	signal.Notify(sigChanKill, syscall.SIGTERM, syscall.SIGINT)
	<-sigChanKill
	return 0
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
