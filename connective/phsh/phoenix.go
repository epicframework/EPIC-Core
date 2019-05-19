package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"
	"github.com/sirupsen/logrus"

	"github.com/KernelDeimos/gottagofast/toolparse"

	"group23.local/ericland/anything/interpreter"
)

var logo = `
[31;1m______ _                      _      
| ___ \ |                    (_)     
| |_/ / |__   ___   ___ _ __  ___  __
|  __/| '_ \ / _ \ / _ \ '_ \| \ \/ /
| |   | | | | (_) |  __/ | | | |>  < 
\_|   |_| |_|\___/ \___|_| |_|_/_/\_\
[33;1m                        Phoenix Shell
[0m`

var help = `
Usage:
	phsh [-file=FILE]
`

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	argPtrFile := flag.String("file", "", "A script file")
	argPtrDebug := flag.Bool("debug", false, "Enable debug mode")

	fmt.Fprintln(os.Stderr, logo)

	flag.Parse()

	if *argPtrDebug {
		logrus.Warn("Debug mode is on")
	}

	if *argPtrFile != "" {
		logrus.Fatal("Script execution not supported yet")
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "ERROR SETTING PROMPT",
		HistoryFile:     "/tmp/phsh_history_file.tmp",
		InterruptPrompt: "^SIGINT",
	})
	defer l.Close()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	ifa := interpreter.InterpreterFactoryA{}
	exec := ifa.MakeExec()

	statEndl := "-\033[0m-"

	for {
		statExit := "\033[32;1mOK\033[0m"
		statLoc := "/"
		fmt.Println(statEndl)
		l.SetPrompt(
			statExit + ":" + statLoc + "$ ",
		)
		input, err := l.Readline()
		if err == nil {
			// No action
		} else if err == readline.ErrInterrupt {
			break
		} else if err == io.EOF {
			break
		} else {
			logrus.Error(err)
		}
		input = fmt.Sprintf("%s", input)

		list, err := toolparse.ParseListSimple(input)

		outBytes, _ := json.Marshal(list)
		logrus.Debug(string(outBytes))

		results, err := exec(list)
		if err != nil {
			logrus.Error(err)
			continue
		}

		outBytes, err = json.Marshal(results)
		logrus.Info(string(outBytes))
	}

}
