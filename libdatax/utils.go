package libdatax

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func ReadJsonFile(path string) (data interface{}, err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &data)
	return data, err
}

func ReadYamlFile(filepath string) (data interface{}, err error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}
	return convertYamlInterface(data), nil
}

func ReadJsonOrYaml(filepath string) (data interface{}, err error) {
	switch strings.ToLower(path.Ext(filepath)) {
	case ".json":
		data, err = ReadJsonFile(filepath)
	case ".yaml", ".yml":
		data, err = ReadYamlFile(filepath)
	default:
		err = errors.New("not supported file: " + filepath)
	}
	return data, err
}

func ReadEnvFile(filepath string) (map[string]string, error) {
	if filepath == "" {
		return nil, nil
	}
	if !Exists(filepath) {
		return nil, errors.New(fmt.Sprintf("config file: [%s] not exists!\n", filepath))
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	data := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		arr := strings.Split(line, "=")
		key := arr[0]
		if env := os.Getenv(key); env != "" {
			data[key] = env
		} else {
			data[key] = strings.Join(arr[1:], "=")
		}
		data[key] = strings.Trim(data[key], "\"")
	}
	return data, nil
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) // os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func SaveJsonFile(data interface{}, filepath string) error {
	dir := path.Dir(filepath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func SaveStringFile(text string, filepath string) error {
	dir := path.Dir(filepath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, []byte(text), os.ModePerm)
}

func convertYamlInterface(data interface{}) interface{} {
	switch x := data.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertYamlInterface(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertYamlInterface(v)
		}
	}
	return data
}
