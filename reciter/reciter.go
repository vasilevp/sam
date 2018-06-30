package reciter

import (
	"github.com/exploser/sam/config"
)

type Reciter struct {
	inputtemp [256]byte
}

func (rec *Reciter) TextToPhonemes(input []byte, cfg *config.Config) bool {
	var mem56, //output position for phonemes
		mem58,
		mem60,
		mem61,

		mem64, // position of '=' or current character
		mem65, // position of ')'
		mem66 byte // position of '('
	var mem62 uint16 // memory position of current rule

	rec.inputtemp[0] = ' '

	// secure copy of input
	// because input will be overwritten by phonemes
	X = 0
	for X < 255 {
		A = input[X] & 127
		if A >= 112 {
			A = A & 95
		} else if A >= 96 {
			A = A & 79
		}
		X++
		rec.inputtemp[X] = A
	}
	rec.inputtemp[255] = 27
	mem56 = 255
	mem61 = 255

	var mem57 byte
pos36554:
	for {
		for {
			mem61++
			X = mem61
			mem64 = rec.inputtemp[X]
			if mem64 == '[' {
				mem56++
				X = mem56
				input[X] = 155
				return true
			}

			if mem64 != '.' {
				break
			}
			X++
			A = flags36376[rec.inputtemp[X]] & 1
			if A != 0 {
				break
			}
			mem56++
			X = mem56
			A = '.'
			input[X] = '.'
		}
		mem57 = flags36376[mem64]
		if (mem57 & 2) != 0 {
			mem62 = 37541
			goto pos36700
		}

		if mem57 != 0 {
			break
		}
		rec.inputtemp[X] = ' '
		mem56++
		X = mem56
		if X > 120 {
			input[X] = 155
			return true
		}
		input[X] = 32
	}

	if !(mem57&128 != 0) {
		return false
	}

	// go to the right rules for this character.
	X = mem64 - 'A'
	mem62 = uint16(Tab37489[X]) | uint16(Tab37515[X])<<8

pos36700:
	mem62++
	// find next rule
	for (getRuleByte(mem62, 0) & 128) == 0 {
		mem62++
	}
	var Y byte
	Y++
	for getRuleByte(mem62, Y) != '(' {
		Y++
	}
	mem66 = Y
	Y++
	for getRuleByte(mem62, Y) != ')' {
		Y++
	}
	mem65 = Y
	Y++
	for (getRuleByte(mem62, Y) & 127) != '=' {
		Y++
	}
	mem64 = Y

	mem60 = mem61
	X = mem61
	// compare the string within the bracket
	Y = mem66 + 1

	for {
		if getRuleByte(mem62, Y) != rec.inputtemp[X] {
			goto pos36700
		}
		Y++
		if Y == mem65 {
			break
		}
		X++
		mem60 = X
	}

	// the string in the bracket is correct

	mem59 := mem61

	for {
		for {
			mem66--
			mem57 = getRuleByte(mem62, mem66)
			if (mem57 & 128) != 0 {
				mem58 = mem60
				goto pos37184
			}
			X = mem57 & 127
			if (flags36376[X] & 128) == 0 {
				break
			}
			if rec.inputtemp[mem59-1] != mem57 {
				goto pos36700
			}
			mem59--
		}

		ch := mem57

		r := rec.handle_ch2(ch, mem59-1)
		if r == -1 {
			switch ch {
			case '&':
				if rec.getFlag(mem59-1, 16) == 0 {
					if rec.inputtemp[X] != 'H' {
						r = 1
					} else {
						X--
						A = rec.inputtemp[X]
						if (A != 'C') && (A != 'S') {
							r = 1
						}
					}
				}
				break

			case '@':
				if rec.getFlag(mem59-1, 4) == 0 {
					A = rec.inputtemp[X]
					if A != 72 {
						r = 1
					}
					if (A != 84) && (A != 67) && (A != 83) {
						r = 1
					}
				}
				break
			case '+':
				X = mem59
				X--
				A = rec.inputtemp[X]
				if (A != 'E') && (A != 'I') && (A != 0) {
					r = 1
				}
				break
			case ':':
				for rec.getFlag(mem59-1, 32) != 0 {
					mem59--
				}
				continue
			default:
				return false
			}
		}

		if r == 1 {
			goto pos36700
		}
		mem59 = X
	}

doWhileAEqualsPercent:
	// do ... while(A=='%')
	// for ok := true; ok; ok = (A == '%') {
	X = mem58 + 1
	if rec.inputtemp[X] == 'E' {
		if (flags36376[rec.inputtemp[X+1]] & 128) != 0 {
			X++
			A = rec.inputtemp[X]
			if A == 'L' {
				X++
				if rec.inputtemp[X] != 'Y' {
					goto pos36700
				}
			} else if (A != 'R') && (A != 'S') && (A != 'D') && rec.match("FUL") == 0 {
				goto pos36700
			}
		}
	} else {
		if rec.match("ING") == 0 {
			goto pos36700
		}
		mem58 = X
	}

pos37184:
	r := 0
	// do ... while (r == 0);
	for ok := true; ok; ok = (r == 0) {
		var mem57 byte
		for {
			Y := mem65 + 1
			if Y == mem64 {
				mem61 = mem60

				if cfg.Debug {
					printRule(int(mem62))
				}

				for {
					A = getRuleByte(mem62, Y)
					mem57 = A
					A = A & 127
					if A != '=' {
						mem56++
						input[mem56] = A
					}
					if (mem57 & 128) != 0 {
						goto pos36554
					}
					Y++
				}
			}
			mem65 = Y
			mem57 = getRuleByte(mem62, Y)
			if (flags36376[mem57] & 128) == 0 {
				break
			}
			if rec.inputtemp[mem58+1] != mem57 {
				r = 1
				break
			}
			mem58++
		}

		if r == 0 {
			A = mem57
			if A == '@' {
				if rec.getFlag(mem58+1, 4) == 0 {
					A = rec.inputtemp[X]
					if (A != 82) && (A != 84) &&
						(A != 67) && (A != 83) {
						r = 1
					}
				} else {
					r = -2
				}
			} else if A == ':' {
				for rec.getFlag(mem58+1, 32) != 0 {
					mem58 = X
				}
				r = -2
			} else {
				r = rec.handle_ch(A, mem58+1)
			}
		}

		if r == 1 {
			goto pos36700
		}
		if r == -2 {
			r = 0
			continue
		}
		if r == 0 {
			mem58 = X
		}
	}
	// }
	if A == '%' {
		goto doWhileAEqualsPercent
	}
	return false
}
