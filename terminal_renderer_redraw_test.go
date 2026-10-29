package uv

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

// redrawFrame is a frame of styled rows, some with wide cells, row n
// counting k: what an application draws whole over a Clear on every frame.
func redrawFrame(w, h, n, k int) string {
	rows := make([]string, h)
	for y := range rows {
		rows[y] = fmt.Sprintf("\x1b[1m%02d\x1b[m 日本語 👍 \x1b[38;5;%dmrow\x1b[m", y, 16+y)
		if y == n {
			rows[y] += fmt.Sprintf(" %d", k)
		}
	}
	return strings.Join(rows, "\n")
}

// TestRedrawWholeWritesChangedRows: a frame drawn whole over a Clear, as
// Bubble Tea draws each one, touches every row, but only the row that
// changed reaches the terminal, wide cells and all.
func TestRedrawWholeWritesChangedRows(t *testing.T) {
	const w, h = 40, 10
	var out bytes.Buffer
	r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	buf := NewScreenBuffer(w, h)

	draw := func(k int) string {
		t.Helper()
		out.Reset()
		buf.Clear()
		NewStyledString(redrawFrame(w, h, 3, k)).Draw(buf, buf.Bounds())
		r.Render(buf.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}

	draw(0)
	got := draw(1)
	if n := strings.Count(got, "row"); n != 1 || !strings.Contains(got, "1") {
		t.Fatalf("one row changed, %d rows written: %q", n, got)
	}
	if got := draw(1); strings.Contains(got, "row") {
		t.Fatalf("nothing changed, rows written: %q", got)
	}
}

// BenchmarkRedrawWhole draws a 220x55 frame whole over a Clear and renders
// it, one row changed each time, as Bubble Tea does on every frame.
func BenchmarkRedrawWhole(b *testing.B) {
	const w, h = 220, 55
	r := NewTerminalRenderer(io.Discard, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	buf := NewScreenBuffer(w, h)
	frames := []string{redrawFrame(w, h, 20, 0), redrawFrame(w, h, 20, 1)}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		buf.Clear()
		NewStyledString(frames[i%2]).Draw(buf, buf.Bounds())
		r.Render(buf.RenderBuffer)
		_ = r.Flush()
		i++
	}
}
