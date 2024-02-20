package helper

// StringInArray is value in string array
func StringInArray(value string, lists []string) bool {
	for _, s := range lists {
		if value == s {
			return true
		}
	}
	return false
}
