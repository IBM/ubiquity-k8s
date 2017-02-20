package plugin

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration"
)

// PluginConfig defines methods to access plug-in's private configuration stored in a JSON format.
type PluginConfig interface {
	// Get returns the value for the given key.
	// The return value is nil, float64, bool, string, []interface{} or map[string]interface.
	// If key not exists, return nil.
	Get(key string) interface{}

	// Get returns the value for the given key.
	// The return value is nil, float64, bool, string, []interface{} or map[string]interface.
	// If key not exists, return defaultVal.
	GetWithDefault(key string, defaultVal interface{}) interface{}

	// GetString returns string value for the given key.
	// If key not exists, return "".
	GetString(key string) (string, error)

	// GetStringWithDefault returns string value for the given key.
	// If key not exists, return defaultVal.
	GetStringWithDefault(key string, defaultVal string) (string, error)

	// GetBool returns boolean value for the given key.
	// If key not exists, return false.
	// If the value is a string, attempts to convert it to bool.
	GetBool(key string) (bool, error)

	// GetBoolWithDefault returns boolean value for the given key.
	// If key not exists, return defaultVal.
	// If the value is a string, attempts to convert it to bool.
	GetBoolWithDefault(key string, defaultVal bool) (bool, error)

	// GetInt returns int value for the given key.
	// If key not exists, return 0.
	// If the value is float or string, attempts to convert it to int.
	GetInt(key string) (int, error)

	// GetIntWithDefault returns int value for the given key.
	// If key not exists, return defaultVal.
	// If the value is float or string, attempts to convert it to int.
	GetIntWithDefault(key string, defaultVal int) (int, error)

	// GetFloat returns float64 value for the given key.
	// If key not exists, return 0.0.
	// If the value is int or string, attempts to convert it to float64.
	GetFloat(key string) (float64, error)

	// GetFloat returns float64 value for the given key.
	// If key not exists, return defaultVal
	// If the value is int or string, attempts to convert it to float64.
	GetFloatWithDefault(key string, defaultVal float64) (float64, error)

	// GetStringSlice return string slice for the given key.
	// If key not exists, return empty string slice.
	GetStringSlice(key string) ([]string, error)

	// GetIntSlice return string slice for the given key.
	// If key not exists, return empty int slice.
	GetIntSlice(key string) ([]int, error)

	// GetFloatSlice return string slice for the given key.
	// If key not exists, return empty float slice.
	GetFloatSlice(key string) ([]float64, error)

	// GetStringMap return map[string]interface{} for the given key.
	// If key not exists, return empty map.
	GetStringMap(key string) (map[string]interface{}, error)

	// GetStringMap return map[string]string for the given key.
	// If key not exists, return empty map.
	GetStringMapString(key string) (map[string]string, error)

	// Exists checks whether the key exists or not.
	Exists(key string) bool

	// Set sets the value for the given key.
	Set(string, interface{}) error

	// Erase delete the given key.
	Erase(key string) error
}

type TypeError struct {
	Key          string
	ExpectedType string
}

func NewTypeError(key string, expectedType string) *TypeError {
	return &TypeError{
		Key:          key,
		ExpectedType: expectedType,
	}
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("plugin config: %s - unable to convert value to type %s.", e.Key, e.ExpectedType)
}

type pluginConfigImpl struct {
	initOnce  *sync.Once
	lock      sync.RWMutex
	data      map[string]interface{}
	persistor configuration.Persistor
}

func NewPluginConfig(path string) PluginConfig {
	return &pluginConfigImpl{
		initOnce:  new(sync.Once),
		data:      make(map[string]interface{}),
		persistor: configuration.NewDiskPersistor(path),
	}
}

func (c *pluginConfigImpl) init() {
	c.initOnce.Do(func() {
		err := c.persistor.Load(&c.data)
		if err != nil && os.IsNotExist(err) {
			err = c.persistor.Save(c.data)
		}
		if err != nil {
			panic(err)
		}
	})
}

func (c *pluginConfigImpl) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.init()

	return c.data[key]
}

func (c *pluginConfigImpl) GetWithDefault(key string, defaultVal interface{}) interface{} {
	v := c.Get(key)
	if v == nil {
		return defaultVal
	}
	return v
}

func (c *pluginConfigImpl) GetString(key string) (string, error) {
	return c.GetStringWithDefault(key, "")
}

func (c *pluginConfigImpl) GetStringWithDefault(key string, defaultVal string) (string, error) {
	v := c.Get(key)
	if v == nil {
		return defaultVal, nil
	}

	ret, ok := toString(v)
	if !ok {
		return defaultVal, NewTypeError(key, "string")
	}
	return ret, nil
}

func toString(v interface{}) (string, bool) {
	switch v.(type) {
	case bool:
		return strconv.FormatBool(v.(bool)), true
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64), true
	case string:
		return v.(string), true
	case nil:
		return "", true
	}
	return "", false
}

func (c *pluginConfigImpl) GetBool(key string) (bool, error) {
	return c.GetBoolWithDefault(key, false)
}

func (c *pluginConfigImpl) GetBoolWithDefault(key string, defaultVal bool) (bool, error) {
	v := c.Get(key)
	if v == nil {
		return defaultVal, nil
	}

	ret, ok := toBool(v)
	if !ok {
		return defaultVal, NewTypeError(key, "bool")
	}
	return ret, nil
}

func toBool(v interface{}) (bool, bool) {
	switch v.(type) {
	case bool:
		return v.(bool), true
	case string:
		b, err := strconv.ParseBool(v.(string))
		if err == nil {
			return b, true
		}
	}
	return false, false
}

func (c *pluginConfigImpl) GetInt(key string) (int, error) {
	return c.GetIntWithDefault(key, 0)
}

func (c *pluginConfigImpl) GetIntWithDefault(key string, defaultVal int) (int, error) {
	v := c.Get(key)
	if v == nil {
		return defaultVal, nil
	}

	ret, ok := toInt(v)
	if !ok {
		return defaultVal, NewTypeError(key, "int")
	}
	return ret, nil
}

func toInt(v interface{}) (int, bool) {
	switch v.(type) {
	case float64:
		return int(v.(float64)), true
	case string:
		i, err := strconv.ParseInt(v.(string), 0, 0)
		if err == nil {
			return int(i), true
		}
	}
	return 0, false
}

func (c *pluginConfigImpl) GetFloat(key string) (float64, error) {
	return c.GetFloatWithDefault(key, 0.00)
}

func (c *pluginConfigImpl) GetFloatWithDefault(key string, defaultVal float64) (float64, error) {
	v := c.Get(key)
	if v == nil {
		return defaultVal, nil
	}

	ret, ok := toFloat(v)
	if !ok {
		return defaultVal, NewTypeError(key, "float64")
	}
	return ret, nil
}

func toFloat(v interface{}) (float64, bool) {
	switch v.(type) {
	case float64:
		return v.(float64), true
	case string:
		f, err := strconv.ParseFloat(v.(string), 64)
		if err == nil {
			return f, true
		}
	}
	return 0.00, false
}

func (c *pluginConfigImpl) GetSlice(key string) ([]interface{}, error) {
	v := c.Get(key)
	if v == nil {
		return []interface{}{}, nil
	}

	_, ok := v.([]interface{})
	if !ok {
		return []interface{}{}, NewTypeError(key, "[]interface{}")
	}
	return v.([]interface{}), nil
}

func (c *pluginConfigImpl) GetStringSlice(key string) ([]string, error) {
	v := c.Get(key)
	if v == nil {
		return []string{}, nil
	}

	ret, ok := toStringSlice(v)
	if !ok {
		return []string{}, NewTypeError(key, "[]string")
	}
	return ret, nil
}

func toStringSlice(v interface{}) ([]string, bool) {
	s, ok := v.([]interface{})
	if !ok {
		return []string{}, false
	}

	ret := make([]string, len(s), len(s))
	for i := 0; i < len(s); i++ {
		ii, ok := toString(s[i])
		if !ok {
			return []string{}, false
		}
		ret[i] = ii
	}
	return ret, true
}

func (c *pluginConfigImpl) GetIntSlice(key string) ([]int, error) {
	v := c.Get(key)
	if v == nil {
		return []int{}, nil
	}

	ret, ok := toIntSlice(v)
	if !ok {
		return []int{}, NewTypeError(key, "[]int")
	}
	return ret, nil
}

func toIntSlice(v interface{}) ([]int, bool) {
	s, ok := v.([]interface{})
	if !ok {
		return []int{}, false
	}

	ret := make([]int, len(s), len(s))
	for i := 0; i < len(s); i++ {
		ii, ok := toInt(s[i])
		if !ok {
			return []int{}, false
		}
		ret[i] = ii
	}
	return ret, true
}

func (c *pluginConfigImpl) GetFloatSlice(key string) ([]float64, error) {
	v := c.Get(key)
	if v == nil {
		return []float64{}, nil
	}

	ret, ok := toFloatSlice(v)
	if !ok {
		return []float64{}, NewTypeError(key, "[]float64")
	}
	return ret, nil
}

func toFloatSlice(v interface{}) ([]float64, bool) {
	s, ok := v.([]interface{})
	if !ok {
		return []float64{}, false
	}

	ret := make([]float64, len(s), len(s))
	for i := 0; i < len(s); i++ {
		ii, ok := toFloat(s[i])
		if !ok {
			return []float64{}, false
		}
		ret[i] = ii
	}
	return ret, true
}

func (c *pluginConfigImpl) GetStringMap(key string) (map[string]interface{}, error) {
	v := c.Get(key)
	if v == nil {
		return map[string]interface{}{}, nil
	}

	_, ok := v.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, NewTypeError(key, "map[string]interface{}")
	}
	return v.(map[string]interface{}), nil
}

func (c *pluginConfigImpl) GetStringMapString(key string) (map[string]string, error) {
	v := c.Get(key)
	if v == nil {
		return map[string]string{}, nil
	}

	ret, ok := toMapStringMapString(v)
	if !ok {
		return map[string]string{}, NewTypeError(key, "map[string]string")
	}
	return ret, nil
}

func toMapStringMapString(v interface{}) (map[string]string, bool) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return map[string]string{}, false
	}

	ret := make(map[string]string)
	for k, v := range m {
		s, ok := toString(v)
		if !ok {
			return map[string]string{}, false
		}
		ret[k] = s
	}

	return ret, true
}

func (c *pluginConfigImpl) Exists(key string) bool {
	return c.Get(key) != nil
}

func (c *pluginConfigImpl) Set(key string, v interface{}) error {
	return c.write(func() {
		c.data[key] = v
	})
}

func (c *pluginConfigImpl) Erase(key string) error {
	return c.write(func() {
		delete(c.data, key)
	})
}

func (c *pluginConfigImpl) write(cb func()) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.init()

	cb()

	err := c.persistor.Save(c.data)
	if err != nil {
		err = fmt.Errorf("unable to save plugin config: %v", err)
	}
	return err
}
