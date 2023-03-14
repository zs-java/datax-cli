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

var pool *FixedSizeThreadPool
var timeDir = time.Now().Format("2006-01-02-150405")

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
			Name:  "datax-home",
			Usage: "datax home, default read env DATAX_HOME",
			Value: os.Getenv("DATAX_HOME"),
		},
		&cli.StringFlag{
			Name:  "loglevel",
			Usage: "datax log level",
			Value: "info",
		},
		&cli.IntFlag{
			Name:    "thread-size",
			Usage:   "Thread Pool Size",
			Value:   1,
			Aliases: []string{"t"},
		},
		&cli.BoolFlag{
			Name:    "daemon",
			Usage:   "Running As Daemon",
			Value:   false,
			Aliases: []string{"d"},
		},
	},
	Action: doRunAction,
}

type RunConfig struct {
	Jobs       string
	Env        string
	Output     string
	DataxHome  string
	Loglevel   string
	ThreadSize int
	Daemon     bool
}

func ParseRunConfig(ctx *cli.Context) RunConfig {
	return RunConfig{
		Jobs:       ctx.String("jobs"),
		Env:        ctx.String("env"),
		Output:     ctx.String("output"),
		DataxHome:  ctx.String("datax-home"),
		Loglevel:   ctx.String("loglevel"),
		ThreadSize: ctx.Int("thread-size"),
		Daemon:     ctx.Bool("daemon"),
	}
}

func doRunAction(ctx *cli.Context) error {
	config := ParseRunConfig(ctx)

	if config.Daemon {
		return runAsDaemon()
	}

	err := checkDataxHome(config)
	if err != nil {
		return err
	}

	startTime := time.Now()

	// pool
	pool = NewFixedSizeThreadPool(config.ThreadSize)

	// logfile
	logFile, err := getLogfile(config, "info.log")
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
		_ = runJobs(config, config.Jobs, stdout, stderr)
	} else {
		_ = runJob(config, config.Jobs, stdout, stderr)
	}
	pool.Wait()
	fmt.Printf("Execution completed, taking %d second.\n", time.Now().Unix()-startTime.Unix())
	return nil
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
	pool.SubmitTask(Task{
		Name: filepath,
		Action: func() {
			fmt.Println("begin ", filepath)
			var envMap map[string]string
			if config.Env != "" {
				if envMap, err = ReadEnvFile(config.Env); err != nil {
					panic(err)
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

			if config.ThreadSize > 1 {

				file, err := getLogfile(config, filepath+".log")
				if err != nil {
					panic(err)
				}

				defer func() {
					_ = file.Close()
				}()

				stdout = NewMultipleWriter(file, stdout)
				stderr = NewMultipleWriter(file, stderr)
			}

			command.Stdin = os.Stdin
			command.Stdout = stdout
			command.Stderr = stderr

			_, _ = fmt.Fprintf(stdout, "========= begin job: %s, args: %s\n", filepath, strings.Join(command.Args[1:], " "))

			err = command.Run()

			_, _ = fmt.Fprintf(stdout, "========= end job: %s\n", filepath)
		},
	})
	return nil
}

func checkDataxHome(config RunConfig) (err error) {
	if config.DataxHome == "" {
		err = errors.New("not found datax-home")
	}
	return err
}

func getLogDir(config RunConfig) (string, error) {
	dir := path.Join(config.Output, timeDir)
	return dir, nil
}

func getLogfile(config RunConfig, filename string) (file *os.File, err error) {
	dir, err := getLogDir(config)
	if err != nil {
		return nil, err
	}

	var filePath string
	if config.ThreadSize == 1 {
		filePath = dir + "." + filename
	} else {
		filePath = path.Join(dir, filename)
	}

	baseDir := path.Dir(filePath)
	err = os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.Create(filePath)
}

func runAsDaemon() error {
	var newArgs []string
	for _, arg := range os.Args[1:] {
		if arg != "-d" && arg != "--daemon" {
			newArgs = append(newArgs, arg)
		}
	}
	cmd := exec.Command(os.Args[0], newArgs...)
	if err := cmd.Start(); err != nil {
		return errors.New(fmt.Sprintf("start %s failed, error: %v\n", os.Args[0], err))
	}
	fmt.Printf("%s [PID] %d running...\n", os.Args[0], cmd.Process.Pid)
	return nil
}
