package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-logr/logr"
	"github.com/metalkast/metalkast/cmd/kast/options"
	"github.com/muesli/termenv"
)

type model struct {
	logs              chan logEntry
	spinner           spinner.Model
	activeTopLevelLog *logEntry
	subLevelLogs      []logEntry
	subLevelLogsCap   int
	terminalWidth     int
	logFile           *os.File
}

type quit struct{}

var (
	// Color code to name reference: https://github.com/muesli/termenv#color-chart
	spinnerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	checkmarkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorLogStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	subLevelLogStyle = lipgloss.NewStyle().Faint(true)
	appStyle         = lipgloss.NewStyle()
	logLineStyle     = lipgloss.NewStyle()
)

// A command that waits for the activity on a channel.
func waitForActivity(sub chan logEntry) tea.Cmd {
	return func() tea.Msg {
		for s := range sub {
			return s
		}
		return quit{}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForActivity(m.logs),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logEntry:
		if m.logFile != nil {
			fileLogLine := msg.String()
			if msg.err != nil {
				fileLogLine = fmt.Sprintf("[ERROR]%s: %s", fileLogLine, msg.err)
			}
			m.logFile.WriteString(fileLogLine + "\n")
		}

		var permanentLogs []tea.Cmd
		if msg.level == 0 && msg.err == nil {
			m.subLevelLogs = []logEntry{}
			if m.activeTopLevelLog != nil {
				m.activeTopLevelLog.success = true
				renderLine := logLineStyle.Width(m.terminalWidth).Render(checkmarkStyle.Render("✓ ") + m.activeTopLevelLog.String())
				for _, l := range strings.Split(renderLine, "\n") {
					permanentLogs = append(permanentLogs, tea.Println(l))
				}
			}
			if msg.success {
				msg.success = true
				renderLine := logLineStyle.Width(m.terminalWidth).Render(checkmarkStyle.Render("✓ ") + msg.String())
				for _, l := range strings.Split(renderLine, "\n") {
					permanentLogs = append(permanentLogs, tea.Println(l))
				}
				m.activeTopLevelLog = nil
			} else {
				m.activeTopLevelLog = &msg
			}
		} else if m.activeTopLevelLog != nil {
			m.subLevelLogs = append(m.subLevelLogs, msg)
		} else {
			renderLine := msg.String()
			if msg.err != nil {
				renderLine = errorLogStyle.Render(msg.String())
				permanentLogs = append(permanentLogs, tea.Println(logLineStyle.Width(m.terminalWidth).Render(msg.String())))
			}
			permanentLogs = append(permanentLogs, tea.Println(logLineStyle.Width(m.terminalWidth).Render(renderLine)))
		}
		return m, tea.Batch(waitForActivity(m.logs), tea.Sequence(permanentLogs...))
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		return m, nil
	case quit:
		var commands []tea.Cmd
		if m.activeTopLevelLog != nil {
			m.activeTopLevelLog.success = true
			renderLine := logLineStyle.Width(m.terminalWidth).Render(checkmarkStyle.Render("✓ ") + m.activeTopLevelLog.String())
			for _, l := range strings.Split(renderLine, "\n") {
				commands = append(commands, tea.Println(l))
			}
		}
		printSubLevelLogs := m.activeTopLevelLog == nil
		for _, l := range m.subLevelLogs {
			if l.err != nil {
				printSubLevelLogs = true
				break
			}
		}
		if printSubLevelLogs {
			for _, l := range m.subLevelLogs {
				renderLine := l.String()
				if l.err != nil {
					renderLine = errorLogStyle.Render(l.String())
					commands = append(commands, tea.Println(logLineStyle.Width(m.terminalWidth).Render(l.String())))
				}
				commands = append(commands, tea.Println(logLineStyle.Width(m.terminalWidth).Render(renderLine)))
			}
		}
		m.activeTopLevelLog = nil
		m.subLevelLogs = nil
		commands = append(commands, tea.Quit)
		return m, tea.Sequence(commands...)
	default:
		return m, nil
	}
}

func (m model) View() string {
	builder := strings.Builder{}
	if m.activeTopLevelLog != nil {
		builder.WriteString(
			appStyle.Width(m.terminalWidth).Render(m.spinner.View()+m.activeTopLevelLog.String()) + "\n",
		)
	}

	subLevelLogDisplay := []logEntry{}
	subLevelLogStartIndex := 0
	for i := len(m.subLevelLogs) - 1; i >= 0; i-- {
		if m.subLevelLogs[i].refreshable {
			subLevelLogDisplay = append(subLevelLogDisplay, m.subLevelLogs[i])
			subLevelLogStartIndex = i + 1
			break
		}
	}
	if capEntriesIndexStart := len(m.subLevelLogs) - m.subLevelLogsCap + len(subLevelLogDisplay); capEntriesIndexStart > subLevelLogStartIndex {
		subLevelLogStartIndex = capEntriesIndexStart
	}
	subLevelLogDisplay = append(subLevelLogDisplay, m.subLevelLogs[subLevelLogStartIndex:]...)

	for _, logLine := range subLevelLogDisplay {
		line := logLine.String()
		if logLine.err != nil {
			line = errorLogStyle.Render(line)
		}
		builder.WriteString(subLevelLogStyle.Width(m.terminalWidth).Render(line) + "\n")
	}

	s := builder.String()
	return appStyle.Width(m.terminalWidth).Render(s)
}

type logKeyPair struct {
	key   interface{}
	value interface{}
}

type logEntry struct {
	msg         string
	keyPairs    []logKeyPair
	refreshable bool
	level       int
	success     bool
	err         error
}

func newLogEntry(level int, msg string, keysAndValues ...interface{}) logEntry {
	keysAndValues = keysAndValues[:len(keysAndValues)/2*2]
	res := logEntry{
		msg:   msg,
		level: level,
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		key := keysAndValues[i]
		value := keysAndValues[i+1]
		if key == "refreshable" {
			if refreshable, ok := value.(bool); ok {
				res.refreshable = refreshable
			}
			continue
		}
		if key == "success" {
			if success, ok := value.(bool); ok {
				res.success = success
			}
			continue
		}
		res.keyPairs = append(res.keyPairs, logKeyPair{
			key:   key,
			value: value,
		})
	}

	return res
}

func (e logEntry) keyPairsString() string {
	builder := strings.Builder{}
	for _, kp := range e.keyPairs {
		builder.Write([]byte(fmt.Sprintf("%v=%v ", kp.key, kp.value)))
	}
	return strings.TrimSpace(builder.String())
}

func (e logEntry) String() string {
	s := fmt.Sprintf("%s  %s\n", e.msg, e.keyPairsString())
	if e.err != nil {
		s = fmt.Sprintf("%s: %s  %s\n", e.msg, e.err, e.keyPairsString())
	}
	return strings.TrimSpace(s)
}

type TeaLogSink struct {
	logs        chan logEntry
	prefixes    []string
	teaShutdown chan error
	levelLimit  int
}

var _ logr.LogSink = &TeaLogSink{}

type LoggerOptions struct {
	OutputPath string
}

func NewLogger(opts LoggerOptions) (logr.Logger, error) {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	s.Style = spinnerStyle
	lipgloss.SetColorProfile(termenv.ANSI)

	var logFile *os.File
	if opts.OutputPath != "" {
		var err error
		logFile, err = tea.LogToFile(opts.OutputPath, "")
		if err != nil {
			return logr.Logger{}, err
		}
	}

	m := model{
		logs:            make(chan logEntry),
		spinner:         s,
		subLevelLogsCap: 10,
		// Arbitrary initlal terminal width
		terminalWidth: 80,
		logFile:       logFile,
	}

	l := &TeaLogSink{
		logs:        m.logs,
		teaShutdown: make(chan error),
		levelLimit:  2 + options.Verbosity,
	}

	p := tea.NewProgram(m, tea.WithoutCatchPanics(), tea.WithoutSignalHandler(), tea.WithInput(nil))
	go func() {
		_, err := p.Run()
		l.teaShutdown <- err
	}()

	return logr.New(l), nil
}

func (l *TeaLogSink) Close() error {
	close(l.logs)
	err := <-l.teaShutdown
	return err
}

// Enabled implements logr.LogSink.
func (l *TeaLogSink) Enabled(level int) bool {
	return l.levelLimit > level
}

// Error implements logr.LogSink.
func (l *TeaLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	if prefix := l.prefix(); prefix != "" {
		msg = prefix + " " + msg
	}
	entry := newLogEntry(0, msg, keysAndValues...)
	entry.err = err
	l.logs <- entry
}

// Info implements logr.LogSink.
func (l *TeaLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	if prefix := l.prefix(); prefix != "" {
		msg = prefix + " " + msg
	}
	entry := newLogEntry(level, msg, keysAndValues...)
	l.logs <- entry
}

// Init implements logr.LogSink.
func (*TeaLogSink) Init(info logr.RuntimeInfo) {
	// Not important
}

// WithName implements logr.LogSink.
func (l *TeaLogSink) WithName(name string) logr.LogSink {
	newLogger := &TeaLogSink{
		prefixes:   l.prefixes,
		logs:       l.logs,
		levelLimit: l.levelLimit,
	}
	newLogger.prefixes = append(newLogger.prefixes, name)
	return newLogger
}

// WithValues implements logr.LogSink.
func (l *TeaLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	newLogger := &TeaLogSink{
		prefixes:   l.prefixes,
		logs:       l.logs,
		levelLimit: l.levelLimit,
	}
	for i, key := range keysAndValues {
		if vIndex := i + 1; vIndex < len(keysAndValues) {
			newLogger.prefixes = append(newLogger.prefixes, fmt.Sprintf("%v=%v", key, keysAndValues[vIndex]))
		}
	}
	return newLogger
}

func (l *TeaLogSink) prefix() string {
	builder := strings.Builder{}
	for _, p := range l.prefixes {
		builder.WriteString(fmt.Sprintf("[%s]", p))
	}
	return builder.String()
}
