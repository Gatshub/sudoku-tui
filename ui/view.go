package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"sudoku-tui/sudoku"
)

// cellW renvoie la largeur d'une case, en colonnes.
func (m Model) cellW() int {
	if m.cw > 0 {
		return m.cw
	}
	return 3
}

// cellH renvoie la hauteur d'une case (3 = avec notes, 1 = compact).
func (m Model) cellH() int {
	if m.ch > 0 {
		return m.ch
	}
	return 3
}

// chooseCellW choisit la largeur de case qui rend la grille carrée à l'écran,
// compte tenu du rapport hauteur/largeur des caractères du terminal :
// on cherche 9*largeur+4 ≈ aspect * (9*hauteur+4). Parmi les largeurs entières
// possibles, on retient celle dont le rapport visuel est le plus proche de 1
// (comparaison en échelle logarithmique, donc symétrique « trop large » /
// « trop haut »).
func (m Model) chooseCellW(ch int) int {
	a := m.aspectOr()
	lo, hi := 2, 9
	if ch == 3 {
		lo = 4 // il faut au moins 3 colonnes pour la mini-grille de notes
	}
	visualH := a * float64(9*ch+4)
	best, bestErr := lo, math.MaxFloat64
	for w := lo; w <= hi; w++ {
		err := math.Abs(math.Log(float64(9*w+4) / visualH))
		if err < bestErr {
			best, bestErr = w, err
		}
	}
	return best
}

// aspectOr renvoie le rapport hauteur/largeur d'un caractère, avec la valeur
// usuelle 2:1 en secours.
func (m Model) aspectOr() float64 {
	if m.aspect <= 0 {
		return 2.0
	}
	return m.aspect
}

func (m Model) gridW() int { return 9*m.cellW() + 4 }
func (m Model) gridH() int { return 9*m.cellH() + 4 }

// withBG applique un fond seulement s'il est défini (sinon on laisse le
// terminal tel quel, ce qui rend bien sur les terminaux transparents).
func withBG(st lipgloss.Style, bg lipgloss.TerminalColor) lipgloss.Style {
	if bg != nil {
		return st.Background(bg)
	}
	return st
}

// cellColors calcule le fond et la couleur de texte d'une case.
func (m Model) cellColors(r, c int) (bg, fg lipgloss.TerminalColor) {
	given := m.puzzle[r][c] != 0
	cursor := !m.celebrate && r == m.cursorR && c == m.cursorC

	// Sur l'écran de victoire, on n'affiche aucun surlignage : la grille
	// résolue se lit d'un seul coup d'œil.
	if !m.celebrate {
		switch {
		case cursor:
			bg = colCursorBG
		case m.sameDigit(r, c):
			bg = colSameBG
		case m.isPeer(r, c):
			bg = colPeerBG
		}
	}

	switch {
	case m.conflicts[r][c]:
		fg = colConflict
	case cursor:
		fg = colCursorFG
	case given:
		fg = colGiven
	case m.celebrate:
		fg = colGreen
	default:
		fg = colUser
	}
	return
}

// cellContent renvoie le texte brut d'une case pour une ligne donnée.
func (m Model) cellContent(r, c, line int) string {
	v := m.board[r][c]
	if m.cellH() == 1 {
		if v != 0 {
			return strconv.Itoa(v)
		}
		if m.notes[r][c] != 0 {
			return "·"
		}
		return " "
	}
	if v != 0 {
		if line == 1 {
			return strconv.Itoa(v)
		}
		return " "
	}
	var b [3]byte
	for i := 0; i < 3; i++ {
		d := line*3 + i + 1
		if m.notes[r][c]&(1<<uint(d)) != 0 {
			b[i] = byte('0' + d)
		} else {
			b[i] = ' '
		}
	}
	return string(b[:])
}

// renderCellLine rend une ligne (3 colonnes) d'une case, entièrement stylée.
func (m Model) renderCellLine(r, c, line int) string {
	cw := m.cellW()
	bg, fg := m.cellColors(r, c)
	given := m.puzzle[r][c] != 0
	cursor := !m.celebrate && r == m.cursorR && c == m.cursorC

	base := lipgloss.NewStyle().Width(cw).Align(lipgloss.Center).Foreground(fg)
	if given && !cursor {
		base = base.Bold(true)
	}
	base = withBG(base, bg)

	// Cas des notes : chaque chiffre est coloré séparément (gris).
	if m.cellH() == 3 && m.board[r][c] == 0 {
		noteFG := colNote
		if cursor {
			noteFG = colCursorFG
		}
		// La mini-grille 3x3 des notes est centrée dans la case.
		padL := (cw - 3) / 2
		padR := cw - 3 - padL
		pad := func(n int) string {
			return withBG(lipgloss.NewStyle(), bg).Render(strings.Repeat(" ", n))
		}
		out := pad(padL)
		for i := 0; i < 3; i++ {
			d := line*3 + i + 1
			ch := " "
			if m.notes[r][c]&(1<<uint(d)) != 0 {
				ch = strconv.Itoa(d)
			}
			st := withBG(lipgloss.NewStyle().Width(1).Foreground(noteFG), bg)
			out += st.Render(ch)
		}
		return out + pad(padR)
	}

	return base.Render(m.cellContent(r, c, line))
}

func (m Model) borderLine(left, mid, right string) string {
	seg := strings.Repeat(bHoriz, 3*m.cellW())
	return styleLine.Render(left + seg + mid + seg + mid + seg + right)
}

// renderGrid produit toutes les lignes de la grille 9x9 encadrée.
func (m Model) renderGrid() []string {
	ch := m.cellH()
	lines := make([]string, 0, 9*ch+4)
	lines = append(lines, m.borderLine(bTopLeft, bTopT, bTopRight))
	for r := 0; r < 9; r++ {
		for ln := 0; ln < ch; ln++ {
			var sb strings.Builder
			sb.WriteString(styleLine.Render(bVert))
			for c := 0; c < 9; c++ {
				sb.WriteString(m.renderCellLine(r, c, ln))
				if c%3 == 2 && c != 8 {
					sb.WriteString(styleLine.Render(bVert))
				}
			}
			sb.WriteString(styleLine.Render(bVert))
			lines = append(lines, sb.String())
		}
		if r%3 == 2 && r != 8 {
			lines = append(lines, m.borderLine(bLeftT, bCross, bRightT))
		}
	}
	lines = append(lines, m.borderLine(bBotLeft, bBotT, bBotRight))
	return lines
}

// ---------------------------------------------------------------- écrans --

// View construit l'affichage complet.
func (m Model) View() string {
	if !m.ready {
		return ""
	}
	// Largeur minimale : celle de la mise en page la plus étroite possible.
	if m.width < 9*m.chooseCellW(1)+4 || m.height < 20 {
		return m.viewTooSmall()
	}
	if m.showHelp {
		return m.assemble(m.helpLines())
	}
	switch m.screen {
	case scrMenu:
		return m.viewMenu()
	case scrWin:
		return m.viewWin()
	default:
		return m.viewPlay()
	}
}

func (m Model) viewTooSmall() string {
	msg := styleWarn.Render(fmt.Sprintf("Terminal trop petit : %dx%d", m.width, m.height)) +
		"\n\n" + styleMuted.Render(fmt.Sprintf("Minimum requis : %d colonnes", 9*m.chooseCellW(1)+4)) +
		"\n" + styleMuted.Render("et 20 lignes.") +
		"\n\n" + styleMuted.Render("Agrandissez la fenêtre.")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Align(lipgloss.Center).Render(msg))
}

func (m Model) viewMenu() string {
	items := make([]string, 0, len(sudoku.Difficulties))
	for i, d := range sudoku.Difficulties {
		text := fmt.Sprintf("%s  ·  %d indices", d.Name, d.Clues)
		if i == m.menuCursor {
			items = append(items, styleSelItem.Render("▶ "+text))
		} else {
			items = append(items, styleMenuItem.Render("  "+text))
		}
	}
	lines := []string{
		styleTitle.Render("S U D O K U"),
		styleMuted.Render("Un sudoku élégant pour votre terminal"),
		"",
		styleBox.Render(strings.Join(items, "\n")),
		"",
		styleMuted.Render("↑ ↓ choisir   ·   Entrée lancer   ·   q quitter"),
	}
	return m.assemble(lines)
}

func (m Model) viewPlay() string {
	lines, _, _ := m.fitPlay()
	return m.assemble(lines)
}

// fitPlay choisit la mise en page : grille haute (avec notes) si elle tient
// dans le terminal, sinon grille compacte. Dans les deux cas la largeur de case
// est calculée pour que la grille paraisse carrée. Renvoie les lignes ainsi que
// les dimensions de case retenues, pour que le clic souris reste synchronisé.
func (m Model) fitPlay() ([]string, int, int) {
	cw, ch := m.chooseLayout(func(mm Model) []string { return mm.buildPlay() })
	mm := m
	mm.cw, mm.ch = cw, ch
	return mm.buildPlay(), cw, ch
}

// chooseLayout retient la première mise en page (haute, puis compacte) dont le
// contenu tient dans le terminal. « build » construit le contenu de l'écran
// concerné, car la grille jouée et la grille de victoire n'ont pas la même
// hauteur de pied de page.
func (m Model) chooseLayout(build func(Model) []string) (int, int) {
	for _, ch := range []int{3, 1} {
		lo := 2
		if ch == 3 {
			lo = 4
		}
		visualH := m.aspectOr() * float64(9*ch+4)
		// On part de la largeur idéale (grille carrée) et on la réduit si le
		// terminal est trop étroit, tant que la grille reste plausible.
		for cw := m.chooseCellW(ch); cw >= lo; cw-- {
			if math.Abs(math.Log(float64(9*cw+4)/visualH)) > 0.25 {
				break // trop étirée : mieux vaut changer de mise en page
			}
			mm := m
			mm.cw, mm.ch = cw, ch
			if lines := build(mm); len(lines) <= m.height && mm.gridW() <= m.width {
				return cw, ch
			}
		}
	}
	return m.chooseCellW(1), 1
}

func (m Model) buildPlay() []string {
	lines := []string{styleTitle.Render("S U D O K U"), styleSubtitle.Render(m.subtitle())}

	if m.paused {
		lines = append(lines, m.pauseLines()...)
	} else {
		lines = append(lines, m.renderGrid()...)
	}

	lines = append(lines, "")
	lines = append(lines, m.statusStyle.Render(m.status))
	if rl := m.remainingLine(); rl != "" {
		lines = append(lines, rl)
	}
	lines = append(lines, m.hintsLine())
	return lines
}

func (m Model) pauseLines() []string {
	box := styleBox.Render(
		styleTitle.Render("EN PAUSE") + "\n\n" +
			styleMuted.Render("Appuyez sur p pour reprendre"))
	// On garde la même hauteur que la grille pour éviter tout saut.
	h := m.gridH()
	pad := (h - lipgloss.Height(box)) / 2
	var out []string
	for i := 0; i < pad; i++ {
		out = append(out, "")
	}
	out = append(out, strings.Split(box, "\n")...)
	for len(out) < h {
		out = append(out, "")
	}
	return out
}

func (m Model) viewWin() string {
	// La grille de victoire doit elle aussi paraître carrée.
	cw, ch := m.chooseLayout(func(mm Model) []string { return mm.buildWin() })
	mm := m
	mm.cw, mm.ch = cw, ch
	return m.assemble(mm.buildWin())
}

func (m Model) buildWin() []string {
	lines := []string{styleTitle.Render("S U D O K U"), styleSubtitle.Render(m.subtitle())}
	lines = append(lines, m.renderGrid()...)
	lines = append(lines, "")
	lines = append(lines, styleGood.Render("★  Bravo, grille résolue !  ★"))
	lines = append(lines, styleMuted.Render(fmt.Sprintf(
		"Temps %s   ·   Erreurs %d   ·   Indices %d",
		fmtDuration(m.elapsed), m.mistakes, m.hints)))
	lines = append(lines, "")
	lines = append(lines, styleMuted.Render("Entrée nouvelle partie   ·   m menu   ·   q quitter"))
	return lines
}

// ------------------------------------------------------------- fragments --

func (m Model) subtitle() string {
	if m.screen == scrMenu {
		return ""
	}
	parts := []string{m.difficulty.Name, "◷ " + fmtDuration(m.elapsed)}
	if m.mistakes > 0 {
		parts = append(parts, fmt.Sprintf("erreurs %d", m.mistakes))
	}
	if m.hints > 0 {
		parts = append(parts, fmt.Sprintf("indices %d", m.hints))
	}
	if m.noteMode {
		parts = append(parts, styleAccent.Render("✎ NOTES"))
	}
	if m.paused {
		parts = append(parts, styleWarn.Render("EN PAUSE"))
	}
	return strings.Join(parts, styleMuted.Render("  ·  "))
}

// remainingLine affiche combien de fois chaque chiffre reste à placer.
func (m Model) remainingLine() string {
	cur := 0
	if !m.celebrate {
		cur = m.board[m.cursorR][m.cursorC]
	}
	groups := func() string {
		var b strings.Builder
		for d := 1; d <= 9; d++ {
			n := m.remaining(d)
			switch {
			case n == 0:
				b.WriteString(styleMuted.Render(fmt.Sprintf("%d✓", d)))
			case d == cur:
				b.WriteString(lipgloss.NewStyle().Foreground(colAccent).Bold(true).Render(strconv.Itoa(d)))
				b.WriteString(styleMuted.Render(strconv.Itoa(n)))
			default:
				b.WriteString(lipgloss.NewStyle().Foreground(colFG).Render(strconv.Itoa(d)))
				b.WriteString(styleMuted.Render(strconv.Itoa(n)))
			}
			b.WriteString("  ")
		}
		return b.String()
	}
	// Avec libellé si la place le permet, sinon sans, sinon on masque la ligne.
	if full := styleMuted.Render("Reste  ") + groups(); lipgloss.Width(full) <= m.width {
		return full
	}
	if compact := strings.TrimRight(groups(), " "); lipgloss.Width(compact) <= m.width {
		return compact
	}
	return ""
}

func (m Model) hintsLine() string {
	full := [][2]string{
		{"hjkl/←→", "déplacer"}, {"1-9", "chiffre"}, {"0", "effacer"},
		{"n", "notes"}, {"u", "annuler"}, {"H", "indice"},
		{"?", "aide"}, {"q", "quitter"},
	}
	short := [][2]string{
		{"hjkl", "déplacer"}, {"1-9", "chiffre"}, {"n", "notes"},
		{"u", "annuler"}, {"?", "aide"}, {"q", "quitter"},
	}
	mini := [][2]string{{"?", "aide"}, {"q", "quitter"}}

	build := func(items [][2]string) string {
		parts := make([]string, 0, len(items))
		for _, it := range items {
			parts = append(parts, styleHelpKey.Render(it[0])+styleMuted.Render(" "+it[1]))
		}
		return strings.Join(parts, styleMuted.Render("  ·  "))
	}
	// On garde la variante la plus complète qui tient dans la largeur.
	for _, items := range [][][2]string{full, short, mini} {
		if l := build(items); lipgloss.Width(l) <= m.width {
			return l
		}
	}
	return build(mini)
}

func (m Model) helpLines() []string {
	full := [][2]string{
		{"↑ ↓ ← →  ou  h j k l", "déplacer le curseur"},
		{"1 … 9", "saisir un chiffre (ou basculer une note)"},
		{"0 / Retour arrière / Suppr", "effacer la case"},
		{"n  ou  Espace", "activer / désactiver le mode notes"},
		{"u", "annuler le dernier coup"},
		{"H", "révéler la solution de la case"},
		{"r", "réinitialiser la grille"},
		{"p", "mettre en pause"},
		{"souris", "cliquer sur une case pour la sélectionner"},
		{"?", "fermer cette aide"},
		{"q", "quitter"},
	}
	compact := [][2]string{
		{"← ↑ ↓ →  hjkl", "déplacer"},
		{"1 … 9", "chiffre / note"},
		{"0 ⌫ Suppr", "effacer"},
		{"n / Espace", "mode notes"},
		{"u", "annuler"},
		{"H", "indice"},
		{"r", "rejouer"},
		{"p", "pause"},
		{"?", "fermer"},
		{"q", "quitter"},
	}
	const tip = "Astuce : la ligne « Reste » indique les chiffres manquants."

	render := func(rows [][2]string, withTip bool) []string {
		width := 0
		for _, r := range rows {
			if w := lipgloss.Width(r[0]); w > width {
				width = w
			}
		}
		var b strings.Builder
		b.WriteString(styleTitle.Render("Aide — Sudoku") + "\n\n")
		for _, r := range rows {
			pad := strings.Repeat(" ", width-lipgloss.Width(r[0])+3)
			b.WriteString(styleHelpKey.Render(r[0]) + pad + styleMuted.Render(r[1]) + "\n")
		}
		if withTip {
			b.WriteString("\n" + styleMuted.Render(tip))
		}
		return strings.Split(styleBox.Render(strings.TrimRight(b.String(), "\n")), "\n")
	}

	lines := render(full, true)
	if lipgloss.Width(lines[0]) > m.width {
		lines = render(compact, false)
	}
	return lines
}

// ------------------------------------------------------------- assemblage --

// assemble centre horizontalement chaque ligne puis le bloc verticalement.
func (m Model) assemble(lines []string) string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, l)
	}
	top := 0
	if gap := (m.height - len(out)) / 2; gap > 0 {
		top = gap
	}
	return strings.Repeat("\n", top) + strings.Join(out, "\n")
}

// -------------------------------------------------------------- utilitaire --

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	mn := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, mn, s)
	}
	return fmt.Sprintf("%02d:%02d", mn, s)
}

// cellAt convertit des coordonnées terminal en (ligne, colonne) de la grille.
// Renvoie ok=false si le clic tombe hors des cases.
func (m Model) cellAt(x, y int) (int, int, bool) {
	lines, cw, ch := m.fitPlay() // exactement la même mise en page que l'affichage
	mm := m
	mm.cw, mm.ch = cw, ch
	top := 0
	if gap := (m.height - len(lines)) / 2; gap > 0 {
		top = gap
	}
	gridY := top + 2 // titre + sous-titre
	gridX := (m.width - mm.gridW()) / 2

	ly := y - gridY
	lx := x - gridX
	if ly < 1 || ly > 9*ch+2 || lx < 1 || lx > mm.gridW()-2 {
		return 0, 0, false
	}

	ry := ly - 1
	row := -1
	for r := 0; r < 9; r++ {
		if ry < ch {
			row = r
			break
		}
		ry -= ch
		if r%3 == 2 && r != 8 {
			ry--
		}
	}
	cx := lx - 1
	col := -1
	for c := 0; c < 9; c++ {
		if cx < cw {
			col = c
			break
		}
		cx -= cw
		if c%3 == 2 && c != 8 {
			cx--
		}
	}
	if row < 0 || col < 0 {
		return 0, 0, false
	}
	return row, col, true
}
