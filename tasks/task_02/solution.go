package main

func rotateRunes(s string, shift int) string {
	test := []rune(s)
	if len(test) == 0 {
		return ""
	}

	shift = shift % len(test)

	if shift > -1 {
		return string(append(test[shift:], test[:shift]...))
	}

	index := len(test) - shift*-1
	return string(append(test[index:], test[:index]...))
}
