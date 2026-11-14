package kit

import (
	"github.com/dyike/keel/ui/locale"
	"slices"
	"strings"
	"sync"
)

// EmojiEntry describes one selectable base and its optional five uniform skin
// tones, ordered light through dark. Strings are exact output sequences.
// Mixed-tone combinations may be supplied as independent entries.
type EmojiEntry struct {
	Emoji, Name, Group string
	Keywords           []string
	SkinTones          [5]string
}

var emojiCatalogOnce sync.Once
var emojiCatalog []EmojiEntry

func defaultEmojiCatalog() []EmojiEntry {
	emojiCatalogOnce.Do(func() {
		for _, line := range strings.Split(emojiCatalogData, "\n") {
			fields := strings.Split(line, "\t")
			e := EmojiEntry{Emoji: fields[0], Name: fields[1], Group: fields[2]}
			copy(e.SkinTones[:], fields[3:])
			emojiCatalog = append(emojiCatalog, e)
		}
	})
	return emojiCatalog
}

// DefaultEmojis returns an independent copy of the Unicode 17.0 catalog.
// It contains every fully-qualified sequence, either as a base or skin tone.
// Names and groups are Unicode English names; applications can replace them.
func DefaultEmojis() []EmojiEntry { return cloneEmojiEntries(defaultEmojiCatalog()) }

func cloneEmojiEntries(entries []EmojiEntry) []EmojiEntry {
	out := slices.Clone(entries)
	for i := range out {
		out[i].Keywords = slices.Clone(out[i].Keywords)
	}
	return out
}

func emojiOutput(entry EmojiEntry, tone int) string {
	if tone > 0 && tone <= 5 && entry.SkinTones[tone-1] != "" {
		return entry.SkinTones[tone-1]
	}
	return entry.Emoji
}

func emojiGroupLabel(group string) (symbol, name string) {
	groups := [9]string{"Smileys & Emotion", "People & Body", "Animals & Nature", "Food & Drink", "Travel & Places", "Activities", "Objects", "Symbols", "Flags"}
	symbols := [9]string{"😀", "👋", "🐻", "🍔", "🚗", "⚽", "💡", "🔣", "🏁"}
	for i, s := range groups {
		if s == group {
			return symbols[i], locale.Current().EmojiGroups[i]
		}
	}
	return group, group
}
