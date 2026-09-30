package input

import (
	"errors"
	"math"
	"testing"

	"github.com/dyike/keel/native"
)

func TestRejectsBadArguments(t *testing.T) {
	if err := Click(Button(9)); !errors.Is(err, native.ErrInvalidArgument) {
		t.Errorf("Click: %v", err)
	}
	if err := MouseMove(math.NaN(), 0); !errors.Is(err, native.ErrInvalidArgument) {
		t.Errorf("MouseMove: %v", err)
	}
}
