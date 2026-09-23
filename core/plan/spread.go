package plan

type Spreadable interface {
	IsSpread() bool
}

type StringWrapper struct {
	Value string
}

func (s StringWrapper) IsSpread() bool {
	return s.Value == "..."
}

func Spread[T Spreadable](left []T, right []T) []T {
	if left == nil {
		return right
	}

	result := make([]T, 0, len(left)+len(right))

	for _, val := range left {
		if val.IsSpread() {
			result = append(result, right...)
		} else {
			result = append(result, val)
		}
	}

	return result
}

// expands the input/newList by replacing any "..." with the contents of oldList
func SpreadStrings(newList []string, oldList []string) []string {
	if newList == nil {
		return oldList
	}

	result := make([]string, 0, len(newList)+len(oldList))

	for _, val := range newList {
		if val == "..." {
			result = append(result, oldList...)
		} else {
			result = append(result, val)
		}
	}

	return result
}
