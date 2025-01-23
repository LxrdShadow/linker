package progress

import (
	"fmt"
	"time"

	"github.com/LxrdShadow/linker/pkg/color"
)

type ProgressBar struct {
	current, total, denom uint64
	percent               uint8
	char                  rune
	prefix, repr, unit    string
	start                 time.Time
}

func NewProgressBar(total uint64, char rune, denominator uint64, prefix, unit string) *ProgressBar {
	progress := &ProgressBar{
		current: 0,
		total:   total,
		char:    char,
		denom:   denominator,
		prefix:  prefix,
		unit:    unit,
		repr:    "",
		start:   time.Now(),
	}

	return progress
}

func (progress *ProgressBar) NewValueUpdate(current uint64) {
	progress.current = current
	progress.update()
}

func (progress *ProgressBar) AppendUpdate(value uint64) {
	progress.current += value
	progress.update()
}

func (progress *ProgressBar) Render() {
	fmt.Printf("\r%s %s %s %s %s", progress.prefix, progress.representation(), progress.percentage(), progress.progress(), progress.speed())
}

func (progress *ProgressBar) Finish() {
	progress.current = progress.total
	progress.update()
	fmt.Println()
}

func (progress *ProgressBar) update() {
	if progress.current >= progress.total {
		progress.current = progress.total
	}

	progress.percent = progress.getPercentage()
	progress.repr = progress.getRepresentation()
	progress.Render()
}

func (progress *ProgressBar) getPercentage() uint8 {
	return uint8(float64(progress.current) / float64(progress.total) * 100)
}

func (progress *ProgressBar) getRepresentation() string {
	repr := ""
	for range int(progress.percent / 2) {
		repr += string(progress.char)
	}
	repr += ">"

	return repr
}

func (progress *ProgressBar) divideByDenom(value uint64) float32 {
	return float32(float64(value) / float64(progress.denom))
}

func (progress *ProgressBar) percentage() string {
	var percentColor int

	if progress.percent <= 33 {
		percentColor = color.RED
	} else if progress.percent <= 66 {
		percentColor = color.YELLOW
	} else if progress.percent <= 99 {
		percentColor = color.CYAN
	} else {
		percentColor = color.GREEN
	}

	return color.Sprint(percentColor, fmt.Sprintf("%3d%%", progress.percent))
}

func (progress *ProgressBar) representation() string {
	var reprColor int

	if progress.percent <= 33 {
		reprColor = color.RED
	} else if progress.percent <= 66 {
		reprColor = color.YELLOW
	} else if progress.percent <= 99 {
		reprColor = color.CYAN
	} else {
		reprColor = color.GREEN
	}

	return color.Sprint(reprColor, fmt.Sprintf("[%-50s]", progress.repr))
}

func (progress *ProgressBar) progress() string {
	return fmt.Sprintf("%8.2f%s/%.2f%s", progress.divideByDenom(progress.current), progress.unit, progress.divideByDenom(progress.total), progress.unit)
}

func (progress *ProgressBar) speed() string {
	elapsed := time.Since(progress.start).Seconds()
	if elapsed == 0 {
		return fmt.Sprintf("%7.2f%s/s", 0.0, progress.unit)
	}
	speed := float32(progress.current) / float32(elapsed)

	units := []string{"B", "KB", "MB", "GB"}
	unitIdx := 0

	for speed >= 1000 && unitIdx < len(units)-1 {
		speed /= 1000
		unitIdx++
	}

	return color.Sprintf(color.GRAY, "%7.2f%s/s", speed, units[unitIdx])
}
