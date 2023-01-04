package libdatax

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var BuildCommand = &cli.Command{
	Name:  "build",
	Usage: "build job json file for datax",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "jobs",
			Usage:    "datax job dir or file, support YAML,JSON",
			Value:    "",
			Required: true,
			Aliases:  []string{"j"},
		},
		&cli.StringFlag{
			Name:    "template",
			Usage:   "template file, support YAML,JSON",
			Value:   "",
			Aliases: []string{"t"},
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
			Value:   "dist",
			Aliases: []string{"o"},
		},
	},
	Action: doBuildAction,
}

type BuildConfig struct {
	Jobs     string
	Template string
	env      string
	Output   string
}

func parseBuildConfig(ctx *cli.Context) BuildConfig {
	return BuildConfig{
		Jobs:     ctx.String("jobs"),
		Template: ctx.String("template"),
		env:      ctx.String("env"),
		Output:   ctx.String("output"),
	}
}

func doBuildAction(ctx *cli.Context) error {
	config := parseBuildConfig(ctx)

	fi, err := os.Stat(config.Jobs)
	if err != nil {
		return errors.New(fmt.Sprintf("%s is not exists", config.Jobs))
	}
	if fi.IsDir() {
		return buildJobs(config, config.Jobs, config.Output)
	} else {
		return buildJob(config, config.Jobs, config.Output)
	}
}

func buildJobs(config BuildConfig, jobsDir, parentDir string) (err error) {
	infos, err := ioutil.ReadDir(jobsDir)
	if err != nil {
		return err
	}
	for _, info := range infos {
		name := path.Join(jobsDir, info.Name())
		if info.IsDir() {
			err = buildJobs(config, name, path.Join(parentDir, info.Name()))
		} else {
			err = buildJob(config, name, parentDir)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func buildJob(config BuildConfig, filepath, outputDir string) (err error) {
	fullName := path.Base(filepath)
	ext := path.Ext(filepath)
	baseName := strings.TrimSuffix(fullName, ext)
	outputPath := path.Join(outputDir, baseName+".json")

	srcData, err := ReadJsonOrYaml(filepath)
	if err != nil {
		return errors.New(fmt.Sprintf("read file [%s] error: %v", filepath, err))
	}
	if config.Template != "" {
		templateData, err := ReadJsonOrYaml(config.Template)
		if err != nil {
			return errors.New(fmt.Sprintf("read template file [%s] error: %v", config.Template, err))
		}
		// merge template
		srcData = JsonMerge(templateData.(map[string]interface{}), srcData.(map[string]interface{}))
	}

	jsonBuf, err := json.MarshalIndent(srcData, "", "  ")
	if err != nil {
		return err
	}
	jsonStr := string(jsonBuf)

	if config.env != "" {
		envMap, err := ReadEnvFile(config.env)
		if err != nil {
			return errors.New(fmt.Sprintf("read env file [%s] error: %v", config.env, err))
		}
		if envMap != nil {
			for k, v := range envMap {
				jsonStr = strings.ReplaceAll(jsonStr, fmt.Sprintf("${%s}", k), v)
			}
		}
	}

	return SaveStringFile(jsonStr, outputPath)
}
