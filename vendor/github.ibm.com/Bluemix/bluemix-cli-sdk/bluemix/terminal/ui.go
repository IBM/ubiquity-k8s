package terminal

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-colorable"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/trace"
)

type UI interface {
	Say(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Failed(format string, args ...interface{})
	Ok()

	Prompt(message string, options *PromptOptions) *Prompt
	ChoicesPrompt(message string, choices []string, options *PromptOptions) *Prompt

	// Deprecated: use Prompt() instead
	Ask(format string, args ...interface{}) (answer string)
	// Deprecated: use Prompt() instead
	AskForPassword(format string, args ...interface{}) (answer string)
	// Deprecated: use Prompt() instead
	Confirm(format string, args ...interface{}) bool
	// Deprecated: use Prompt() instead
	ConfirmWithDefault(defaultBool bool, format string, args ...interface{}) bool
	// Deprecated: use ChoicesPrompt() instead
	SelectOne(choices []string, format string, args ...interface{}) int

	Table(headers []string) Table
	Writer() io.Writer
}

type terminalUI struct {
	In  io.Reader
	Out io.Writer
}

func NewStdUI() UI {
	return NewUI(os.Stdin, colorable.NewColorableStdout())
}

func NewUI(in io.Reader, out io.Writer) UI {
	return &terminalUI{
		In:  in,
		Out: out,
	}
}

func (ui *terminalUI) Say(format string, args ...interface{}) {
	fmt.Fprintf(ui.Out, format+"\n", args...)
}

func (ui *terminalUI) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	ui.Say(WarningColor(message))
}

func (ui *terminalUI) Ok() {
	ui.Say(SuccessColor("OK"))
}

func (ui *terminalUI) Failed(format string, args ...interface{}) {
	ui.Say(FailureColor("FAILED"))
	ui.Say(format, args...)
	ui.Say("")

	trace.Logger.Print("FAILED")
	trace.Logger.Printf(format+"\n", args...)
}

func (ui *terminalUI) Prompt(message string, options *PromptOptions) *Prompt {
	p := NewPrompt(message, options)
	p.Reader = ui.In
	p.Writer = ui.Out
	return p
}

func (ui *terminalUI) ChoicesPrompt(message string, choices []string, options *PromptOptions) *Prompt {
	p := NewChoicesPrompt(message, choices, options)
	p.Reader = ui.In
	p.Writer = ui.Out
	return p
}

func (ui *terminalUI) Ask(format string, args ...interface{}) (answer string) {
	message := fmt.Sprintf(format, args...)
	ui.Prompt(message, &PromptOptions{HideDefault: true, NoLoop: true}).Resolve(&answer)
	return
}

func (ui *terminalUI) AskForPassword(format string, args ...interface{}) (passwd string) {
	message := fmt.Sprintf(format, args...)
	ui.Prompt(message, &PromptOptions{HideInput: true, HideDefault: true, NoLoop: true}).Resolve(&passwd)
	return
}

func (ui *terminalUI) Confirm(format string, args ...interface{}) (yn bool) {
	message := fmt.Sprintf(format, args...)
	ui.Prompt(message, &PromptOptions{HideDefault: true, NoLoop: true}).Resolve(&yn)
	return
}

func (ui *terminalUI) ConfirmWithDefault(defaultBool bool, format string, args ...interface{}) (yn bool) {
	yn = defaultBool
	message := fmt.Sprintf(format, args...)
	ui.Prompt(message, &PromptOptions{HideDefault: true, NoLoop: true}).Resolve(&yn)
	return
}

func (ui *terminalUI) SelectOne(choices []string, format string, args ...interface{}) int {
	var selected string
	message := fmt.Sprintf(format, args...)

	err := ui.ChoicesPrompt(message, choices, &PromptOptions{HideDefault: true}).Resolve(&selected)
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

func (ui *terminalUI) Table(headers []string) Table {
	return NewTable(ui.Out, headers)
}

func (ui *terminalUI) Writer() io.Writer {
	return ui.Out
}
