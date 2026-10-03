package ui

import (
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/sys/unix"

	"sudoku-tui/sudoku"
)

// Écrans de l'application.
type screen int

const (
	scrMenu screen = iota
	scrPlay
	scrWin
)

// snapshot permet un annuler/refaire fiable (état complet avant un coup).
type snapshot struct {
	board    sudoku.Board
	notes    [9][9]uint16
	mistakes int
}

// Model est l'état complet du programme Bubble Tea.
type Model struct {
	width, height int
	ready         bool

	screen     screen
	menuCursor int

	// Partie en cours.
	difficulty sudoku.Difficulty
	puzzle     sudoku.Board // indices de départ (immuables)
	board      sudoku.Board // état courant
	solution   sudoku.Board
	notes      [9][9]uint16 // masque de bits des notes (bit v = chiffre v)
	conflicts  [9][9]bool

	cursorR, cursorC int
	noteMode         bool
	showHelp         bool
	paused           bool
	celebrate        bool    // grille résolue : coloration « victoire »
	ch               int     // hauteur de case choisie par la vue (3 ou 1)
	cw               int     // largeur de case choisie par la vue
	aspect           float64 // rapport hauteur/largeur d'un caractère
	history          []snapshot

	elapsed  time.Duration
	mistakes int
	hints    int

	status      string
	statusStyle lipgloss.Style

	rng *rand.Rand
}

// New construit le modèle initial (écran menu).
func New() Model {
	return Model{
		screen:      scrMenu,
		statusStyle: styleMuted,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		aspect:      terminalAspect(),
	}
}

// terminalAspect déduit le rapport hauteur/largeur d'un caractère du terminal
// à partir de la taille de la fenêtre en pixels. C'est ce qui permet à la
// grille de paraître carrée quelle que soit la police employée. Si le terminal
// ne communique pas sa taille en pixels (cas de tmux, par exemple), on retombe
// sur 2.0, valeur courante pour une police à chasse fixe.
func terminalAspect() float64 {
	const fallback = 2.0
	f, err := os.Open("/dev/tty")
	if err != nil {
		return fallback
	}
	defer f.Close()
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 || ws.Xpixel == 0 || ws.Ypixel == 0 {
		return fallback
	}
	cellW := float64(ws.Xpixel) / float64(ws.Col)
	cellH := float64(ws.Ypixel) / float64(ws.Row)
	if cellW <= 0 {
		return fallback
	}
	a := cellH / cellW
	if a < 1.4 || a > 3.2 { // valeur invraisemblable : on ne s'y fie pas
		return fallback
	}
	return a
}

// Init démarre la boucle d'horloge (une seule, pour toute la session).
func (m Model) Init() tea.Cmd { return tickCmd() }

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update traite les messages Bubble Tea.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		return m, nil

	case tickMsg:
		if m.screen == scrPlay && !m.paused && !m.showHelp {
			m.elapsed += time.Second
		}
		return m, tickCmd()

	case tea.MouseMsg:
		m.handleMouse(msg)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// ---------------------------------------------------------------- clavier --

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Quitter, où qu'on soit.
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.showHelp {
		switch key {
		case "?", "esc", "q", "enter", " ":
			m.showHelp = false
		}
		return m, nil
	}

	switch m.screen {
	case scrMenu:
		return m.menuKey(key)
	case scrPlay:
		return m.playKey(key)
	case scrWin:
		return m.winKey(key)
	}
	return m, nil
}

func (m Model) menuKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < len(sudoku.Difficulties)-1 {
			m.menuCursor++
		}
	case "1", "2", "3", "4":
		m.menuCursor = int(key[0] - '1')
		return m.start(sudoku.Difficulties[m.menuCursor]), nil
	case "enter", " ":
		return m.start(sudoku.Difficulties[m.menuCursor]), nil
	}
	return m, nil
}

func (m Model) playKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc", "m":
		m.screen = scrMenu
		m.paused = false
		return m, nil
	case "?":
		m.showHelp = true
		return m, nil
	case "up", "k":
		m.move(-1, 0)
	case "down", "j":
		m.move(1, 0)
	case "left", "h":
		m.move(0, -1)
	case "right", "l":
		m.move(0, 1)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if m.paused {
			break
		}
		m.place(int(key[0] - '0'))
	case "0", "backspace", "delete":
		if !m.paused {
			m.erase()
		}
	case "n", " ":
		m.noteMode = !m.noteMode
		if m.noteMode {
			m.setStatus(styleAccent, "✎ Mode notes activé")
		} else {
			m.setStatus(styleMuted, "Mode notes désactivé")
		}
	case "u":
		m.undo()
	case "H":
		if !m.paused {
			m.hint()
		}
	case "r":
		m.restart()
	case "p":
		m.paused = !m.paused
		if m.paused {
			m.setStatus(styleWarn, "‖  En pause")
		} else {
			m.setStatus(styleMuted, "Reprise")
		}
	}
	return m, nil
}

func (m Model) winKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc", "m":
		m.screen = scrMenu
	case "enter", " ":
		m.menuCursor = 0
		for i, d := range sudoku.Difficulties {
			if d.Name == m.difficulty.Name {
				m.menuCursor = i
			}
		}
		return m.start(sudoku.Difficulties[m.menuCursor]), nil
	case "r":
		m.restart()
	}
	return m, nil
}

// ------------------------------------------------------------------ souris --

func (m *Model) handleMouse(msg tea.MouseMsg) {
	if m.screen != scrPlay || m.paused || m.showHelp {
		return
	}
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return
	}
	r, c, ok := m.cellAt(msg.X, msg.Y)
	if !ok {
		return
	}
	m.cursorR, m.cursorC = r, c
}

// ------------------------------------------------------------------- jeu --

func (m *Model) start(d sudoku.Difficulty) Model {
	puzzle, solution := sudoku.Generate(d, m.rng)
	m.difficulty = d
	m.puzzle = puzzle
	m.board = puzzle
	m.solution = solution
	m.notes = [9][9]uint16{}
	m.history = nil
	m.elapsed = 0
	m.mistakes = 0
	m.hints = 0
	m.noteMode = false
	m.paused = false
	m.showHelp = false
	m.celebrate = false
	m.screen = scrPlay
	m.cursorR, m.cursorC = m.firstEmpty()
	m.recompute()
	m.setStatus(styleGood, "Bonne chance !")
	return *m
}

func (m Model) firstEmpty() (int, int) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if m.board[r][c] == 0 {
				return r, c
			}
		}
	}
	return 0, 0
}

func (m *Model) move(dr, dc int) {
	if nr := m.cursorR + dr; nr >= 0 && nr < 9 {
		m.cursorR = nr
	}
	if nc := m.cursorC + dc; nc >= 0 && nc < 9 {
		m.cursorC = nc
	}
}

func (m *Model) push() {
	m.history = append(m.history, snapshot{m.board, m.notes, m.mistakes})
	if len(m.history) > 500 {
		m.history = m.history[1:]
	}
}

func (m *Model) undo() {
	if len(m.history) == 0 {
		m.setStatus(styleWarn, "Rien à annuler")
		return
	}
	s := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.board, m.notes, m.mistakes = s.board, s.notes, s.mistakes
	m.recompute()
	m.setStatus(styleMuted, "↶ Coup annulé")
}

func (m *Model) place(v int) {
	r, c := m.cursorR, m.cursorC
	if m.puzzle[r][c] != 0 {
		m.setStatus(styleWarn, "Case donnée par l'énigme")
		return
	}
	if m.noteMode {
		if m.board[r][c] != 0 {
			return
		}
		m.push()
		m.notes[r][c] ^= 1 << uint(v)
		return
	}
	m.push()
	if m.board[r][c] == v {
		m.board[r][c] = 0
		m.setStatus(styleMuted, "Case effacée")
	} else {
		m.board[r][c] = v
		m.notes[r][c] = 0
		m.clearPeerNotes(r, c, v)
		if m.conflicts[r][c] {
			m.mistakes++
			m.setStatus(styleErr, "✗ Conflit sur cette case")
		} else {
			m.setStatus(styleGood, "✓ Chiffre placé")
		}
	}
	m.recompute()
	m.checkWin()
}

func (m *Model) erase() {
	r, c := m.cursorR, m.cursorC
	if m.puzzle[r][c] != 0 {
		m.setStatus(styleWarn, "Case donnée par l'énigme")
		return
	}
	if m.board[r][c] == 0 && m.notes[r][c] == 0 {
		return
	}
	m.push()
	m.board[r][c] = 0
	m.notes[r][c] = 0
	m.recompute()
	m.setStatus(styleMuted, "Case effacée")
}

func (m *Model) hint() {
	r, c := m.cursorR, m.cursorC
	if m.puzzle[r][c] != 0 {
		m.setStatus(styleWarn, "Case donnée par l'énigme")
		return
	}
	if m.board[r][c] == m.solution[r][c] {
		m.setStatus(styleMuted, "Déjà correct")
		return
	}
	m.push()
	m.board[r][c] = m.solution[r][c]
	m.notes[r][c] = 0
	m.clearPeerNotes(r, c, m.solution[r][c])
	m.hints++
	m.recompute()
	m.setStatus(styleAccent, "✦ Indice révélé")
	m.checkWin()
}

func (m *Model) restart() {
	m.board = m.puzzle
	m.notes = [9][9]uint16{}
	m.history = nil
	m.elapsed = 0
	m.mistakes = 0
	m.hints = 0
	m.paused = false
	m.celebrate = false
	m.screen = scrPlay
	m.cursorR, m.cursorC = m.firstEmpty()
	m.recompute()
	m.setStatus(styleMuted, "Grille réinitialisée")
}

func (m *Model) clearPeerNotes(r, c, v int) {
	bit := uint16(1) << uint(v)
	for i := 0; i < 9; i++ {
		m.notes[r][i] &^= bit
		m.notes[i][c] &^= bit
	}
	br, bc := r/3*3, c/3*3
	for i := br; i < br+3; i++ {
		for j := bc; j < bc+3; j++ {
			m.notes[i][j] &^= bit
		}
	}
}

func (m *Model) recompute() { m.conflicts = m.board.Conflicts() }

func (m *Model) checkWin() {
	if m.board.IsSolved() {
		m.screen = scrWin
		m.noteMode = false
		m.celebrate = true
	}
}

func (m *Model) setStatus(st lipgloss.Style, text string) {
	m.status, m.statusStyle = text, st
}

// remaining renvoie combien de fois le chiffre v reste à placer.
func (m Model) remaining(v int) int {
	n := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if m.board[r][c] == v {
				n++
			}
		}
	}
	return 9 - n
}

// isPeer indique si la case partage ligne, colonne ou bloc avec le curseur.
func (m Model) isPeer(r, c int) bool {
	return r == m.cursorR || c == m.cursorC ||
		(r/3 == m.cursorR/3 && c/3 == m.cursorC/3)
}

// sameDigit indique si la case porte le même chiffre que la case curseur.
func (m Model) sameDigit(r, c int) bool {
	v := m.board[m.cursorR][m.cursorC]
	return v != 0 && m.board[r][c] == v
}
