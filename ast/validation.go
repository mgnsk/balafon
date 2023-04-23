package ast

import "fmt"

func validateRange(v, minIncl, maxIncl int32) error {
	if v < minIncl || v > maxIncl {
		return fmt.Errorf("value must be in range [%d, %d], got: %d", minIncl, maxIncl, v)
	}
	return nil
}

func validateNoteValue(v int) error {
	if uv := uint8(v); v < 1 || v > 128 || uv&(uv-1) != 0 {
		return fmt.Errorf("note value must be a power of 2 in the range [1, 128], got: %d", v)
	}
	return nil
}

func validateTuplet(v int) error {
	if v == 3 || v == 5 {
		return nil
	}
	return fmt.Errorf("invalid tuplet value, got: %d", v)
}
