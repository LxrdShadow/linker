package progress

import (
	"testing"
)

func TestUpdateProgressBar(t *testing.T) {
	t.Run("progress initial values", func(t *testing.T) {
		progress := NewProgressBar(1000, '#', 1, "", "")

		assertEqual(t, progress.total, uint64(1000), "total")
		assertEqual(t, progress.current, uint64(0), "current value")
		assertEqual(t, progress.percent, uint8(0), "percentage")
		assertEqual(t, progress.char, '#', "character")
	})

	t.Run("progress update from new value", func(t *testing.T) {
		progress := NewProgressBar(1000, '#', 1, "", "")
		progress.NewValueUpdate(200)

		assertEqual(t, progress.total, uint64(1000), "total")
		assertEqual(t, progress.current, uint64(200), "current value")
		assertEqual(t, progress.percent, uint8(20), "percentage")

		progress.NewValueUpdate(300)
		assertEqual(t, progress.percent, uint8(30), "percentage")
	})

	t.Run("progress update from added value", func(t *testing.T) {
		progress := NewProgressBar(1000, '#', 1, "", "")
		progress.AppendUpdate(200)

		assertEqual(t, progress.total, uint64(1000), "total")
		assertEqual(t, progress.current, uint64(200), "current value")
		assertEqual(t, progress.percent, uint8(20), "percentage")

		progress.AppendUpdate(100)
		assertEqual(t, progress.current, uint64(300), "current value")
		assertEqual(t, progress.percent, uint8(30), "percentage")
	})
}

func assertEqual(t *testing.T, got, want any, label string) {
	t.Helper()
	if got != want {
		t.Errorf("%v mismatch: got %v want %v", label, got, want)
	}
}
