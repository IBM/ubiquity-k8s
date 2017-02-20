package terminal

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	choices    []string
	defaultVal interface{}

	input string

	outputContains []string
	expected       interface{}
	err            error

	loop bool
}

func TestStringPrompt(t *testing.T) {
	assert := assert.New(t)

	promptTests := []TestData{
		{
			input:    "foo\n",
			expected: "foo",
		},
		{
			input: "\n",
			err:   ErrInputEmpty,
		},
		{
			defaultVal: "",
			input:      "\n",
			expected:   "",
		},
		{
			defaultVal: "bar",
			input:      "foo\n",
			expected:   "foo",
		},
		{
			input:          "\nfoo\n",
			outputContains: []string{"Enter a value"},
			expected:       "foo",
			loop:           true,
		},
	}

	for _, t := range promptTests {
		var answer string
		testPrompt(assert, t, &answer)
	}
}

func TestBoolPrompt(t *testing.T) {
	assert := assert.New(t)

	promptTests := []TestData{
		{
			input:    "yes\n",
			expected: true,
		},
		{
			input:    "Yes\n",
			expected: true,
		},
		{
			input:    "y\n",
			expected: true,
		},
		{
			input:    "Y\n",
			expected: true,
		},
		{
			input:    "no\n",
			expected: false,
		},
		{
			input:    "No\n",
			expected: false,
		},
		{
			input:    "n\n",
			expected: false,
		},
		{
			input:          "N\n",
			outputContains: []string{" [y/n]"},
			expected:       false,
		},
		{
			defaultVal:     true,
			input:          "\n",
			outputContains: []string{" [Y/n]"},
			expected:       true,
		},
		{
			defaultVal:     false,
			input:          "y\n",
			outputContains: []string{" [y/N]"},
			expected:       true,
		},
		{
			input: "NA\n",
			err:   ErrInputNotBool,
		},
	}

	for _, t := range promptTests {
		var answer bool
		testPrompt(assert, t, &answer)
	}
}

func TestIntPrompt(t *testing.T) {
	assert := assert.New(t)

	var intVar int
	var int8Var int8
	var int16Var int16
	var int32Var int32
	var int64Var int64

	ints := map[interface{}]interface{}{
		intVar:   100,
		int8Var:  int8(100),
		int16Var: int16(100),
		int32Var: int32(100),
		int64Var: int64(100),
	}

	for acutal, expected := range ints {
		testPrompt(
			assert,
			TestData{
				input:    "100\n",
				expected: expected,
			},
			&acutal)
	}

	intTests := []TestData{
		{
			input: "NA\n",
			err:   ErrInputNotNumber,
		},
		{
			input: "NA\n\n100\n",
			outputContains: []string{
				"Invalid input: input is not a number",
				"Enter a value",
			},
			expected: 100,
			loop:     true,
		},
	}

	for _, t := range intTests {
		var answer int
		testPrompt(assert, t, &answer)
	}
}

func TestFloatPrompt(t *testing.T) {
	assert := assert.New(t)

	var floatVar float64
	var float32Var float32

	floats := map[interface{}]interface{}{
		floatVar:   3.1415926,
		float32Var: float32(3.1415926),
	}

	for acutal, expected := range floats {
		testPrompt(
			assert,
			TestData{
				input:    "3.1415926\n",
				expected: expected,
			},
			&acutal)
	}

	var answer float64
	testPrompt(
		assert,
		TestData{
			input: "NA\n",
			err:   ErrInputNotFloatNumber,
		},
		&answer)
}

func TestSinglePrompt_ValidateFunc(t *testing.T) {
	assert := assert.New(t)

	var day string
	p := NewPrompt("input week day", &PromptOptions{
		ValidateFunc: func(input string) error {
			switch strings.ToUpper(input) {
			case "SUN":
			case "MON":
			case "TUE":
			case "WED":
			case "THU":
			case "FRI":
			case "SAT":
			default:
				return errors.New(input + " is not a valid week day")
			}
			return nil
		},
	})

	in := strings.NewReader("wedd\nwed\n")
	out := new(bytes.Buffer)
	p.Reader = in
	p.Writer = out
	err := p.Resolve(&day)

	assert.NoError(err)
	assert.Contains(out.String(), "Invalid input: wedd is not a valid week day")
	assert.Equal("wed", day)
}

func TestReadpasswordNonTTY(t *testing.T) {
	assert := assert.New(t)

	in := strings.NewReader("password\n")
	out := new(bytes.Buffer)

	p := NewPrompt("Input password", &PromptOptions{HideInput: true})
	p.Reader = in
	p.Writer = out

	var passwd string
	err := p.Resolve(&passwd)

	assert.NoError(err)
	assert.Equal("password", passwd)
	assert.NotContains("password", out.String())
}

func TestChoicesPrompt2(t *testing.T) {
	assert := assert.New(t)

	choices := []string{"foo", "bar"}
	choicesPromptTests := []TestData{
		{
			choices: choices,
			input:   "1\n",
			outputContains: []string{
				"1. foo",
				"2. bar",
				"Enter a number",
			},
			expected: "foo",
		},
		{
			choices:    choices,
			defaultVal: "bar",
			input:      "\n",
			outputContains: []string{
				"Enter a number (2)",
			},
			expected: "bar",
		},
		{
			choices:    choices,
			defaultVal: "jar",
			input:      "\n",
			outputContains: []string{
				"Enter a number ()",
			},
			expected: "jar",
		},
		{
			choices: choices,
			input:   "\n",
			err:     ErrInputEmpty,
		},
		{
			choices: choices,
			input:   "NA\n3\n2\n",
			outputContains: []string{
				"Invalid selection: input is not a number",
				"Invalid selection: input is out of range",
			},
			expected: "bar",
			loop:     true,
		},
	}

	for _, t := range choicesPromptTests {
		var answer string
		testPrompt(assert, t, &answer)
	}
}

func testPrompt(assert *assert.Assertions, d TestData, dest interface{}) {
	e := reflect.ValueOf(dest).Elem()
	if d.defaultVal != nil {
		e.Set(reflect.ValueOf(d.defaultVal))
	}

	var p *Prompt
	var msg string
	options := &PromptOptions{NoLoop: !d.loop, Required: (d.defaultVal == nil)}
	if len(d.choices) == 0 {
		p = NewPrompt("input something", options)
		msg = fmt.Sprintf("Prompt(%s): input: %q, default: %v", e.Type(), d.input, d.defaultVal)
	} else {
		p = NewChoicesPrompt("select an item", d.choices, options)
		msg = fmt.Sprintf("choices prompt[%s]: input: %q, default: %v", strings.Join(d.choices, ","), d.input, d.defaultVal)
	}

	in := strings.NewReader(d.input)
	out := new(bytes.Buffer)

	p.Reader = in
	p.Writer = out
	err := p.Resolve(dest)

	for _, s := range d.outputContains {
		assert.Contains(out.String(), s, msg)
	}

	if d.err != nil {
		assert.Equal(d.err, err, msg)
		return
	}
	assert.NoError(err, msg)
	assert.Equal(d.expected, e.Interface(), msg)
}
