package dict

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	// PathSeparator is the character used to separate the elements
	// of the keypath.
	//
	// For example, `location.address.city`
	PathSeparator string = "."

	Strict bool = false
)

// Map is a map[string]interface{} with additional helpful functionality.
//
// You can use Map functionality on any map[string]interface{} using the following
// format:
//
//     data := map[string]interface{}{"name": "Stew"}
//     dict.Map(data).Get("name")
//     // returns "Stew"
type Map map[string]interface{}

func (m Map) String() string {
	var b []byte
	var err error
	if b, err = json.Marshal(&m); err != nil {
		return fmt.Sprintf("%T", m)
	}
	return string(b)
}

// NewMap creates a new map.  You may also use the M shortcut method.
//
// The arguments follow a key, value pattern.
//
// Panics
//
// Panics if any key arugment is non-string or if there are an odd number of arguments.
//
// Example
//
// To easily create Maps:
//
//     m := objects.M("name", "Mat", "age", 29, "subobj", objects.M("active", true))
//
//     // creates a Map equivalent to
//     m := map[string]interface{}{"name": "Mat", "age": 29, "subobj": map[string]interface{}{"active": true}}
func NewMap(keyAndValuePairs ...interface{}) Map {

	newMap := make(Map)
	keyAndValuePairsLen := len(keyAndValuePairs)

	if keyAndValuePairsLen%2 != 0 {
		panic("NewMap must have an even number of arguments following the 'key, value' pattern.")
	}

	for i := 0; i < keyAndValuePairsLen; i = i + 2 {

		key := keyAndValuePairs[i]
		value := keyAndValuePairs[i+1]

		// make sure the key is a string
		keyString, keyStringOK := key.(string)
		if !keyStringOK {
			panic(fmt.Sprintf("NewMap must follow 'string, interface{}' pattern.  %s is not a valid key.", keyString))
		}

		newMap[keyString] = value

	}

	return newMap
}

// M is a shortcut method for NewMap.
func M(keyAndValuePairs ...interface{}) Map {
	return NewMap(keyAndValuePairs...)
}

// NewMapFromJSON creates a new map from a JSON string representation
func NewMapFromJSON(data []byte) (Map, error) {

	var unmarshalled map[string]interface{}

	err := json.Unmarshal(data, &unmarshalled)

	if err != nil {
		return nil, errors.New("Map: JSON decode failed with: " + err.Error())
	}

	return Map(unmarshalled), nil

}

func get(m Map, key string) interface{} {
	if v, ok := m[key]; ok || Strict {
		return v
	} else if v, ok := m[strings.ToLower(key)]; ok {
		return v
	}
	return nil
}

// Get gets the value from the map.  Supports deep nesting of other maps.
func (d Map) Get(keypath string) interface{} {

	fi := strings.Index(keypath, PathSeparator)
	li := strings.LastIndex(keypath, PathSeparator)
	obj := d

	for true {
		if fi == -1 {
			return get(obj, keypath)
		}

		switch obj[keypath[:fi]].(type) {
		case Map:
			obj = obj[keypath[:fi]].(Map)
		case map[string]interface{}:
			obj = obj[keypath[:fi]].(map[string]interface{})
		default:
			f := get(obj, keypath[:fi])
			switch f.(type) {
			case Map:
				obj = f.(Map)
			case map[string]interface{}:
				obj = f.(map[string]interface{})
			}
			//case nil:
			//	obj = get(obj, keypath[:fi])
		}

		if fi == li {
			return get(obj, keypath[:fi])
			break

		}

		fi = strings.Index(keypath[fi:], ".")
	}

	return obj

}

// GetMap gets another Map from this one, or panics if the object is missing or not a Map.
func (d Map) GetMap(keypath string) Map {
	return d.Get(keypath).(Map)
}

// GetString gets a string value from the map at the given keypath, or panics if one
// is not available, or is of the wrong type.
func (d Map) GetString(keypath string) string {
	return d.Get(keypath).(string)
}

// GetWithDefault gets the value at the specified keypath, or returns the defaultValue if
// none could be found.
func (d Map) GetOrDefault(keypath string, defaultValue interface{}) interface{} {
	obj := d.Get(keypath)
	if obj == nil {
		return defaultValue
	}
	return obj
}

// GetStringOrDefault gets the string value at the specified keypath,
// or returns the defaultValue if none could be found.  Will panic if the
// object is there but of the wrong type.
func (d Map) GetStringOrDefault(keypath, defaultValue string) string {
	obj := d.Get(keypath)
	if obj == nil {
		return defaultValue
	}
	return obj.(string)
}

// GetStringOrEmpty gets the string value at the specified keypath or returns
// an empty string if none could be fo und. Will panic if the object is there
// but of the wrong type.
func (d Map) GetStringOrEmpty(keypath string) string {
	return d.GetStringOrDefault(keypath, "")
}

// Set sets a value in the map.  Supports dot syntax to set deep values.
//
// For example,
//
//     m.Set("name.first", "Mat")
//
// The above code sets the 'first' field on the 'name' object in the m Map.
//
// If objects are nil along the way, Set creates new Map objects as needed.
func (d Map) Set(keypath string, value interface{}) Map {

	var segs []string
	segs = strings.Split(keypath, PathSeparator)

	obj := d

	for fieldIndex, field := range segs {

		if fieldIndex == len(segs)-1 {
			obj[field] = value
		}

		if _, exists := obj[field]; !exists {
			obj[field] = make(Map)
			obj = obj[field].(Map)
		} else {
			switch obj[field].(type) {
			case Map:
				obj = obj[field].(Map)
			case map[string]interface{}:
				obj = Map(obj[field].(map[string]interface{}))
			}
		}

	}

	// chain
	return d
}

// Exclude returns a new Map with the keys in the specified []string
// excluded.
func (d Map) Omit(exclude []string) Map {

	excluded := make(Map)
	for k, v := range d {
		var shouldInclude bool = true
		for _, toExclude := range exclude {
			if k == toExclude {
				shouldInclude = false
				break
			}
		}
		if shouldInclude {
			excluded[k] = v
		}
	}

	return excluded
}

func (d Map) Pick(include []string) Map {
	out := make(Map)
	for _, v := range include {
		if vv := d.Get(v); vv != nil {
			out[v] = vv
		}
	}
	return out
}

// Copy creates a shallow copy of the Map.
func (d Map) Copy() Map {
	copied := make(Map)
	for k, v := range d {
		copied[k] = v
	}
	return copied
}

// Merge blends the specified map with this map and returns the current map.
//
// Keys that appear in both will be selected from the specified map.  The original map
// will be modified.
func (d Map) Extend(merge Map) Map {
	for k, v := range merge {
		d[k] = v
	}
	return d

}

// Has gets whether the Map has the specified field or not. Supports deep nesting of other maps.
//
// For example:
//     m := map[string]interface{}{"parent": map[string]interface{}{"childname": "Luke"}}
//     m.Has("parent.childname")
//     // return true
func (d Map) Has(path string) bool {
	return d.Get(path) != nil
}

// MSI is a shortcut method to get the current map as a
// normal map[string]interface{}.
func (d Map) ToMap() map[string]interface{} {
	return map[string]interface{}(d)
}

// JSON converts the map to a JSON string
/*func (d Map) MarshalJSON() ([]byte, error) {

	result, err := json.Marshal(&d)

	if err != nil {
		err = errors.New("Map: JSON encode failed with: " + err.Error())
	}

	return result, err

}

func (d *Map) UnmarshalJSON(bs []byte) error {
	return json.Unmarshal(bs, d)
}*/

// Transform builds a new map giving the transformer a chance
// to change the keys and values as it goes.
func (d Map) Transform(transformer func(key string, value interface{}) (string, interface{})) Map {
	m := make(Map)
	for k, v := range d {
		modifiedKey, modifiedVal := transformer(k, v)
		m[modifiedKey] = modifiedVal
	}
	return m
}

// TransformKeys builds a new map using the specified key mapping.
//
// Unspecified keys will be unaltered.
func (d Map) TransformKeys(mapping map[string]string) Map {
	return d.Transform(func(key string, value interface{}) (string, interface{}) {

		if newKey, ok := mapping[key]; ok {
			return newKey, value
		}

		return key, value
	})
}
