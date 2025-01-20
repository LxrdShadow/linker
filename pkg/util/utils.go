package util

func ByteDecodeUnit(num uint64) (string, uint64) {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	base := uint64(1)
	var unit string

	for i, u := range units {
		if num < base*1000 || i == len(units)-1 {
			unit = u
			break
		}
		base *= 1000
	}

	return unit, base
}
