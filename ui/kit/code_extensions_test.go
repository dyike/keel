package kit

import (
	"image/color"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
)

func TestCodeSearchSessionPublic(t *testing.T) {
	ed, h := focusedEditor("猫 Cat catapult\ncat 猫")
	if s := ed.SearchSession(); s.Active || s.Current != -1 {
		t.Fatalf("fresh: %+v", s)
	}
	ed.SetSearchQuery("cat", CodeSearchOptions{WholeWord: true})
	ed.Searchable(false)
	s := ed.SearchSession()
	if !s.Active || s.PanelOpen || len(s.Matches) != 2 || s.Current != -1 {
		t.Fatalf("custom: %+v", s)
	}
	s.Matches[0].Col = 99
	if !ed.SelectSearchMatch(0) || ed.Selection() != "Cat" || ed.SearchSession().Current != 0 {
		t.Fatal("selection or snapshot ownership")
	}
	ed.PreviousSearchMatch()
	if ed.SearchSession().Current != 1 {
		t.Fatal("previous wrap")
	}
	ed.NextSearchMatch()
	calls := 0
	ed.OnChange(func(string) { calls++ })
	if !ed.ReplaceCurrentSearchMatch("犬") || ed.Value() != "猫 犬 catapult\ncat 猫" || calls != 1 {
		t.Fatal("replace current", ed.Value(), calls)
	}
	h.Key("Z", key.ModShortcut)
	if ed.Value() != "猫 Cat catapult\ncat 猫" {
		t.Fatal("undo", ed.Value())
	}
	ed.SetSearchQuery("(", CodeSearchOptions{Regex: true})
	if !ed.SearchSession().InvalidPattern || ed.ReplaceAllSearchMatches("x") != 0 {
		t.Fatal("invalid regex")
	}
	ed.SetSearchQuery("猫", CodeSearchOptions{})
	ed.SetReadOnly(true)
	if ed.ReplaceAllSearchMatches("x") != 0 {
		t.Fatal("read only")
	}
	ed.SetReadOnly(false)
	ed.SetDisabled(true)
	if ed.ReplaceAllSearchMatches("x") != 0 {
		t.Fatal("disabled")
	}
	ed.CloseSearch()
	if s = ed.SearchSession(); s.Active || len(s.Matches) != 0 || s.Current != -1 {
		t.Fatalf("closed: %+v", s)
	}
}

func TestCodeSearchAllBeyondCapAndUndo(t *testing.T) {
	original := strings.Repeat("a a\n", 5001)
	ed := CodeEditor(original)
	ed.SetSearchQuery("a", CodeSearchOptions{})
	if s := ed.SearchSession(); !s.Truncated || len(s.Matches) != 10000 {
		t.Fatal("cap", len(s.Matches), s.Truncated)
	}
	calls := 0
	ed.OnChange(func(string) { calls++ })
	if n := ed.ReplaceAllSearchMatches("猫\n犬"); n != 10002 || calls != 1 {
		t.Fatal("replace count", n, calls)
	}
	want := strings.ReplaceAll(original, "a", "猫\n犬")
	if ed.Value() != want {
		t.Fatal("replace all text")
	}
	if _, ok := ed.buf.undoStep(); !ok || ed.Value() != original {
		t.Fatal("single undo")
	}
	if _, ok := ed.buf.redoStep(); !ok || ed.Value() != want {
		t.Fatal("single redo")
	}
	ed.SetValue("a1 a22\na333")
	ed.SetSearchQuery(`a(\d+)`, CodeSearchOptions{Regex: true})
	if ed.ReplaceAllSearchMatches("${1}猫") != 3 || ed.Value() != "1猫 22猫\n333猫" {
		t.Fatal("regex captures", ed.Value())
	}
}

func TestCodeDecorationTracking(t *testing.T) {
	ed := CodeEditor("甲abc乙\ntail")
	c := ed.Decorations(CodeDecoration{Range: CodeRange{0, 1, 0, 4}})
	check := func(want CodeRange) {
		t.Helper()
		got := c.Get()
		if len(got) != 1 || got[0].Range != want {
			t.Fatalf("decoration = %+v want %+v", got, want)
		}
	}
	edit := func(a, z codePos, s string) {
		ed.buf.begin(ed.sels)
		ed.buf.edit(a, z, s)
		ed.buf.commit(ed.sels, false)
	}
	edit(codePos{0, 1}, codePos{0, 1}, "始") // start excluded
	check(CodeRange{0, 2, 0, 5})
	edit(codePos{0, 5}, codePos{0, 5}, "終") // end excluded
	check(CodeRange{0, 2, 0, 5})
	edit(codePos{0, 3}, codePos{0, 3}, "中\n文")
	check(CodeRange{0, 2, 1, 3})
	ed.buf.undoStep()
	check(CodeRange{0, 2, 0, 5})
	ed.buf.redoStep()
	check(CodeRange{0, 2, 1, 3})
	ed.SetValue("前\n" + ed.Value())
	check(CodeRange{1, 2, 2, 3})
	edit(codePos{1, 2}, codePos{2, 3}, "")
	if len(c.Get()) != 0 {
		t.Fatal("delete removes annotation")
	}
	ed.buf.undoStep()
	if len(c.Get()) != 0 {
		t.Fatal("undo must not resurrect deleted annotation")
	}
}

func TestCodeDecorationOwnershipAndIndex(t *testing.T) {
	ed := CodeEditor(strings.Repeat("abcdef\n", 100))
	red := color.NRGBA{R: 255, A: 255}
	c := ed.Decorations(CodeDecoration{Range: CodeRange{-1, 999, 101, 0}, Color: &red})
	red.R = 0
	got := c.Get()
	if len(got) != 1 || got[0].Range != (CodeRange{0, 0, 100, 0}) || got[0].Color.R != 255 {
		t.Fatal("clipping/input ownership", got)
	}
	got[0].Color.R = 0
	if c.Get()[0].Color.R != 255 {
		t.Fatal("output ownership")
	}
	other := ed.Decorations(CodeDecoration{Range: CodeRange{0, 1, 0, 3}, Style: CodeDecorationText})
	c.Clear()
	if len(other.Get()) != 1 {
		t.Fatal("owner isolation")
	}
	rng := rand.New(rand.NewSource(31))
	for i := 0; i < 500; i++ {
		a := rng.Intn(99)
		z := a + 1 + rng.Intn(100-a)
		c.Append(CodeDecoration{Range: CodeRange{a, 1, z, 3}})
	}
	for pass := 0; pass < 2; pass++ {
		for line := 0; line < ed.buf.count(); line++ {
			want := []int{}
			for i, d := range c.entries {
				if d.Range.Line <= line && d.Range.EndLine >= line {
					want = append(want, i)
				}
			}
			if !reflect.DeepEqual(c.atLine(line), want) {
				t.Fatal("interval query", pass, line)
			}
		}
		ed.buf.replace(codePos{20, 0}, codePos{40, 0}, "new\n")
	}
	c.Dispose()
	c.Append(CodeDecoration{Range: CodeRange{0, 0, 0, 2}})
	if len(c.Get()) != 0 || len(ed.decorations) != 1 {
		t.Fatal("dispose")
	}
	other.Dispose()
	if ed.buf.onEdit != nil {
		t.Fatal("last owner hook")
	}
}

func TestCodeLanguageRulesEditing(t *testing.T) {
	ed, h := focusedEditor("")
	rules := CodeLanguageRules{Brackets: []CodePair{{Open: "<!--", Close: "-->"}, {Open: "«", Close: "»"}}}
	if err := ed.SetEditingRules(&rules); err != nil {
		t.Fatal(err)
	}
	for _, ch := range []string{"<", "!", "-", "-"} {
		typeAt(h, ch)
	}
	if ed.Value() != "<!---->" {
		t.Fatal("multi pair", ed.Value())
	}
	for _, ch := range []string{"-", "-", ">"} {
		typeAt(h, ch)
	}
	if _, col := ed.Cursor(); col != 7 || ed.Value() != "<!---->" {
		t.Fatal("skip close", col, ed.Value())
	}
	ed.SetValue("")
	ed.SetCursor(0, 0)
	h.Frame()
	typeAt(h, "<!--")
	h.Key(key.NameDeleteBackward, 0)
	if ed.Value() != "" {
		t.Fatal("empty pair backspace", ed.Value())
	}
	ed.SetValue("猫")
	ed.sels = []codeSel{{codePos{}, codePos{0, 1}}}
	h.Frame()
	typeAt(h, "<!--")
	if ed.Value() != "<!--猫-->" || ed.Selection() != "猫" {
		t.Fatal("asymmetric wrap", ed.Value(), ed.Selection())
	}
	ed.SetValue("")
	h.Frame()
	rules.AutoClosingPairs = []CodePair{}
	if err := ed.SetEditingRules(&rules); err != nil {
		t.Fatal(err)
	}
	typeAt(h, "«")
	if ed.Value() != "«" {
		t.Fatal("explicit empty pairs")
	}
	rules.AutoClosingPairs = []CodePair{{Open: "«", Close: "»", NotIn: []CodeSyntaxContext{CodeSyntaxComment}}}
	ed.SetEditingRules(&rules)
	ed.SyntaxContext(func(int, int) CodeSyntaxContext { return CodeSyntaxComment })
	ed.SetValue("")
	h.Frame()
	typeAt(h, "«")
	if ed.Value() != "«" {
		t.Fatal("custom syntax filter")
	}
}

func TestCodeLanguageRulesRegistryAndIndent(t *testing.T) {
	defer ClearCodeLanguageRules("go")
	ed := CodeEditor("{}").Language("golang")
	rules := CodeLanguageRules{Brackets: []CodePair{{Open: "{", Close: "}"}}, Increase: `\{$`}
	if err := SetCodeLanguageRules("go", rules); err != nil {
		t.Fatal(err)
	}
	rules.Brackets[0].Open = "broken"
	if ed.languageRules().config.Brackets[0].Open != "{" {
		t.Fatal("canonical registration and ownership")
	}
	p := codePos{0, 1}
	rep := ed.newline(codeSel{p, p})
	if rep.text != "\n    \n" || rep.caret != 5 {
		t.Fatalf("regex plus structural decrease: %+v", rep)
	}
	if err := SetCodeLanguageRules("go", CodeLanguageRules{Increase: "["}); err == nil || ed.languageRules().increase == nil {
		t.Fatal("atomic invalid update")
	}
	ed.SmartIndent(false).AutoClose(false).Language("go")
	if ed.newline(codeSel{p, p}).text != "\n" || ed.autoClose {
		t.Fatal("independent preferences")
	}
	ed.SmartIndent(true)
	ed.SetValue("  beginend")
	if err := ed.SetEditingRules(&CodeLanguageRules{Increase: `begin$`, Decrease: `^end`}); err != nil {
		t.Fatal(err)
	}
	p = codePos{0, 7}
	rep = ed.newline(codeSel{p, p})
	if rep.text != "\n      \n  " {
		t.Fatalf("regex indentation: %+v", rep)
	}
	ed.SetEditingRules(nil)
	ed.Language("python").SetValue("if x:")
	p = codePos{0, 5}
	if ed.newline(codeSel{p, p}).text != "\n    " {
		t.Fatal("python default")
	}
}
