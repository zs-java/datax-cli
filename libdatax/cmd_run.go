package libdatax

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var RunCommand = &cli.Command{
	Name:  "run",
	Usage: "run datax jobs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "jobs",
			Usage:   "datax job dir or file",
			Value:   "dist",
			Aliases: []string{"j"},
		},
		&cli.StringFlag{
			Name:    "env",
			Usage:   "env file",
			Value:   "",
			Aliases: []string{"e"},
		},
		&cli.StringFlag{
			Name:    "output",
			Usage:   "result output dir",
			Value:   "logs",
			Aliases: []string{"o"},
		},
		&cli.StringFlag{
			Name:    "datax-home",
			Usage:   "datax home, default read env DATAX_HOME",
			Value:   os.Getenv("DATAX_HOME"),
			Aliases: []string{"d"},
		},
		&cli.StringFlag{
			Name:  "loglevel",
			Usage: "datax log level",
			Value: "info",
		},
	},
	Action: doRunAction,
}

type RunConfig struct {
	Jobs      string
	Env       string
	Output    string
	DataxHome string
	Loglevel  string
}

func ParseRunConfig(ctx *cli.Context) RunConfig {
	return RunConfig{
		Jobs:      ctx.String("jobs"),
		Env:       ctx.String("env"),
		Output:    ctx.String("output"),
		DataxHome: ctx.String("datax-home"),
		Loglevel:  ctx.String("loglevel"),
	}
}

func doRunAction(ctx *cli.Context) error {
	config := ParseRunConfig(ctx)

	err := checkDataxHome(config)
	if err != nil {
		return err
	}

	// logfile
	logFile, err := getLogfile(config)
	if err != nil {
		return err
	}
	defer func() {
		_ = logFile.Close()
	}()

	stdout := NewMultipleWriter(logFile, os.Stdout)
	stderr := NewMultipleWriter(logFile, os.Stderr)

	fi, err := os.Stat(config.Jobs)
	if err != nil {
		return errors.New(fmt.Sprintf("%s is not exists", config.Jobs))
	}
	if fi.IsDir() {
		return runJobs(config, config.Jobs, stdout, stderr)
	} else {
		return runJob(config, config.Jobs, stdout, stderr)
	}
}

func runJobs(config RunConfig, inputDir string, stdout, stderr io.Writer) (err error) {
	infos, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, info := range infos {
		name := path.Join(inputDir, info.Name())
		if info.IsDir() {
			err = runJobs(config, name, stdout, stderr)
		} else {
			err = runJob(config, name, stdout, stderr)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func runJob(config RunConfig, filepath string, stdout, stderr io.Writer) (err error) {
	var envMap map[string]string
	if config.Env != "" {
		if envMap, err = ReadEnvFile(config.Env); err != nil {
			return err
		}
	}

	var jvmOpts string
	if envMap != nil {
		for k, v := range envMap {
			jvmOpts += fmt.Sprintf(" -D%s='%s'", k, v)
		}
	}

	var args []string
	if jvmOpts != "" {
		args = append(args, "-p", jvmOpts)
	}
	if config.Loglevel != "" {
		args = append(args, "--loglevel", config.Loglevel)
	}

	command := exec.Command(config.DataxHome+"/bin/datax.py", append(args, filepath)...)

	command.Stdin = os.Stdin
	command.Stdout = stdout
	command.Stderr = stderr

	_, _ = fmt.Fprintf(stdout, "========= begin job: %s, args: %s\n", filepath, strings.Join(command.Args[1:], " "))

	err = command.Run()

	_, _ = fmt.Fprintf(stdout, "========= end job: %s\n", filepath)

	return err
}

func checkDataxHome(config RunConfig) (err error) {
	if config.DataxHome == "" {
		err = errors.New("not found datax-home")
	}
	return err
}

func getLogfile(config RunConfig) (file *os.File, err error) {
	logfilePath := path.Join(config.Output, time.Now().Format("2006-01-02-150405")+".log")
	dir := path.Dir(logfilePath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return file, err
	}
	return os.Create(logfilePath)
}
