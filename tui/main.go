// checkpoint-view is an interactive terminal viewer for checkpoint-diagram logs.
// It parses a .claude/checkpoints/<date>.md file, renders each section's Mermaid
// flowchart to ASCII via mermaid-ascii, and shows it in a scrollable, pannable,
// color-coded Bubble Tea UI with a checkpoint sidebar.
//
// Usage: checkpoint-view [path/to/checkpoint.md]
// With no argument it opens the newest file in ./.claude/checkpoints/.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const sidebarW = 22

type section struct {
	title   string
	mermaid string
	ascii   []string
}

type model struct {
	sections      []section
	idx           int
	xoff, yoff    int
	width, height int
	status        string
}

func (m model) viewDims() (int, int) {
	vw := m.width - sidebarW - 2
	vh := m.height - 2
	if vw < 1 {
		vw = 1
	}
	if vh < 1 {
		vh = 1
	}
	return vw, vh
}

func (m model) current() section {
	if len(m.sections) == 0 {
		return section{}
	}
	return m.sections[m.idx]
}

func (m *model) clamp() {
	vw, vh := m.viewDims()
	lines := m.current().ascii
	if maxY := len(lines) - vh; m.yoff > maxY {
		m.yoff = maxY
	}
	if m.yoff < 0 {
		m.yoff = 0
	}
	maxLen := 0
	for _, l := range lines {
		if n := len([]rune(l)); n > maxLen {
			maxLen = n
		}
	}
	if maxX := maxLen - vw; m.xoff > maxX {
		m.xoff = maxX
	}
	if m.xoff < 0 {
		m.xoff = 0
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			m.yoff--
		case "down", "j":
			m.yoff++
		case "left", "h":
			m.xoff -= 4
		case "right", "l":
			m.xoff += 4
		case "pgup", "b":
			m.yoff -= 10
		case "pgdown", " ":
			m.yoff += 10
		case "g", "home":
			m.xoff, m.yoff = 0, 0
		case "n", "tab":
			if m.idx < len(m.sections)-1 {
				m.idx++
				m.xoff, m.yoff = 0, 0
			}
		case "p", "shift+tab":
			if m.idx > 0 {
				m.idx--
				m.xoff, m.yoff = 0, 0
			}
		}
	}
	m.clamp()
	return m, nil
}

var (
	stepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	decisionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	deferStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	borderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sidebarStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	footerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// styleFor colors a line by node role: pure box/arrow lines are dim, decision
// boxes (with "?") are amber, deferred work is gray, everything else is cyan.
func styleFor(full string) lipgloss.Style {
	if strings.IndexFunc(full, unicode.IsLetter) < 0 {
		return borderStyle
	}
	low := strings.ToLower(full)
	switch {
	case strings.Contains(low, "deferred"):
		return deferStyle
	case strings.Contains(full, "?"):
		return decisionStyle
	default:
		return stepStyle
	}
}

func (m model) View() string {
	if len(m.sections) == 0 {
		return "no checkpoint sections found\n"
	}
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}
	vw, vh := m.viewDims()

	var sb strings.Builder
	for i, s := range m.sections {
		t := s.title
		if len([]rune(t)) > sidebarW-2 {
			t = string([]rune(t)[:sidebarW-2])
		}
		if i == m.idx {
			sb.WriteString(selStyle.Render("▸ " + t))
		} else {
			sb.WriteString(sidebarStyle.Render("  " + t))
		}
		sb.WriteString("\n")
	}
	sidebar := lipgloss.NewStyle().
		Width(sidebarW).
		Height(vh).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(sb.String())

	lines := m.current().ascii
	var mv strings.Builder
	for r := 0; r < vh; r++ {
		li := m.yoff + r
		if li >= 0 && li < len(lines) {
			runes := []rune(lines[li])
			seg := ""
			if m.xoff < len(runes) {
				end := m.xoff + vw
				if end > len(runes) {
					end = len(runes)
				}
				seg = string(runes[m.xoff:end])
			}
			mv.WriteString(styleFor(lines[li]).Render(seg))
		}
		if r < vh-1 {
			mv.WriteString("\n")
		}
	}
	main := lipgloss.NewStyle().Width(vw).Height(vh).Render(mv.String())

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main)
	pos := fmt.Sprintf("[%d/%d]", m.idx+1, len(m.sections))
	help := footerStyle.Render(fmt.Sprintf(" %s %s  ·  ↑↓ scroll · ←→ pan · n/p checkpoint · g reset · q quit", m.status, pos))
	return body + "\n" + help
}

func mermaidBin() string {
	if p, err := exec.LookPath("mermaid-ascii"); err == nil {
		return p
	}
	var cands []string
	if out, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
		cands = append(cands, filepath.Join(strings.TrimSpace(string(out)), "bin", "mermaid-ascii"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		cands = append(cands, filepath.Join(home, "go", "bin", "mermaid-ascii"))
	}
	for _, c := range cands {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	return ""
}

func renderMermaid(bin, src string) []string {
	if bin == "" {
		return []string{"mermaid-ascii not found.", "install: go install github.com/AlexanderGrooff/mermaid-ascii@latest"}
	}
	cmd := exec.Command(bin, "-y", "1", "-x", "3", "-f", "-")
	cmd.Stdin = strings.NewReader(src)
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\n")
	if err != nil {
		return append([]string{"render error:"}, strings.Split(text, "\n")...)
	}
	if text == "" {
		return []string{"(empty render)"}
	}
	return strings.Split(text, "\n")
}

func parseFile(path string) ([]section, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var secs []section
	inBlock := false
	var block strings.Builder

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "## "):
			secs = append(secs, section{title: strings.TrimSpace(line[3:])})
			inBlock = false
			block.Reset()
		case trimmed == "```mermaid":
			inBlock = true
			block.Reset()
		case inBlock && trimmed == "```":
			inBlock = false
			if len(secs) > 0 {
				secs[len(secs)-1].mermaid = block.String()
			}
		case inBlock:
			block.WriteString(line)
			block.WriteByte('\n')
		}
	}
	return secs, sc.Err()
}

func newestCheckpoint(arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	dir := filepath.Join(".claude", "checkpoints")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("no file argument and %s not found", dir)
	}
	var mds []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mds = append(mds, e.Name())
		}
	}
	if len(mds) == 0 {
		return "", fmt.Errorf("no .md checkpoints in %s", dir)
	}
	sort.Strings(mds)
	return filepath.Join(dir, mds[len(mds)-1]), nil
}

func main() {
	arg := ""
	dump := false
	for _, a := range os.Args[1:] {
		if a == "--dump" {
			dump = true
			continue
		}
		arg = a
	}
	path, err := newestCheckpoint(arg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	secs, err := parseFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(secs) == 0 {
		fmt.Fprintf(os.Stderr, "no checkpoint sections in %s\n", path)
		os.Exit(1)
	}
	bin := mermaidBin()
	for i := range secs {
		if strings.TrimSpace(secs[i].mermaid) != "" {
			secs[i].ascii = renderMermaid(bin, secs[i].mermaid)
		} else {
			secs[i].ascii = []string{"(no diagram in this section)"}
		}
	}
	if dump {
		for _, s := range secs {
			fmt.Println("## " + s.title)
			for _, l := range s.ascii {
				fmt.Println(l)
			}
			fmt.Println()
		}
		return
	}
	m := model{sections: secs, status: filepath.Base(path)}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
