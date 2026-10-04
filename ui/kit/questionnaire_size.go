package kit

import "github.com/dyike/keel/ui/theme"

// QuestionnaireSize controls all built-in question parts and answer controls.
type QuestionnaireSize uint8

const (
	QuestionnaireSizeMedium QuestionnaireSize = iota
	QuestionnaireSizeXSmall
	QuestionnaireSizeSmall
	QuestionnaireSizeLarge
)

func (v *QuestionnaireView) Size(size QuestionnaireSize) *QuestionnaireView {
	if size <= QuestionnaireSizeLarge {
		v.size = size
	}
	return v
}

type questionnaireMetrics struct {
	title, description, errorText, gap, section, choice, optionFont, star, button float32
	input                                                                         InputSize
}

func (v *QuestionnaireView) sizeMetrics() questionnaireMetrics {
	switch v.size {
	case QuestionnaireSizeXSmall:
		return questionnaireMetrics{theme.TextControl, theme.TextXs, theme.TextXs, 4, 12, 14, theme.TextSm, 16, 24, InputSizeXSmall}
	case QuestionnaireSizeSmall:
		return questionnaireMetrics{theme.TextBody, theme.TextSm, theme.TextXs, 6, 16, 16, theme.TextControl, 18, 28, InputSizeSmall}
	case QuestionnaireSizeLarge:
		return questionnaireMetrics{theme.TextXl, theme.TextBody, theme.TextMd, 12, 24, 22, theme.TextLg, 28, 40, InputSizeLarge}
	default:
		return questionnaireMetrics{theme.TextLg, theme.TextMd, theme.TextSm, 8, 20, 0, 0, 22, 32, InputSizeMedium}
	}
}
func (v *QuestionnaireView) sizeControls(c *questionControl, m questionnaireMetrics) {
	if c.radio != nil {
		c.radio.Size(m.choice).TextSize(m.optionFont)
	}
	for _, check := range c.checks {
		check.Size(m.choice).TextSize(m.optionFont)
	}
	if c.rating != nil {
		c.rating.Size(m.star)
	}
	if c.input != nil {
		c.input.Size(m.input)
	}
	if c.freeform != nil {
		c.freeform.Size(m.input)
	}
}
