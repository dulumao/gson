package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"strconv"
)

// returns the current implementation version
func Version() string {
	return "0.1.0"
}

type Gson struct {
	data interface{}
}

// NewGson returns a pointer to a new `Gson` object
// after unmarshaling `body` bytes
func NewGson(body []byte) (*Gson, error) {
	self := new(Gson)
	err := self.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return self, nil
}

// NewFromReader returns a *Gson by decoding from an io.Reader
func NewFromReader(r io.Reader) (*Gson, error) {
	self := new(Gson)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	err := dec.Decode(&self.data)
	return self, err
}

// New returns a pointer to a new, empty `Gson` object
func New() *Gson {
	return &Gson{
		data: make(map[string]interface{}),
	}
}

// Interface returns the underlying data
func (self *Gson) Interface() interface{} {
	return self.data
}

// Encode returns its marshaled data as `[]byte`
func (self *Gson) Encode() ([]byte, error) {
	return self.MarshalJSON()
}

// EncodePretty returns its marshaled data as `[]byte` with indentation
func (self *Gson) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&self.data, "", "  ")
}

// Implements the json.Marshaler interface.
func (self *Gson) MarshalJSON() ([]byte, error) {
	return json.Marshal(&self.data)
}

// Set modifies `Gson` map by `key` and `value`
// Useful for changing single key/value in a `Gson` object easily.
func (self *Gson) Set(key string, val interface{}) {
	m, err := self.Map()
	if err != nil {
		return
	}
	m[key] = val
}

// SetPath modifies `Gson`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value
func (self *Gson) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		self.data = val
		return
	}

	// in order to insert our branch, we need map[string]interface{}
	if _, ok := (self.data).(map[string]interface{}); !ok {
		// have to replace with something suitable
		self.data = make(map[string]interface{})
	}
	curr := self.data.(map[string]interface{})

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(map[string]interface{}); !ok {
			// have to replace with something suitable
			n := make(map[string]interface{})
			curr[b] = n
		}

		curr = curr[b].(map[string]interface{})
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

// Del modifies `Gson` map by deleting `key` if it is present.
func (self *Gson) Del(key string) {
	m, err := self.Map()
	if err != nil {
		return
	}
	delete(m, key)
}

// Get returns a pointer to a new `Gson` object
// for `key` in its `map` representation
//
// useful for chaining operations (to traverse a nested JSON):
//    js.Get("top_level").Get("dict").Get("value").Int()
func (self *Gson) Get(key string) *Gson {
	m, err := self.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &Gson{val}
		}
	}
	return &Gson{nil}
}

// GetPath searches for the item as specified by the branch
// without the need to deep dive using Get()'s.
//
//   js.GetPath("top_level", "dict")
func (self *Gson) GetPath(branch ...string) *Gson {
	jin := self
	for _, p := range branch {
		jin = jin.Get(p)
	}
	return jin
}

// GetIndex returns a pointer to a new `Gson` object
// for `index` in its `array` representation
//
// this is the analog to Get when accessing elements of
// a json array instead of a json object:
//    js.Get("top_level").Get("array").GetIndex(1).Get("key").Int()
func (self *Gson) GetIndex(index int) *Gson {
	a, err := self.Array()
	if err == nil {
		if len(a) > index {
			return &Gson{a[index]}
		}
	}
	return &Gson{nil}
}

// CheckGet returns a pointer to a new `Gson` object and
// a `bool` identifying success or failure
//
// useful for chained operations when success is important:
//    if data, ok := selfs.Get("top_level").CheckGet("inner"); ok {
//        log.Println(data)
//    }
func (self *Gson) CheckGet(key string) (*Gson, bool) {
	m, err := self.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &Gson{val}, true
		}
	}
	return nil, false
}

// Map type asserts to `map`
func (self *Gson) Map() (map[string]interface{}, error) {
	if m, ok := (self.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// Array type asserts to an `array`
func (self *Gson) Array() ([]interface{}, error) {
	if a, ok := (self.data).([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// Bool type asserts to `bool`
func (self *Gson) Bool() (bool, error) {
	if s, ok := (self.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

// String type asserts to `string`
func (self *Gson) String() (string, error) {
	if s, ok := (self.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// Bytes type asserts to `[]byte`
func (self *Gson) Bytes() ([]byte, error) {
	if s, ok := (self.data).(string); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

// StringArray type asserts to an `array` of `string`
func (self *Gson) StringArray() ([]string, error) {
	arr, err := self.Array()
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if a == nil {
			retArr = append(retArr, "")
			continue
		}
		s, ok := a.(string)
		if !ok {
			return nil, err
		}
		retArr = append(retArr, s)
	}
	return retArr, nil
}

// MustArray guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, v := range js.Get("results").MustArray() {
//			fmt.Println(i, v)
//		}
func (self *Gson) MustArray(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustArray() received too many arguments %d", len(args))
	}

	a, err := self.Array()
	if err == nil {
		return a
	}

	return def
}

// MustMap guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//		for k, v := range js.Get("dictionary").MustMap() {
//			fmt.Println(k, v)
//		}
func (self *Gson) MustMap(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	a, err := self.Map()
	if err == nil {
		return a
	}

	return def
}

// MustString guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.Get("param1").MustString(), js.Get("optional_param").MustString("my_default"))
func (self *Gson) MustString(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustString() received too many arguments %d", len(args))
	}

	s, err := self.String()
	if err == nil {
		return s
	}

	return def
}

// MustStringArray guarantees the return of a `[]string` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, s := range js.Get("results").MustStringArray() {
//			fmt.Println(i, s)
//		}
func (self *Gson) MustStringArray(args ...[]string) []string {
	var def []string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustStringArray() received too many arguments %d", len(args))
	}

	a, err := self.StringArray()
	if err == nil {
		return a
	}

	return def
}

// MustInt guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.Get("param1").MustInt(), js.Get("optional_param").MustInt(5150))
func (self *Gson) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt() received too many arguments %d", len(args))
	}

	i, err := self.Int()
	if err == nil {
		return i
	}

	return def
}

// MustFloat64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Get("param1").MustFloat64(), js.Get("optional_param").MustFloat64(5.150))
func (self *Gson) MustFloat64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustFloat64() received too many arguments %d", len(args))
	}

	f, err := self.Float64()
	if err == nil {
		return f
	}

	return def
}

// MustBool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (self *Gson) MustBool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustBool() received too many arguments %d", len(args))
	}

	b, err := self.Bool()
	if err == nil {
		return b
	}

	return def
}

// MustInt64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Get("param1").MustInt64(), js.Get("optional_param").MustInt64(5150))
func (self *Gson) MustInt64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt64() received too many arguments %d", len(args))
	}

	i, err := self.Int64()
	if err == nil {
		return i
	}

	return def
}

// MustUInt64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Get("param1").MustUint64(), js.Get("optional_param").MustUint64(5150))
func (self *Gson) MustUint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustUint64() received too many arguments %d", len(args))
	}

	i, err := self.Uint64()
	if err == nil {
		return i
	}

	return def
}

// Implements the json.Unmarshaler interface.
func (self *Gson) UnmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(p))
	dec.UseNumber()
	return dec.Decode(&self.data)
}

// Float64 coerces into a float64
func (self *Gson) Float64() (float64, error) {
	switch self.data.(type) {
	case json.Number:
		return self.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(self.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(self.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(self.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int coerces into an int
func (self *Gson) Int() (int, error) {
	switch self.data.(type) {
	case json.Number:
		i, err := self.data.(json.Number).Int64()
		return int(i), err
	case float32, float64:
		return int(reflect.ValueOf(self.data).Float()), nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(self.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(self.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int64 coerces into an int64
func (self *Gson) Int64() (int64, error) {
	switch self.data.(type) {
	case json.Number:
		return self.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(self.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(self.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(self.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Uint64 coerces into an uint64
func (self *Gson) Uint64() (uint64, error) {
	switch self.data.(type) {
	case json.Number:
		return strconv.ParseUint(self.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(self.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(self.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(self.data).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}
