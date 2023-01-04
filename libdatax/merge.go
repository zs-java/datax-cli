package libdatax

import "reflect"

const (
	mergeMaxDepth = 32
)

func JsonMerge(target map[string]interface{}, sources ...map[string]interface{}) map[string]interface{} {
	// return jsMerge(dst, src, 0)
	if len(sources) == 0 {
		return target
	}
	var result map[string]interface{}
	for _, src := range sources {
		result = jsMerge(target, src, 0)
	}
	return result
}

func jsMerge(dst, src map[string]interface{}, depth int) map[string]interface{} {
	if depth > mergeMaxDepth {
		return dst
	}

	for key, srcVal := range src {

		if dstVal, ok := dst[key]; ok {

			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)

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

func mapify(i interface{}) (map[string]interface{}, bool) {

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

func parseMapArray(data interface{}) ([]map[string]interface{}, bool) {
	value := reflect.ValueOf(data)

	var arr []map[string]interface{}
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice, reflect.Array:
		if value.Len() == 0 {
			return nil, false
		}
		for i := 0; i < value.Len(); i++ {
			mapify, isMap := mapify(value.Index(i).Interface())
			if !isMap {
				return nil, false
			}
			arr = append(arr, mapify)
		}
	}
	return arr, true
}
