package util

// MustUint returns v or panics if err != nil.
func MustUint(v uint64, err error) uint64 {
	if err != nil {
		panic(err)
	}
	return v
}
