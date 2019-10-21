package songtitle

func deleteStringSlice(slice []string, index int) []string {
	if index < len(slice)-1 {
		copy(slice[index:], slice[index+1:])
	}

	// not sure if this is necessary
	slice[len(slice)-1] = ""
	return slice[:len(slice)-1]
}

func mapStringSlice(slice []string, f func(string) string) {
	for i, s := range slice {
		slice[i] = f(s)
	}
}
