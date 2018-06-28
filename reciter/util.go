package reciter

import "fmt"

var A, X, Y byte

/* Retrieve flags for character at mem59-1 */
func (r *Reciter) getFlag(npos, mask byte) byte {
	X = npos
	return flags36376[r.inputtemp[X]] & mask
}

func (r *Reciter) match(str string) int {
	for i := 0; i < len(str); i++ {
		ch := str[i]
		A = r.inputtemp[X]
		X++
		if A != ch {
			return 0
		}
	}
	return 1
}

func printRule(offset int) {
	i := 1
	var A byte
	fmt.Printf("Applying rule: ")
	for (A & 128) == 0 {
		A = getRuleByte(uint16(offset), byte(i))
		if (A & 127) == '=' {
			fmt.Printf(" -> ")
		} else {
			fmt.Printf("%c", A&127)
		}
		i++
	}
	fmt.Printf("\n")
}

func (r *Reciter) handle_ch2(ch, mem byte) int {
	X = mem
	tmp := flags36376[r.inputtemp[mem]]
	if ch == ' ' {
		if tmp&128 != 0 {
			return 1
		}
	} else if ch == '#' {
		if !(tmp&64 != 0) {
			return 1
		}
	} else if ch == '.' {
		if !(tmp&8 != 0) {
			return 1
		}
	} else if ch == '^' {
		if !(tmp&32 != 0) {
			return 1
		}
	} else {
		return -1
	}
	return 0
}

func (r *Reciter) handle_ch(A, mem byte) int {
	X = mem
	tmp := flags36376[r.inputtemp[X]]
	if A == ' ' {
		if (tmp & 128) != 0 {
			return 1
		}
	} else if A == '#' {
		if (tmp & 64) == 0 {
			return 1
		}
	} else if A == '.' {
		if (tmp & 8) == 0 {
			return 1
		}
	} else if A == '&' {
		if (tmp & 16) == 0 {
			if r.inputtemp[X] != 72 {
				return 1
			}
			X++
		}
	} else if A == '^' {
		if (tmp & 32) == 0 {
			return 1
		}
	} else if A == '+' {
		X = mem
		A = r.inputtemp[X]
		if (A != 69) && (A != 73) && (A != 89) {
			return 1
		}
	} else {
		return -1
	}
	return 0
}

func getRuleByte(mem62 uint16, Y byte) byte {
	var address uint = uint(mem62)
	// fmt.Println(address, uint16(Y))
	if mem62 >= 37541 {
		address -= 37541
		return Rules2[address+uint(Y)]
	}
	address -= 32000
	// fmt.Println(address, uint(Y))
	return Rules[address+uint(Y)]
}
