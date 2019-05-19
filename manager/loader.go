package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		logrus.Fatal("oof")
	}
	args := []string{"load"}
	exeCmd := exec.Command("docker", args...)

	var a *io.PipeReader
	var b *io.PipeWriter
	a, b = io.Pipe()

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		logrus.Fatal(err)
	}

	exeCmd.Stdin = a
	exeCmd.Stdout = os.Stdout
	exeCmd.Stderr = os.Stderr
	startErr := exeCmd.Start()
	if startErr != nil {
		logrus.Fatal(startErr)
	}

	go func() {
		_, err = b.Write(data)
		if err != nil {
			logrus.Fatal(err)
		}
		b.Close()
	}()

	err = exeCmd.Wait()
	if err != nil {
		logrus.Fatal(startErr)
	}
}

