package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fr0stylo/chateh/client"
	"github.com/fr0stylo/chateh/protocol"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	conn := client.Connect(ctx, "0.0.0.0:9494")
	defer cancel()
	defer conn.Close()

	tui := initialModel()
	tui.WriteMessageC = conn.SendChannel()
	go pipeMessages(ctx, tui, conn.ReceiveChannel())
	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func pipeMessages(ctx context.Context, m *model, ch chan *protocol.ChatMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			m.WriteMessage(msg)
		}
	}
}

type (
	errMsg error
)

type model struct {
	viewport      viewport.Model
	messages      []string
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	err           error
	WriteMessageC chan string
}

func initialModel() *model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 30)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &model{
		textarea:      ta,
		messages:      []string{},
		WriteMessageC: make(chan string),
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:           nil,
	}
}

func (m *model) WriteMessage(msg *protocol.ChatMessage) {
	m.messages = append(m.messages, fmt.Sprintf("%s: %s", m.senderStyle.Render(msg.User), msg.Message))
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
}

func (m *model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.WriteMessageC <- m.textarea.Value()
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
