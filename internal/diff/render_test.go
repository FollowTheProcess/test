package diff_test

import (
	"testing"

	"go.followtheprocess.codes/hue"
	"go.followtheprocess.codes/test/internal/diff"
)

func TestMain(m *testing.M) {
	// Force colour on so rendered output is predictable in tests.
	hue.Enabled(true)
	m.Run()
}

func TestRender(t *testing.T) {
	tests := []struct {
		name  string
		want  string
		lines []diff.Line
	}{
		{
			name:  "nil input returns empty string",
			lines: nil,
			want:  "",
		},
		{
			name:  "empty slice returns empty string",
			lines: []diff.Line{},
			want:  "",
		},
		{
			name: "context line has no colour and double-space prefix",
			lines: []diff.Line{
				{Kind: diff.KindContext, Content: []byte("unchanged\n")},
			},
			want: "  unchanged\n",
		},
		{
			name: "header line is cyan+bold",
			lines: []diff.Line{
				{Kind: diff.KindHeader, Content: []byte("diff want got\n")},
			},
			want: (hue.Cyan | hue.Bold).Text("diff want got\n"),
		},
		{
			name: "standalone removed line has red prefix and content",
			lines: []diff.Line{
				{Kind: diff.KindRemoved, Content: []byte("old line\n")},
			},
			want: hue.Red.Text("- ") + hue.Red.Text("old line\n"),
		},
		{
			name: "standalone added line has green prefix and content",
			lines: []diff.Line{
				{Kind: diff.KindAdded, Content: []byte("new line\n")},
			},
			want: hue.Green.Text("+ ") + hue.Green.Text("new line\n"),
		},
		{
			// "old" and "new" share no characters, so CharDiff produces a single
			// Changed segment per side with the \n reattached as a separate unchanged segment.
			name: "matched removed+added pair uses inline char diff with coloured prefixes",
			lines: []diff.Line{
				{Kind: diff.KindRemoved, Content: []byte("old\n")},
				{Kind: diff.KindAdded, Content: []byte("new\n")},
			},
			want: hue.Red.Text("- ") + hue.BrightRedBackground.Text("old") + hue.Red.Text("\n") +
				hue.Green.Text("+ ") + hue.BrightGreenBackground.Text("new") + hue.Green.Text("\n"),
		},
		{
			name: "mismatched count (2 removed, 1 added) uses whole-line colour with coloured prefixes",
			lines: []diff.Line{
				{Kind: diff.KindRemoved, Content: []byte("line one\n")},
				{Kind: diff.KindRemoved, Content: []byte("line two\n")},
				{Kind: diff.KindAdded, Content: []byte("replacement\n")},
			},
			want: hue.Red.Text("- ") + hue.Red.Text("line one\n") +
				hue.Red.Text("- ") + hue.Red.Text("line two\n") +
				hue.Green.Text("+ ") + hue.Green.Text("replacement\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diff.Render(tt.lines)

			if got != tt.want {
				t.Errorf("Render() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

// TestVisualDiff is a manual smoke-check for the diff renderer.
// Run with go test -v to see the colourised output in your terminal.
func TestVisualDiff(t *testing.T) {
	// TestMain enables colour, so all rendering below is colourised.
	scenarios := []struct {
		name string
		old  string
		new  string
	}{
		{
			// Single changed line: char-level inline highlighting should show
			// exactly which characters differ.
			name: "single line change (inline char diff)",
			old: `func greet(name string) string {
	return "Hello, " + name
}
`,
			new: `func greet(name string) string {
	return "Hello, " + name + "!"
}
`,
		},
		{
			// Two changed lines paired 1:1: each pair gets its own inline diff.
			name: "multi-line paired change",
			old: `func (s *Server) Start(port int) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	return http.ListenAndServe(addr, s.mux)
}
`,
			new: `func (s *Server) Start(ctx context.Context, port int) error {
	addr := fmt.Sprintf(":%d", port)
	return s.httpServer.ListenAndServeContext(ctx, addr)
}
`,
		},
		{
			// More removed than added: mismatched counts fall back to whole-line colour.
			name: "mismatched counts (whole-line fallback)",
			old: `case "json":
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
`,
			new: `case "json":
	return json.NewEncoder(w).Encode(v)
`,
		},
		{
			// Unicode content: char diff should handle multi-byte runes correctly.
			name: "unicode content",
			old:  "Héllo, wörld! Ünïcödé is fün.\n",
			new:  "Héllo, wörld! Ünïcödé is grëat.\n",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			lines := diff.Lines("want", []byte(sc.old), "got", []byte(sc.new))
			t.Logf("\n=== %s ===\n%s\n", sc.name, diff.Render(lines))
		})
	}
}

// BenchmarkRender benchmarks Render using a realistic diff.
func BenchmarkRender(b *testing.B) {
	old := []byte("the quick brown fox\njumps over the lazy dog\nsome context\nmore context\n")
	newContent := []byte("the quick brown cat\njumps over the lazy frog\nsome context\nmore context\n")
	lines := diff.Lines("want", old, "got", newContent)

	b.ResetTimer()

	for b.Loop() {
		diff.Render(lines)
	}
}
