package tui

import (
	"fmt"
	"godex/pkg/mangadex"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

const (
	username = iota
	password
	clientId
	clientSecret
	downloadPath
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type TuiModel struct {
	inputs  []textinput.Model
	focused int
	err     error
}

func InitialModel() TuiModel {
	var inputs []textinput.Model = make([]textinput.Model, 5)
	inputs[username] = textinput.New()
	inputs[username].Focus()
	inputs[username].CharLimit = 100
	inputs[username].Width = 30
	inputs[username].Prompt = ""

	inputs[password] = textinput.New()
	inputs[password].CharLimit = 100
	inputs[password].Width = 100
	inputs[password].EchoMode = textinput.EchoPassword
	inputs[password].EchoCharacter = '•'
	inputs[password].Prompt = ""

	inputs[clientId] = textinput.New()
	inputs[clientId].CharLimit = 200
	inputs[clientId].Width = 200
	inputs[clientId].Prompt = ""

	inputs[clientSecret] = textinput.New()
	inputs[clientSecret].CharLimit = 200
	inputs[clientSecret].Width = 200
	inputs[clientSecret].EchoMode = textinput.EchoPassword
	inputs[clientSecret].EchoCharacter = '•'
	inputs[clientSecret].Prompt = ""

	inputs[downloadPath] = textinput.New()
	inputs[downloadPath].CharLimit = 400
	inputs[downloadPath].Width = 200
	inputs[downloadPath].Prompt = ""

	return TuiModel{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m TuiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				return m, tea.Quit
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m TuiModel) View() string {
	return fmt.Sprintf(
		` Please fill out the following fields:

 %s
 %s

 %s
 %s

 %s
 %s

 %s
 %s

 %s
 %s

 %s
`,
		inputStyle.Width(20).Render("Username"),
		m.inputs[username].View(),
		inputStyle.Width(20).Render("Password"),
		m.inputs[password].View(),
		inputStyle.Width(20).Render("ClientId"),
		m.inputs[clientId].View(),
		inputStyle.Width(20).Render("ClientSecret"),
		m.inputs[clientSecret].View(),
		inputStyle.Width(20).Render("Download Path"),
		m.inputs[downloadPath].View(),
		continueStyle.Render("Continue ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *TuiModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *TuiModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func (m TuiModel) GetConfig() *mangadex.EnvConfigs {
	return &mangadex.EnvConfigs{
		Username:     m.inputs[username].Value(),
		Password:     m.inputs[password].Value(),
		ClientId:     m.inputs[clientId].Value(),
		ClientSecret: m.inputs[clientSecret].Value(),
		DownloadPath: m.inputs[downloadPath].Value(),
	}
}
