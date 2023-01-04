package libdatax

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

func TestMerge(t *testing.T) {

	// a := readJson("a.json")
	// b := readJson("b.json")
	a := readYaml("src/job/test_table1.yml")
	// a := readJson("test.json")
	b := readJson("src/template/mysql2mysql.template.json")
	fmt.Println(a, b)

	result := JsonMerge(a.(map[string]interface{}), b.(map[string]interface{}))

	// result, info := jsonmerge.Merge(a, b)
	// if len(info.Errors) > 0 {
	// 	panic(info)
	// }
	_ = saveJson(result, "result.json")
}

func TestReadYaml(t *testing.T) {
	data := readYaml("src/job/test_table1.yml")
	fmt.Println(data)
	_ = saveJson(data, "test.json")
}

func readJson(path string) (data interface{}) {
	file, _ := os.Open(path)
	defer file.Close()
	buf, _ := ioutil.ReadAll(file)
	_ = json.Unmarshal(buf, &data)
	return data
}

func readYaml(path string) (data interface{}) {
	file, _ := ioutil.ReadFile(path)
	_ = yaml.Unmarshal(file, &data)
	data = convert(data)
	return data
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func saveJson(data interface{}, path string) error {
	file, _ := os.Create(path)
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func TestArray(t *testing.T) {
	data := readJson("arr.json")
	array, ok := parseMapArray(data)
	fmt.Println(ok, array)
}

func TestDir(t *testing.T) {
	infos, err := ioutil.ReadDir("src/job")
	if err != nil {
		panic(err)
	}
	for _, info := range infos {
		fmt.Println(info.Name())
	}
}
