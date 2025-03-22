package plan

type Filter struct {
	Include []string `json:"include,omitempty" jsonschema:"description=Files or directories to include"`
	Exclude []string `json:"exclude,omitempty" jsonschema:"description=Files or directories to exclude"`
}

func NewFilter(include []string, exclude []string) Filter {
	return Filter{
		Include: include,
		Exclude: exclude,
	}
}

func NewIncludeFilter(include []string) Filter {
	return Filter{
		Include: include,
	}
}
