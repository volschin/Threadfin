// Copyright 2019 marmei. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the
// LICENSE file.
// GitHub: https://github.com/Threadfin/Threadfin

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"threadfin/src"
)

// GitHubStruct : GitHub Account. Über diesen Account werden die Updates veröffentlicht
type GitHubStruct struct {
	Branch  string
	Repo    string
	Update  bool
	User    string
	TagName string
}

// GitHub : GitHub Account
// If you want to fork this project, enter your Github account here. This prevents a newer version of Threadfin from updating your version.
var GitHub = GitHubStruct{Branch: "Main", User: "Threadfin", Repo: "Threadfin", Update: true}

/*
	Branch: GitHub Branch
	User: 	GitHub Username
	Repo: 	GitHub Repository
	Update: Automatic updates from the GitHub repository [true|false]
*/

// Name : Programmname
const Name = "Threadfin"

// Version : Version, die Build Nummer wird in der main func geparst.
const Version = "3.2.1"

// DBVersion : Datanbank Version
const DBVersion = "0.5.0"

// APIVersion : API Version
const APIVersion = "3.2.1"

var homeDirectory = fmt.Sprintf("%s%s.%s%s", src.GetUserHomeDirectory(), string(os.PathSeparator), strings.ToLower(Name), string(os.PathSeparator))
var samplePath = fmt.Sprintf("%spath%sto%sthreadfin%s", string(os.PathSeparator), string(os.PathSeparator), string(os.PathSeparator), string(os.PathSeparator))
var sampleRestore = fmt.Sprintf("%spath%sto%sfile%s", string(os.PathSeparator), string(os.PathSeparator), string(os.PathSeparator), string(os.PathSeparator))

var configFolder = flag.String("config", "", ": Config Folder        ["+samplePath+"] (default: "+homeDirectory+")")
var port = flag.String("port", "", ": Server port          [34400] (default: 34400)")
var restore = flag.String("restore", "", ": Restore from backup  ["+sampleRestore+"threadfin_backup.zip]")

var gitBranch = flag.String("branch", "", ": Git Branch           [main|beta] (default: main)")
var debug = flag.Int("debug", 0, ": Debug level          [0 - 3] (default: 0)")
var info = flag.Bool("info", false, ": Show system info")
var h = flag.Bool("h", false, ": Show help")

// Aktiviert den Entwicklungsmodus. Für den Webserver werden dann die lokalen Dateien verwendet.
var dev = flag.Bool("dev", false, ": Activates the developer mode, the source code must be available. The local files for the web interface are used.")
var bindIpAddress = flag.String("bind", "", ": Bind IP address")

func main() {
	os.Exit(dispatchThreadfinStartup(os.Args, src.PrepareUpdateStartup, runThreadfinApplication))
}

func dispatchThreadfinStartup(
	args []string,
	prepare func([]string) (src.UpdateStartup, error),
	application func([]string, bool) int,
) int {
	startup, err := prepare(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if startup.ExitCode != 0 {
			return startup.ExitCode
		}
		return 1
	}
	if startup.Private && startup.Exit {
		return startup.ExitCode
	}
	applicationArgs := append([]string(nil), args...)
	if startup.Private {
		if len(args) == 0 {
			return 1
		}
		applicationArgs = append([]string{args[0]}, startup.OriginalArgs...)
	}
	return application(applicationArgs, startup.SkipAutomaticUpdate)
}

func performStartupUpdate(skip bool, update func() error) error {
	if skip || src.System.ConfigurationWizard {
		return nil
	}
	return update()
}

func runThreadfinApplication(args []string, skipAutomaticUpdate bool) (exitCode int) {
	os.Args = append([]string(nil), args...)

	// Build-Nummer von der Versionsnummer trennen
	var build = strings.Split(Version, ".")

	var system = &src.System
	system.APIVersion = APIVersion
	system.Branch = strings.ToTitle(GitHub.Branch)
	system.Build = build[len(build)-1:][0]
	system.DBVersion = DBVersion
	system.GitHub = GitHub
	system.Name = Name
	system.Version = strings.Join(build[0:len(build)-1], ".")

	// Panic !!!
	defer func() {

		if r := recover(); r != nil {
			exitCode = 1

			fmt.Println()
			fmt.Println("* * * * * FATAL ERROR * * * * *")
			fmt.Println("OS:  ", runtime.GOOS)
			fmt.Println("Arch:", runtime.GOARCH)
			fmt.Println("Err: ", r)
			fmt.Println()

			pc := make([]uintptr, 20)
			runtime.Callers(2, pc)

			for i := range pc {

				if runtime.FuncForPC(pc[i]) != nil {

					f := runtime.FuncForPC(pc[i])
					file, line := f.FileLine(pc[i])

					if len(file) > 0 && string(file)[0:1] != "?" {
						fmt.Printf("%s:%d %s\n", filepath.Base(file), line, f.Name())
					}

				}

			}

			fmt.Println()
			fmt.Println("* * * * * * * * * * * * * * * *")

		}

	}()

	flag.Parse()

	if *h {
		flag.Usage()
		return 0
	}

	system.Dev = *dev

	// Systeminformationen anzeigen
	if *info {

		system.Flag.Info = true

		err := src.Init()
		if err != nil {
			src.ShowError(err, 0)
			return 1
		}

		src.ShowSystemInfo()
		return 0

	}

	// Webserver Port
	if len(*port) > 0 {
		system.Flag.Port = *port
	}

	if bindIpAddress != nil && len(*bindIpAddress) > 0 {
		system.Flag.Bind = *bindIpAddress
	}

	// Branch
	system.Flag.Branch = *gitBranch
	if len(system.Flag.Branch) > 0 {
		fmt.Println("Git Branch is now:", system.Flag.Branch)
	}

	// Debug Level
	system.Flag.Debug = *debug
	if system.Flag.Debug > 3 {
		flag.Usage()
		return 1
	}

	// Speicherort für die Konfigurationsdateien
	if len(*configFolder) > 0 {
		system.Folder.Config = *configFolder
	}

	// Backup wiederherstellen
	if len(*restore) > 0 {

		system.Flag.Restore = *restore

		err := src.Init()
		if err != nil {
			src.ShowError(err, 0)
			return 1
		}

		err = src.ThreadfinRestoreFromCLI(*restore)
		if err != nil {
			src.ShowError(err, 0)
			return 1
		}

		return 0
	}

	err := src.Init()
	if err != nil {
		src.ShowError(err, 0)
		return 1
	}

	err = performStartupUpdate(skipAutomaticUpdate, src.BinaryUpdate)
	if src.HandleBinaryUpdateResult(err, os.Exit, func(err error) {
		src.ShowError(err, 0)
	}) {
		return 0
	}

	err = src.StartSystem(false)
	if err != nil {
		src.ShowError(err, 0)
		return 1
	}

	err = src.InitMaintenance()
	if err != nil {
		src.ShowError(err, 0)
		return 1
	}

	err = src.StartWebserver()
	if err != nil {
		src.ShowError(err, 0)
		return 1
	}

	return 0
}

func getPIDs(command string) ([]string, error) {
	var out bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	pids := strings.Fields(out.String())
	return pids, nil
}

// killProcess kills a process by its PID
func killProcess(pid string) error {
	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}
