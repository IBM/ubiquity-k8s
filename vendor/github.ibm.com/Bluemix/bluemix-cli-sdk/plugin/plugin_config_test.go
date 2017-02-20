package plugin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	config_string = `
{
	"name": "joe",
	"age": 30,
	"count": "5",
	"quantity": 200.58,
	"bool": true,
	"bools": "true",
	"foos": ["foo1", "foo2"],
	"nums": [1, 2],
	"strmapstr": {
		"foo1": "bar1",
		"foo2": "bar2"
	}
}`
)

func TestPluginConfig_FileNotExist(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "plugin_config_test")
	assert.NoError(err)

	assert.NotPanics(func() {
		NewPluginConfig(filepath.Join(tmpDir, "whatever"))
	})
}

func TestPluginConfigExists(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	assert.True(config.Exists("name"))
	assert.False(config.Exists("not-exist-key"))
}

func TestPluginConfigGet(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	testData := []struct {
		Key           string
		expectedValue interface{}
	}{
		{"name", "joe"},
		{"bool", true},
		{"bools", "true"},
		{"quantity", 200.58},
		{"age", 30.0}, //return float64
		{"foos", []interface{}{"foo1", "foo2"}},
		{"strmapstr", map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}},
		{"non-existing-key", nil},
	}

	for _, d := range testData {
		actualValue := config.Get(d.Key)
		assert.Equal(d.expectedValue, actualValue, d.Key)
	}
}

func TestPluginConfigGetString(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	testData := []struct {
		Key           string
		defaultValue  string
		expectedValue string
		expectedError bool
	}{
		{"name", "", "joe", false},
		{"bool", "", "true", false},
		{"age", "", "30", false},
		{"quantity", "", "200.58", false},
		{"non-existing-key", "", "", false},
		{"non-existing-key", "defaultS", "defaultS", false},
		{"foos", "", "", true},
	}

	for _, d := range testData {
		actualValue, err := config.GetStringWithDefault(d.Key, d.defaultValue)

		assert.Equal(d.expectedValue, actualValue, d.Key)
		if d.expectedError {
			assert.Error(err, d.Key)
		} else {
			assert.NoError(err, d.Key)
		}
	}
}

func TestPluginConfigGetBool(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	testData := []struct {
		Key           string
		defaultValue  bool
		expectedValue bool
		expectedError bool
	}{

		{"bool", false, true, false},
		{"bools", false, true, false},
		{"name", true, true, true},
	}

	for _, d := range testData {
		actualValue, err := config.GetBoolWithDefault(d.Key, d.defaultValue)

		assert.Equal(d.expectedValue, actualValue, d.Key)
		if d.expectedError {
			assert.Error(err, d.Key)
		} else {
			assert.NoError(err, d.Key)
		}
	}
}

func TestPluginConfigGetInt(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	testData := []struct {
		Key           string
		defaultValue  int
		expectedValue int
		expectedError bool
	}{

		{"age", 0, 30, false},
		{"count", 10, 5, false},
		{"quantity", 0, 200, false},
		{"name", 0, 0, true},
		{"non-existing-key", 0, 0, false},
		{"non-existing-key", 10, 10, false},
	}

	for _, d := range testData {
		actualValue, err := config.GetIntWithDefault(d.Key, d.defaultValue)

		assert.Equal(d.expectedValue, actualValue, d.Key)
		if d.expectedError {
			assert.Error(err, d.Key)
		} else {
			assert.NoError(err, d.Key)
		}
	}
}

func TestPluginConfigGetFloat(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	testData := []struct {
		Key           string
		defaultValue  float64
		expectedValue float64
		expectedError bool
	}{

		{"quantity", 0, 200.58, false},
		{"age", 0, 30.0, false},
		{"non-existing-key", 10.0, 10.0, false},
		{"name", 0.0, 0.0, true},
	}

	for _, d := range testData {
		actualValue, err := config.GetFloatWithDefault(d.Key, d.defaultValue)
		assert.Equal(d.expectedValue, actualValue, d.Key)
		if d.expectedError {
			assert.Error(err, d.Key)
		} else {
			assert.NoError(err, d.Key)
		}
	}
}

func TestPluginConfigGetIntSlice(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	v, err := config.GetIntSlice("nums")
	assert.NoError(err)
	assert.Equal([]int{1, 2}, v)
}

func TestPluginConfigGetStringSlice(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	v, err := config.GetStringSlice("foos")
	assert.NoError(err)
	assert.Equal([]string{"foo1", "foo2"}, v)
}

func TestPluginConfigGetFloatSlice(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	v, err := config.GetFloatSlice("nums")
	assert.NoError(err)
	assert.Equal([]float64{1.0, 2.0}, v)
}

func TestPluginConfigGetStringMap(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	v, err := config.GetStringMap("strmapstr")
	assert.NoError(err)
	assert.Equal(map[string]interface{}{"foo1": "bar1", "foo2": "bar2"}, v)
}

func TestPluginConfigGetStringMapString(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	v, err := config.GetStringMapString("strmapstr")
	assert.NoError(err)
	assert.Equal(map[string]string{"foo1": "bar1", "foo2": "bar2"}, v)
}

func TestPluginConfigErase(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	assert.NotNil(config.Get("name"))
	err := config.Erase("name")
	assert.NoError(err)
	assert.Equal(nil, config.Get("name"))
}

func TestPluginConfigErase_NotExisingKey(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	err := config.Erase("non-existing-key")
	assert.NoError(err)
}

func TestPluginConfigSet(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	err := config.Set("name", "Tom")
	assert.NoError(err)
	assert.Equal("Tom", config.Get("name"))
}

func TestPluginConfigSet_AddNewKey(t *testing.T) {
	assert := assert.New(t)

	path := prepareConfigFile()
	defer os.RemoveAll(filepath.Dir(path))

	config := NewPluginConfig(path)

	err := config.Set("new", "something new")
	assert.NoError(err)
	assert.Equal("something new", config.Get("new"))
}

func prepareConfigFile() string {
	tmpDir, err := ioutil.TempDir("", "plugin_config_test")
	if err != nil {
		panic("Failed to create temp dir:" + err.Error())
	}

	configFile := filepath.Join(tmpDir, "testConfig")
	err = ioutil.WriteFile(configFile, []byte(config_string), 0644)
	if err != nil {
		panic(err)
	}

	return configFile
}
