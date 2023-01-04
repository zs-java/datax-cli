package libdatax

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
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

func JsonMerge(dst, src map[string]interface{}) map[string]interface{} {
	return jsMerge(dst, src, 0)
}

func jsMerge(dst, src map[string]interface{}, depth int) map[string]interface{} {
	jsonMergeDepth := 32
	if depth > jsonMergeDepth {
		return dst
		// panic("too deep!")
	}

	for key, srcVal := range src {

		if dstVal, ok := dst[key]; ok {

			srcMap, srcMapOk := jsMapify(srcVal)
			dstMap, dstMapOk := jsMapify(dstVal)

			if srcMapOk && dstMapOk {
				srcVal = jsMerge(dstMap, srcMap, depth+1)
			} else {
				srcMapArr, srcMapArrOk := parseMapArray(srcVal)
				dstMapArr, dstMapArrOk := parseMapArray(dstVal)
				if srcMapArrOk && dstMapArrOk && len(srcMapArr) == len(dstMapArr) {
					arr := make([]map[string]interface{}, len(srcMapArr))
					for i := 0; i < len(srcMapArr); i++ {
						arr[i] = jsMerge(srcMapArr[i], dstMapArr[i], depth+1)
					}
					srcVal = arr
				}
			}
		}

		dst[key] = srcVal
	}

	return dst
}

func jsMapify(i interface{}) (map[string]interface{}, bool) {

	value := reflect.ValueOf(i)

	if value.Kind() == reflect.Map {

		m := map[string]interface{}{}

		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}

		return m, true
	}

	return map[string]interface{}{}, false
}

func TestArray(t *testing.T) {
	data := readJson("arr.json")
	array, ok := parseMapArray(data)
	fmt.Println(ok, array)
}

func parseMapArray(data interface{}) ([]map[string]interface{}, bool) {
	value := reflect.ValueOf(data)

	var arr []map[string]interface{}
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice, reflect.Array:
		if value.Len() == 0 {
			return nil, false
		}
		for i := 0; i < value.Len(); i++ {
			mapify, isMap := jsMapify(value.Index(i).Interface())
			if !isMap {
				return nil, false
			}
			arr = append(arr, mapify)
		}
	}
	return arr, true
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
