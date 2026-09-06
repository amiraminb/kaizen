package render

import (
	"io"
	"os"
	"strings"

	"github.com/amiraminb/kaizen/internal/model"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	GlyphDone          = "✓"
	GlyphSkipped       = "~"
	GlyphPending       = "◦"
	GlyphMiss          = "✗"
	GlyphNotApplicable = "·"
)

type Renderer struct {
	done    lipgloss.Style
	skipped lipgloss.Style
	pending lipgloss.Style
	miss    lipgloss.Style
	absent  lipgloss.Style
	label   lipgloss.Style
	muted   lipgloss.Style
}

// lipgloss already downgrades a non-TTY writer to Ascii; NO_COLOR is honoured on top
// of that because the glyphs carry every state without needing colour.
func New(out io.Writer) *Renderer {
	renderer := lipgloss.NewRenderer(out)
	if os.Getenv("NO_COLOR") != "" {
		renderer.SetColorProfile(termenv.Ascii)
	}

	style := func(color string) lipgloss.Style {
		return renderer.NewStyle().Foreground(lipgloss.Color(color))
	}
	return &Renderer{
		done:    style("2"),
		skipped: style("4"),
		pending: style("3"),
		miss:    style("1"),
		absent:  style("8"),
		label:   renderer.NewStyle().Bold(true),
		muted:   style("8"),
	}
}

func (r *Renderer) Glyph(status model.DayStatus) string {
	switch status {
	case model.DayDone:
		return r.done.Render(GlyphDone)
	case model.DaySkipped:
		return r.skipped.Render(GlyphSkipped)
	case model.DayPending:
		return r.pending.Render(GlyphPending)
	case model.DayMiss:
		return r.miss.Render(GlyphMiss)
	default:
		return r.absent.Render(GlyphNotApplicable)
	}
}

func (r *Renderer) Label(text string) string {
	return r.label.Render(text)
}

func (r *Renderer) Muted(text string) string {
	return r.muted.Render(text)
}

// Padding uses lipgloss widths rather than tabwriter because tabwriter counts ANSI
// escape bytes as visible characters and misaligns every coloured column.
func Pad(text string, width int) string {
	gap := width - lipgloss.Width(text)
	if gap <= 0 {
		return text
	}
	return text + strings.Repeat(" ", gap)
}

func StatusWord(status model.DayStatus) string {
	switch status {
	case model.DayDone:
		return "done"
	case model.DaySkipped:
		return "skipped"
	case model.DayPending:
		return "pending"
	case model.DayMiss:
		return "missed"
	default:
		return "n/a"
	}
}
