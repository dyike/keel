package shaping

import (
	"testing"

	"github.com/go-text/typesetting/internal/unicodedata"
	"github.com/go-text/typesetting/segmenter"
)

// segmenterMandatoryBreak is what WrapParagraphF computed before
// hasMandatoryBreak replaced it.
func segmenterMandatoryBreak(text []rune) bool {
	var seg segmenter.Segmenter
	br := newBreaker(&seg, text)
	for {
		option, ok := br.nextWordBreak()
		if !ok {
			return false
		}
		if option.required {
			return true
		}
	}
}

func TestHasMandatoryBreakClasses(t *testing.T) {
	const classes = unicodedata.LB_BK | unicodedata.LB_CR | unicodedata.LB_LF | unicodedata.LB_NL
	for r := rune(0); r <= 0x10FFFF; r++ {
		mandatory := unicodedata.LookupLineBreak(r)&classes != 0
		if got := hasMandatoryBreak([]rune{r, 'a'}); got != mandatory {
			t.Fatalf("%U: hasMandatoryBreak %v, line break class mandatory %v", r, got, mandatory)
		}
	}
}

func TestHasMandatoryBreakMatchesSegmenter(t *testing.T) {
	for _, s := range []string{
		"", "a", "\n", "a\n", "\na", "a\nb", "a\r\nb", "a\r\n", "a\rb", "a\r", "\r\n",
		"a b", "a\u0085b", "a\vb", "a\fb", "a ", "hello world", "a  b\t c",
		"a​b", "x́\ny", "日本語\nテキスト", "مرحبا\nبالعالم", "a\r\r\n", "\n\n",
	} {
		text := []rune(s)
		if got, want := hasMandatoryBreak(text), segmenterMandatoryBreak(text); got != want {
			t.Errorf("%q: hasMandatoryBreak %v, segmenter %v", s, got, want)
		}
	}
}
