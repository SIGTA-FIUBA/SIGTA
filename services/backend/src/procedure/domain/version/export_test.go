package version

import "encoding/json"

// The tests, in package version_test, reach what is not exported through this
// file, which only builds with them.

// ReencodeLegacy decodes a format before 3 as it was stored and encodes it
// again, so that a test can check that legacyContent still describes it.
func ReencodeLegacy(stored []byte) ([]byte, error) {
	var legacy legacyContent
	if err := decodeStrict(stored, &legacy); err != nil {
		return nil, err
	}
	return json.Marshal(legacy)
}
