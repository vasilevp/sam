package sam

import (
	"fmt"
	"test/sam/global"
	"test/sam/render"
)

var mem39 byte
var mem44 byte
var mem47 byte
var mem49 byte
var mem50 byte
var mem51 byte
var mem53 byte
var mem59 byte

var X byte

var stress [256]byte        //numbers from 0 to 8
var phonemeLength [256]byte //tab40160
var phonemeindex [256]byte

var phonemeIndexOutput [60]byte  //tab47296
var stressOutput [60]byte        //tab47365
var phonemeLengthOutput [60]byte //tab47416

// contains the final soundbuffer

func SetInput(_input string) {
	for i := 0; i < len(_input); i++ {
		global.Input[i] = _input[i]
	}
	global.Input[len(_input)] = 0
}

func SetSpeed(_speed byte)   { global.Speed = _speed }
func SetPitch(_pitch byte)   { global.Pitch = _pitch }
func SetMouth(_mouth byte)   { global.Mouth = _mouth }
func SetThroat(_throat byte) { global.Throat = _throat }
func EnableSingmode()        { global.Singmode = true }
func GetBuffer() []byte      { return global.Buffer }
func GetBufferLength() int   { return global.Bufferpos }

func Init() {
	var i int
	render.SetMouthThroat(global.Mouth, global.Throat)

	global.Bufferpos = 0
	// TODO, check for free the memory, 10 seconds of output should be more than enough
	global.Buffer = make([]byte, 22050*10)

	for i = 0; i < 256; i++ {
		stress[i] = 0
		phonemeLength[i] = 0
	}

	for i = 0; i < 60; i++ {
		phonemeIndexOutput[i] = 0
		stressOutput[i] = 0
		phonemeLengthOutput[i] = 0
	}
	phonemeindex[255] = global.END //to prevent buffer overflow // ML : changed from 32 to 255 to stop freezing with long inputs
}

func SAMMain() int {
	Init()
	/* FIXME: At odds with assignment in Init() */
	phonemeindex[255] = 32 //to prevent buffer overflow

	if !Parser1() {
		return 0
	}
	if global.Debug {
		global.PrintPhonemes(phonemeindex, phonemeLength, stress)
	}
	Parser2()
	CopyStress()
	SetPhonemeLength()
	AdjustLengths()
	Code41240()
	for ok := true; ok; ok = (X != 0) {
		if phonemeindex[X] > 80 {
			phonemeindex[X] = global.END
			break // error: delete all behind it
		}
		X++
	}
	InsertBreath(mem59)

	if global.Debug {
		global.PrintPhonemes(phonemeindex, phonemeLength, stress)
	}

	PrepareOutput()
	return 1
}

func PrepareOutput() {
	var srcpos byte = 0  // Position in source
	var destpos byte = 0 // Position in output

	for {
		A := phonemeindex[srcpos]
		phonemeIndexOutput[destpos] = A
		switch A {
		case global.END:
			render.Render()
			return
		case global.BREAK:
			phonemeIndexOutput[destpos] = global.END
			render.Render()
			destpos = 0
			break
		case 0:
			break
		default:
			phonemeLengthOutput[destpos] = phonemeLength[srcpos]
			stressOutput[destpos] = stress[srcpos]
			destpos++
		}
		srcpos++
	}
}

func InsertBreath(mem59 byte) {
	var mem54 byte = 255
	var len byte = 0
	var index byte //variable Y

	var pos byte = 0

	index = phonemeindex[pos]
	for phonemeindex[pos] != global.END {
		index = phonemeindex[pos]
		len += phonemeLength[pos]
		if len < 232 {
			if index == global.BREAK {
			} else if !(flags[index]&FLAG_PUNCT != 0) {
				if index == 0 {
					mem54 = pos
				}
			} else {
				len = 0
				pos++
				Insert(pos, global.BREAK, mem59, 0)
			}
		} else {
			pos = mem54
			phonemeindex[pos] = 31 // 'Q*' glottal stop
			phonemeLength[pos] = 4
			stress[pos] = 0

			len = 0
			pos++
			Insert(pos, global.BREAK, mem59, 0)
		}
		pos++
	}
}

// Iterates through the phoneme buffer, copying the stress value from
// the following phoneme under the following circumstance:

//     1. The current phoneme is voiced, excluding plosives and fricatives
//     2. The following phoneme is voiced, excluding plosives and fricatives, and
//     3. The following phoneme is stressed
//
//  In those cases, the stress value+1 from the following phoneme is copied.
//
// For example, the word LOITER is represented as LOY5TER, with as stress
// of 5 on the dipthong OY. This routine will copy the stress value of 6 (5+1)
// to the L that precedes it.

func CopyStress() {
	// loop thought all the phonemes to be output
	var pos byte //mem66
	var Y byte
	Y = phonemeindex[pos]
	for phonemeindex[pos] != global.END {
		Y = phonemeindex[pos]
		// if CONSONANT_FLAG set, skip - only vowels get stress
		if flags[Y]&64 != 0 {
			Y = phonemeindex[pos+1]

			// if the following phoneme is the end, or a vowel, skip
			if Y != global.END && (flags[Y]&128) != 0 {
				// get the stress value at the next position
				Y = stress[pos+1]
				if Y != 0 && !(Y&128 != 0) {
					// if next phoneme is stressed, and a VOWEL OR ER
					// copy stress from next phoneme to this one
					stress[pos] = Y + 1
				}
			}
		}

		pos++
	}
}

func Insert(position /*var57*/, mem60, mem59, mem58 byte) {
	var i int
	for i = 253; i >= int(position); i-- { // ML : always keep last safe-guarding 255
		phonemeindex[i+1] = phonemeindex[i]
		phonemeLength[i+1] = phonemeLength[i]
		stress[i+1] = stress[i]
	}

	phonemeindex[position] = mem60
	phonemeLength[position] = mem59
	stress[position] = mem58
	return
}

func full_match(sign1, sign2 byte) int {
	var Y byte = 0
	for ok := true; ok; ok = (Y != 81) {
		// GET FIRST CHARACTER AT POSITION Y IN signInputTable
		// --> should change name to PhonemeNameTable1
		A := signInputTable1[Y]

		if A == sign1 {
			A = signInputTable2[Y]
			// NOT A SPECIAL AND MATCHES SECOND CHARACTER?
			if (A != '*') && (A == sign2) {
				return int(Y)
			}
		}
		Y++
	}
	return -1
}

func wild_match(sign1, sign2 byte) int {
	var Y int = 0
	for ok := true; ok; ok = (Y != 81) {
		if signInputTable2[Y] == '*' {
			if signInputTable1[Y] == sign1 {
				return Y
			}
		}
		Y++
	}
	return -1
}

// The input[] buffer contains a string of phonemes and stress markers along
// the lines of:
//
//     DHAX KAET IHZ AH5GLIY. <0x9B>
//
// The byte 0x9B marks the end of the buffer. Some phonemes are 2 bytes
// long, such as "DH" and "AX". Others are 1 byte long, such as "T" and "Z".
// There are also stress markers, such as "5" and ".".
//
// The first character of the phonemes are stored in the table signInputTable1[].
// The second character of the phonemes are stored in the table signInputTable2[].
// The stress characters are arranged in low to high stress order in stressInputTable[].
//
// The following process is used to parse the input[] buffer:
//
// Repeat until the <0x9B> character is reached:
//
//        First, a search is made for a 2 character match for phonemes that do not
//        end with the '*' (wildcard) character. On a match, the index of the phoneme
//        is added to phonemeIndex[] and the buffer position is advanced 2 bytes.
//
//        If this fails, a search is made for a 1 character match against all
//        phoneme names ending with a '*' (wildcard). If this succeeds, the
//        phoneme is added to phonemeIndex[] and the buffer position is advanced
//        1 byte.
//
//        If this fails, search for a 1 character match in the stressInputTable[].
//        If this succeeds, the stress value is placed in the last stress[] table
//        at the same index of the last added phoneme, and the buffer position is
//        advanced by 1 byte.
//
//        If this fails, return a 0.
//
// On success:
//
//    1. phonemeIndex[] will contain the index of all the phonemes.
//    2. The last index in phonemeIndex[] will be 255.
//    3. stress[] will contain the stress value for each phoneme

// input[] holds the string of phonemes, each two bytes wide
// signInputTable1[] holds the first character of each phoneme
// signInputTable2[] holds te second character of each phoneme
// phonemeIndex[] holds the indexes of the phonemes after parsing input[]
//
// The parser scans through the input[], finding the names of the phonemes
// by searching signInputTable1[] and signInputTable2[]. On a match, it
// copies the index of the phoneme into the phonemeIndexTable[].
//
// The character <0x9B> marks the end of text in input[]. When it is reached,
// the index 255 is placed at the end of the phonemeIndexTable[], and the
// function returns with a 1 indicating success.
func Parser1() bool {
	var i int
	var sign1 byte
	var sign2 byte
	var position byte = 0
	var srcpos byte = 0

	for i = 0; i < 256; i++ {
		stress[i] = 0
	} // Clear the stress table.
	sign1 = global.Input[srcpos]
	for global.Input[srcpos] != 155 { // 155 (\233) is end of line marker
		sign1 = global.Input[srcpos]
		srcpos++
		sign2 = global.Input[srcpos]
		match := full_match(sign1, sign2)
		if match != -1 {
			// Matched both characters (no wildcards)
			phonemeindex[position] = byte(match)
			srcpos++ // Skip the second character of the input as we've matched it
			continue
		}
		match = wild_match(sign1, sign2)
		if match != -1 {
			// Matched just the first character (with second character matching '*'
			phonemeindex[position] = byte(match)
			position++
		} else {
			// Should be a stress character. Search through the
			// stress table backwards.
			match = 8 // End of stress table. FIXME: Don't hardcode.
			for (sign1 != stressInputTable[match]) && (match > 0) {
				match--
			}

			if match == 0 {
				return false
			} // failure

			stress[position-1] = byte(match) // Set stress for prior phoneme
		}
	} //while

	phonemeindex[position] = global.END
	return true
}

//change phonemelength depedendent on stress
func SetPhonemeLength() {
	position := 0
	for phonemeindex[position] != 255 {
		var A = stress[position]
		if (A == 0) || ((A & 128) != 0) {
			phonemeLength[position] = phonemeLengthTable[phonemeindex[position]]
		} else {
			phonemeLength[position] = phonemeStressedLengthTable[phonemeindex[position]]
		}
		position++
	}
}

func Code41240() {
	var pos byte = 0

	for phonemeindex[pos] != global.END {
		var index = phonemeindex[pos]

		if flags[index]&FLAG_STOPCONS != 0 {
			if flags[index]&FLAG_PLOSIVE != 0 {
				var A byte
				X := pos + 1
				for phonemeindex[X] == 0 {
					X++
					A = phonemeindex[X]
				} /* Skip pause */

				if A != global.END {
					if (flags[A]&8 != 0) || (A == 36) || (A == 37) {
						pos++
						continue
					} // '/H' '/X'
				}

			}
			Insert(pos+1, index+1, phonemeLengthTable[index+1], stress[pos])
			Insert(pos+2, index+2, phonemeLengthTable[index+2], stress[pos])
			pos += 2
		}
		pos++
	}
}

func ChangeRule(position, rule, mem60, mem59, stress byte, descr string) {
	if global.Debug {
		fmt.Printf("RULE: %s\n", descr)
	}
	phonemeindex[position] = rule
	Insert(position+1, mem60, mem59, stress)
}

func drule(str string) {
	if global.Debug {
		fmt.Printf("RULE: %s\n", str)
	}
}

func drule_pre(descr string, X byte) {
	drule(descr)
	if global.Debug {
		fmt.Printf("PRE\n")
	}
	if global.Debug {
		fmt.Printf("phoneme %d (%c%c) length %d\n", X, signInputTable1[phonemeindex[X]], signInputTable2[phonemeindex[X]], phonemeLength[X])
	}
}

func drule_post(X byte) {
	if global.Debug {
		fmt.Printf("POST\n")
	}
	if global.Debug {
		fmt.Printf("phoneme %d (%c%c) length %d\n", X, signInputTable1[phonemeindex[X]], signInputTable2[phonemeindex[X]], phonemeLength[X])
	}
}

// Rewrites the phonemes using the following rules:
//
//       <DIPTHONG ENDING WITH WX> -> <DIPTHONG ENDING WITH WX> WX
//       <DIPTHONG NOT ENDING WITH WX> -> <DIPTHONG NOT ENDING WITH WX> YX
//       UL -> AX L
//       UM -> AX M
//       <STRESSED VOWEL> <SILENCE> <STRESSED VOWEL> -> <STRESSED VOWEL> <SILENCE> Q <VOWEL>
//       T R -> CH R
//       D R -> J R
//       <VOWEL> R -> <VOWEL> RX
//       <VOWEL> L -> <VOWEL> LX
//       G S -> G Z
//       K <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> KX <VOWEL OR DIPTHONG NOT ENDING WITH IY>
//       G <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> GX <VOWEL OR DIPTHONG NOT ENDING WITH IY>
//       S P -> S B
//       S T -> S D
//       S K -> S G
//       S KX -> S GX
//       <ALVEOLAR> UW -> <ALVEOLAR> UX
//       CH -> CH CH' (CH requires two phonemes to represent it)
//       J -> J J' (J requires two phonemes to represent it)
//       <UNSTRESSED VOWEL> T <PAUSE> -> <UNSTRESSED VOWEL> DX <PAUSE>
//       <UNSTRESSED VOWEL> D <PAUSE>  -> <UNSTRESSED VOWEL> DX <PAUSE>

func rule_alveolar_uw(X byte) {
	// ALVEOLAR flag set?
	if flags[phonemeindex[X-1]]&FLAG_ALVEOLAR != 0 {
		drule("<ALVEOLAR> UW -> <ALVEOLAR> UX")
		phonemeindex[X] = 16
	}
}

func rule_ch(X, mem59 byte) {
	drule("CH -> CH CH+1")
	Insert(X+1, 43, mem59, stress[X])
}

func rule_j(X, mem59 byte) {
	drule("J -> J J+1")
	Insert(X+1, 45, mem59, stress[X])
}

func rule_g(pos byte) {
	// G <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> GX <VOWEL OR DIPTHONG NOT ENDING WITH IY>
	// Example: GO

	var index = phonemeindex[pos+1]

	// If dipthong ending with YX, move continue processing next phoneme
	if (index != 255) && ((flags[index] & FLAG_DIP_YX) == 0) {
		// replace G with GX and continue processing next phoneme
		drule("G <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> GX <VOWEL OR DIPTHONG NOT ENDING WITH IY>")
		phonemeindex[pos] = 63 // 'GX'
	}
}

func change(pos, val byte, rule string) {
	drule(rule)
	phonemeindex[pos] = val
}

func rule_dipthong(p, pf, pos, mem59 byte) {
	// <DIPTHONG ENDING WITH WX> -> <DIPTHONG ENDING WITH WX> WX
	// <DIPTHONG NOT ENDING WITH WX> -> <DIPTHONG NOT ENDING WITH WX> YX
	// Example: OIL, COW

	// If ends with IY, use YX, else use WX
	A := 20 // 'WX' = 20 'YX' = 21
	if pf&FLAG_DIP_YX != 0 {
		A++
	}

	// Insert at WX or YX following, copying the stress
	if A == 20 {
		drule("insert WX following dipthong NOT ending in IY sound")
	}
	if A == 21 {
		drule("insert YX following dipthong ending in IY sound")
	}
	Insert(pos+1, byte(A), mem59, stress[pos])

	if p == 53 || p == 42 || p == 44 {
		if p == 53 {
			rule_alveolar_uw(pos) // Example: NEW, DEW, SUE, ZOO, THOO, TOO
		} else if p == 42 {
			rule_ch(pos, mem59) // Example: CHEW
		} else if p == 44 {
			rule_j(pos, mem59)
		} // Example: JAY
	}
}

func Parser2() {
	var pos byte = 0 //mem66;
	var p byte

	if global.Debug {
		fmt.Printf("Parser2\n")
	}
	p = phonemeindex[pos]
	for phonemeindex[pos] != global.END {
		p = phonemeindex[pos]
		if global.Debug {
			fmt.Printf("%d: %c%c\n", pos, signInputTable1[p], signInputTable2[p])
		}

		if p == 0 { // Is phoneme pause?
			pos++
			continue
		}

		var pf = flags[p]
		var prior = phonemeindex[pos-1]

		if pf&FLAG_DIPTHONG != 0 {
			rule_dipthong(p, byte(pf), pos, mem59)
		} else if p == 78 {
			ChangeRule(pos, 13, 24, mem59, stress[pos], "UL -> AX L") // Example: MEDDLE
		} else if p == 79 {
			ChangeRule(pos, 13, 27, mem59, stress[pos], "UM -> AX M") // Example: ASTRONOMY
		} else if p == 80 {
			ChangeRule(pos, 13, 28, mem59, stress[pos], "UN -> AX N") // Example: FUNCTION
		} else if (pf&FLAG_VOWEL != 0) && stress[pos] != 0 {
			// RULE:
			//       <STRESSED VOWEL> <SILENCE> <STRESSED VOWEL> -> <STRESSED VOWEL> <SILENCE> Q <VOWEL>
			// EXAMPLE: AWAY EIGHT
			if phonemeindex[pos+1] == 0 { // If following phoneme is a pause, get next
				p = phonemeindex[pos+2]
				if p != global.END && (flags[p]&FLAG_VOWEL != 0) && stress[pos+2] != 0 {
					drule("Insert glottal stop between two stressed vowels with space between them")
					Insert(pos+2, 31, mem59, 0) // 31 = 'Q'
				}
			}
		} else if p == global.PR { // RULES FOR PHONEMES BEFORE R
			if prior == global.PT {
				change(pos-1, 42, "T R -> CH R") // Example: TRACK
			} else if prior == global.PD {
				change(pos-1, 44, "D R -> J R") // Example: DRY
			} else if flags[prior]&FLAG_VOWEL != 0 {
				change(pos, 18, "<VOWEL> R -> <VOWEL> RX")
			} // Example: ART
		} else if p == 24 && (flags[prior]&FLAG_VOWEL != 0) {
			change(pos, 19, "<VOWEL> L -> <VOWEL> LX") // Example: ALL
		} else if prior == 60 && p == 32 { // 'G' 'S'
			// Can't get to fire -
			//       1. The G -> GX rule intervenes
			//       2. Reciter already replaces GS -> GZ
			change(pos, 38, "G S -> G Z")
		} else if p == 60 {
			rule_g(pos)
		} else {
			if p == 72 { // 'K'
				// K <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> KX <VOWEL OR DIPTHONG NOT ENDING WITH IY>
				// Example: COW
				var Y = phonemeindex[pos+1]
				// If at end, replace current phoneme with KX
				if (flags[Y]&FLAG_DIP_YX) == 0 || Y == global.END { // VOWELS AND DIPTHONGS ENDING WITH IY SOUND flag set?
					change(pos, 75, "K <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> KX <VOWEL OR DIPTHONG NOT ENDING WITH IY>")
					p = 75
					pf = flags[p]
				}
			}

			// Replace with softer version?
			if (flags[p]&FLAG_PLOSIVE != 0) && (prior == 32) { // 'S'
				// RULE:
				//      S P -> S B
				//      S T -> S D
				//      S K -> S G
				//      S KX -> S GX
				// Examples: SPY, STY, SKY, SCOWL

				if global.Debug {
					fmt.Printf("RULE: S* %c%c -> S* %c%c\n", signInputTable1[p], signInputTable2[p], signInputTable1[p-12], signInputTable2[p-12])
				}
				phonemeindex[pos] = p - 12
			} else if !(pf&FLAG_PLOSIVE != 0) {
				p = phonemeindex[pos]
				if p == 53 {
					rule_alveolar_uw(pos) // Example: NEW, DEW, SUE, ZOO, THOO, TOO
				} else if p == 42 {
					rule_ch(pos, mem59) // Example: CHEW
				} else if p == 44 {
					rule_j(pos, mem59) // Example: JAY
				}
			}

			if p == 69 || p == 57 { // 'T', 'D'
				// RULE: Soften T following vowel
				// NOTE: This rule fails for cases such as "ODD"
				//       <UNSTRESSED VOWEL> T <PAUSE> -> <UNSTRESSED VOWEL> DX <PAUSE>
				//       <UNSTRESSED VOWEL> D <PAUSE>  -> <UNSTRESSED VOWEL> DX <PAUSE>
				// Example: PARTY, TARDY
				if flags[phonemeindex[pos-1]]&FLAG_VOWEL != 0 {
					p = phonemeindex[pos+1]
					if p == 0 {
						p = phonemeindex[pos+2]
					}
					if (flags[p]&FLAG_VOWEL != 0) && stress[pos+1] == 0 {
						change(pos, 30, "Soften T or D following vowel or ER and preceding a pause -> DX")
					}
				}
			}
		}
		pos++
	} // while
}

// Applies various rules that adjust the lengths of phonemes
//
//         Lengthen <FRICATIVE> or <VOICED> between <VOWEL> and <PUNCTUATION> by 1.5
//         <VOWEL> <RX | LX> <CONSONANT> - decrease <VOWEL> length by 1
//         <VOWEL> <UNVOICED PLOSIVE> - decrease vowel by 1/8th
//         <VOWEL> <UNVOICED CONSONANT> - increase vowel by 1/2 + 1
//         <NASAL> <STOP CONSONANT> - set nasal = 5, consonant = 6
//         <VOICED STOP CONSONANT> {optional silence} <STOP CONSONANT> - shorten both to 1/2 + 1
//         <LIQUID CONSONANT> <DIPTHONG> - decrease by 2
//
func AdjustLengths() {
	// LENGTHEN VOWELS PRECEDING PUNCTUATION
	//
	// Search for punctuation. If found, back up to the first vowel, then
	// process all phonemes between there and up to (but not including) the punctuation.
	// If any phoneme is found that is a either a fricative or voiced, the duration is
	// increased by (length * 1.5) + 1

	// loop index
	var X byte = 0
	var index byte
	index = phonemeindex[X]
	for phonemeindex[X] != global.END {
		index = phonemeindex[X]
		// not punctuation?
		if (flags[index] & FLAG_PUNCT) == 0 {
			X++
			continue
		}

		var loopIndex byte = X

		X--
		for X != 0 && !(flags[phonemeindex[X]]&FLAG_VOWEL != 0) {
			X--
		} // back up while not a vowel
		if X == 0 {
			break
		}

		for ok := true; ok; ok = (X != loopIndex) {
			// test for vowel
			index = phonemeindex[X]

			// test for fricative/unvoiced or not voiced
			if !(flags[index]&FLAG_FRICATIVE != 0) || (flags[index]&FLAG_VOICED != 0) { //nochmal �berpr�fen
				var A = phonemeLength[X]
				// change phoneme length to (length * 1.5) + 1
				drule_pre("Lengthen <FRICATIVE> or <VOICED> between <VOWEL> and <PUNCTUATION> by 1.5", X)
				phonemeLength[X] = (A >> 1) + A + 1
				drule_post(X)
			}
			X++
		}
		X++
	} // while

	// Similar to the above routine, but shorten vowels under some circumstances

	// Loop throught all phonemes
	var loopIndex byte = 0
	index = phonemeindex[loopIndex]
	for phonemeindex[loopIndex] != global.END {
		index = phonemeindex[loopIndex]
		X := loopIndex

		if flags[index]&FLAG_VOWEL != 0 {
			index = phonemeindex[loopIndex+1]
			if !(flags[index]&FLAG_CONSONANT != 0) {
				if (index == 18) || (index == 19) { // 'RX', 'LX'
					index = phonemeindex[loopIndex+2]
					if flags[index]&FLAG_CONSONANT != 0 {
						drule_pre("<VOWEL> <RX | LX> <CONSONANT> - decrease length of vowel by 1\n", loopIndex)
						phonemeLength[loopIndex]--
						drule_post(loopIndex)
					}
				}
			} else { // Got here if not <VOWEL>
				var flag = flags[index] // 65 if end marker
				if index == global.END {
					flag = 65
				}

				if !(flag&FLAG_VOICED != 0) { // Unvoiced
					// *, .*, ?*, ,*, -*, DX, S*, SH, F*, TH, /H, /X, CH, P*, T*, K*, KX
					if flag&FLAG_PLOSIVE != 0 { // unvoiced plosive
						// RULE: <VOWEL> <UNVOICED PLOSIVE>
						// <VOWEL> <P*, T*, K*, KX>
						drule_pre("<VOWEL> <UNVOICED PLOSIVE> - decrease vowel by 1/8th", loopIndex)
						phonemeLength[loopIndex] -= (phonemeLength[loopIndex] >> 3)
						drule_post(loopIndex)
					}
				} else {
					drule_pre("<VOWEL> <VOICED CONSONANT> - increase vowel by 1/2 + 1\n", X-1)
					// decrease length
					var A = phonemeLength[loopIndex]
					phonemeLength[loopIndex] = (A >> 2) + A + 1 // 5/4*A + 1
					drule_post(loopIndex)
				}
			}
		} else if (flags[index] & FLAG_NASAL) != 0 { // nasal?
			// RULE: <NASAL> <STOP CONSONANT>
			//       Set punctuation length to 6
			//       Set stop consonant length to 5
			X++
			index = phonemeindex[X]
			if index != global.END && (flags[index]&FLAG_STOPCONS != 0) {
				drule("<NASAL> <STOP CONSONANT> - set nasal = 5, consonant = 6")
				phonemeLength[X] = 6   // set stop consonant length to 6
				phonemeLength[X-1] = 5 // set nasal length to 5
			}
		} else if flags[index]&FLAG_STOPCONS != 0 { // (voiced) stop consonant?
			// RULE: <VOICED STOP CONSONANT> {optional silence} <STOP CONSONANT>
			//       Shorten both to (length/2 + 1)

			// move past silence
			X++
			index = phonemeindex[X]
			for phonemeindex[X] == 0 {
				X++
				index = phonemeindex[X]
			}

			if index != global.END && (flags[index]&FLAG_STOPCONS != 0) {
				// FIXME, this looks wrong?
				// RULE: <UNVOICED STOP CONSONANT> {optional silence} <STOP CONSONANT>
				drule("<UNVOICED STOP CONSONANT> {optional silence} <STOP CONSONANT> - shorten both to 1/2 + 1")
				phonemeLength[X] = (phonemeLength[X] >> 1) + 1
				phonemeLength[loopIndex] = (phonemeLength[loopIndex] >> 1) + 1
				X = loopIndex
			}
		} else if flags[index]&FLAG_LIQUIC != 0 { // liquic consonant?
			// RULE: <VOICED NON-VOWEL> <DIPTHONG>
			//       Decrease <DIPTHONG> by 2
			index = phonemeindex[X-1] // prior phoneme;

			// FIXME: The global.Debug code here breaks the rule.
			// prior phoneme a stop consonant>
			if (flags[index] & FLAG_STOPCONS) != 0 {
				drule_pre("<LIQUID CONSONANT> <DIPTHONG> - decrease by 2", X)
			}

			phonemeLength[X] -= 2 // 20ms
			drule_post(X)
		}

		loopIndex++
	}
}
