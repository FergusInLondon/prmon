package github

// ChangeCheckableCollection is an interface that a collection is expected
// to implement if it's to be accepted by the ChangeChecker.
type ChangeCheckableCollection interface {
	// VersionKeys provides a slice of strings - where each string is a unique
	// key for that item.
	VersionKeys() []string
}

// ChangeChecker is a simple mechanism for comparing two collections, and
// determining whether or not any changes are present.
type ChangeChecker struct {
	genKeyChecker keyCheckerGen
	keyChecker    keyChecker
}

// A keychecker accepts a slice of strings (keys) and returns a boolean indicating
// whether any changes have been observed. Whereas a keyCheckerGen is a function
// that returns a keyChecker when provided a list of existing keys.
type keyCheckerGen func([]string) keyChecker
type keyChecker func([]string) bool

var (
	// DEFAULT_KEY_CHECKER generates a map of keys, and checks against this map -
	// as a result it is *not sensitive* to the ordering of a list of keys.
	DEFAULT_KEY_CHECKER = func(keyList []string) keyChecker {
		// Simple closure to capture the state of the current batch of keys.
		keyPresentMap := make(map[string]struct{})
		for _, item := range keyList {
			keyPresentMap[item] = struct{}{}
		}

		return func(newKeylist []string) bool {
			// Check One: Are the two Key Lists of the same length?
			if len(newKeylist) != len(keyList) {
				return false
			}

			// Check Two: Iterate through available keys - are they present in the
			// key map we generated during initialisation?
			for _, key := range newKeylist {
				if _, isPresent := keyPresentMap[key]; !isPresent {
					return false
				}
			}

			return true
		}
	}
	// SIMPLE_KEY_CHECKER simply compares two lists, and as such is sensitive
	// to the ordering of the input lists.
	SIMPLE_KEY_CHECKER = func(keylist []string) keyChecker {
		return func(newKeylist []string) bool {
			if len(newKeylist) != len(keylist) {
				return false
			}

			for idx, val := range keylist {
				if val != newKeylist[idx] {
					return false
				}
			}

			return true
		}
	}
)

// NewChangeChecker accepts a ChangeCheckableCollection and initialises a new
// ChangeChecker struct. If a keyCheckerGen is not explicitly provided, then
// the default one will be used - which is a map-based implementation and not
// sensitive to ordering.
func NewChangeChecker(items ChangeCheckableCollection, gen keyCheckerGen) *ChangeChecker {
	if gen == nil {
		gen = DEFAULT_KEY_CHECKER
	}
	return &ChangeChecker{
		genKeyChecker: gen,
		keyChecker:    gen(items.VersionKeys()),
	}
}

// HasChanged gets the key list for a new ChangeCheckableCollection, and then
// runs the keyChecker against it. If a change is detected, then it generates
// a new keyChecker for subsequent runs.
func (cache ChangeChecker) HasChanged(items ChangeCheckableCollection) bool {
	keyList := items.VersionKeys()
	if cache.keyChecker(keyList) {
		// Keys match
		return false
	}

	cache.keyChecker = cache.genKeyChecker(keyList)
	return true
}
