package main

import (
	"datax-cli/libdatax"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	defer func() {
		err := recover()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(0)
			// panic(err)
		}
	}()

	app := cli.NewApp()
	app.Name = "DataX-cli"
	app.Usage = "simple datax client utils"
	app.Version = "1.0.0"

	app.Commands = []*cli.Command{
		libdatax.BuildCommand,
		libdatax.RunCommand,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

}
