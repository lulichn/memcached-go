package memcache
import "errors"

type HashAlgorithm int
const (
	NATIVE_HASH HashAlgorithm = iota
	// TODO Add algorithm if needed
)

func hash(key string, algorithm HashAlgorithm) (int, error) {
	switch algorithm {
	case NATIVE_HASH:
		return string_hash(key), nil
	default:
		return -1, errors.New("algorithm")
	}
}

func string_hash(key string) int {
	h := 0
	values := []byte(key)
	for _, v := range values {
		h = 31 * h + int(v)
	}
	return h & 0xffffffff
}

