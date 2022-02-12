package commons

import "log"

func WalkMap(path string, m map[string]interface{}, fx func(path, currentKey string, v interface{}) interface{}) map[string]interface{} {
	for k, v := range m {
		if vv, ok := v.(map[string]interface{}); ok {
			if path == "" {
				path = k
			}
			path = path + "." + k

			m1 := WalkMap(path, vv, fx)
			m[k] = m1
			continue
		}
		v1 := fx(path+"."+k, k, v)
		m[k] = v1
	}
	return m
}

func ExampleWalkMap() {
	m := make(map[string]interface{})
	m["item"] = make(map[string]interface{})
	m["item"].(map[string]interface{})["data"] = "data1"
	m["item"].(map[string]interface{})["empresa"] = make(map[string]interface{})
	m["item"].(map[string]interface{})["empresa"].(map[string]interface{})["nombre"] = "intel"
	m["item"].(map[string]interface{})["persona"] = make(map[string]interface{})
	m["item"].(map[string]interface{})["persona"].(map[string]interface{})["nombre"] = "juan"
	m["item"].(map[string]interface{})["persona"].(map[string]interface{})["apellido"] = "perez"
	m["item"].(map[string]interface{})["persona"].(map[string]interface{})["age"] = 17
	m["item"].(map[string]interface{})["persona"].(map[string]interface{})["carrera"] = make(map[string]interface{})
	m["item"].(map[string]interface{})["persona"].(map[string]interface{})["carrera"].(map[string]interface{})["universidad"] = "utn"
	WalkMap("", m, func(path, field string, v interface{}) interface{} {
		log.Println("----"+path, v)
		return v //"xxxxx"
	})
	WalkMap("", m, func(path, field string, v interface{}) interface{} {
		log.Println("----"+path, v)
		return v
	})
}
