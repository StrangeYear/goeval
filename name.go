package goeval

func indexName(base string, key Value) string {
	return base + "[" + key.String() + "]"
}

func callName(base string) string {
	return base + "()"
}
