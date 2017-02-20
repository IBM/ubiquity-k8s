package terminal

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	term "github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/terminal"
)

// TODO: to be deprecated
type FakeUI struct {
	Inputs      bytes.Buffer
	Outputs     bytes.Buffer
	WarnOutputs bytes.Buffer
	Prompts     []string
}

func NewFakeUI() *FakeUI {
	return &FakeUI{}
}

func (ui *FakeUI) Say(template string, args ...interface{}) {
	message := fmt.Sprintf(template, args...)
	fmt.Fprintln(&ui.Outputs, message)
}

func (ui *FakeUI) Ok() {
	ui.Say("OK")
}

func (ui *FakeUI) Failed(template string, args ...interface{}) {
	message := fmt.Sprintf(template, args...)
	ui.Say("FAILED")
	ui.Say(message)
}

func (ui *FakeUI) Warn(template string, args ...interface{}) {
	message := fmt.Sprintf(template, args...)
	fmt.Fprintln(&ui.WarnOutputs, message)
}

func (ui *FakeUI) Prompt(message string, options *term.PromptOptions) *term.Prompt {
	p := term.NewPrompt(message, options)
	p.Reader = &ui.Inputs
	p.Writer = &ui.Outputs
	return p
}

func (ui *FakeUI) ChoicesPrompt(message string, choices []string, options *term.PromptOptions) *term.Prompt {
	p := term.NewChoicesPrompt(message, choices, options)
	p.Reader = &ui.Inputs
	p.Writer = &ui.Outputs
	return p
}

func (ui *FakeUI) Ask(template string, args ...interface{}) string {
	message := fmt.Sprintf(template, args...)

	var answer string
	err := ui.Prompt(message, &term.PromptOptions{HideDefault: true, NoLoop: true}).Resolve(&answer)
	if err == term.ErrInputEmpty {
		panic("No input provided to Fake UI for prompt: " + message)
	}

	return answer
}

func (ui *FakeUI) AskForPassword(template string, args ...interface{}) string {
	message := fmt.Sprintf(template, args...)

	var passwd string
	err := ui.Prompt(message, &term.PromptOptions{HideInput: true, NoLoop: true}).Resolve(&passwd)
	if err == term.ErrInputEmpty {
		panic("No input provided to Fake UI for password prompt: " + message)
	}

	return passwd
}

func (ui *FakeUI) Confirm(template string, args ...interface{}) bool {
	var yn bool
	message := fmt.Sprintf(template, args...)
	ui.Prompt(message, &term.PromptOptions{HideDefault: true, NoLoop: true}).Resolve(&yn)
	return yn
}

func (ui *FakeUI) ConfirmWithDefault(defaultBool bool, template string, args ...interface{}) bool {
	var yn = defaultBool
	message := fmt.Sprintf(template, args...)
	ui.Prompt(message, &term.PromptOptions{HideDefault: true, NoLoop: true}).Resolve(&yn)
	return yn
}

func (ui *FakeUI) SelectOne(choices []string, template string, args ...interface{}) int {
	var selected string
	message := fmt.Sprintf(template, args...)
	err := ui.ChoicesPrompt(message, choices, &term.PromptOptions{HideDefault: true}).Resolve(&selected)

	if err != nil {
		return -1
	}

	for i, c := range choices {
		if selected == c {
			return i
		}
	}

	return -1
}

func (ui *FakeUI) Table(headers []string) term.Table {
	return term.NewTable(&ui.Outputs, headers)
}

func (ui *FakeUI) ContainsOutput(text string) bool {
	return strings.Contains(ui.Outputs.String(), text)
}

func (ui *FakeUI) ContainsWarn(text string) bool {
	return strings.Contains(ui.WarnOutputs.String(), text)
}

func (ui *FakeUI) Writer() io.Writer {
	return &ui.Outputs
}
