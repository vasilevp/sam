package sammain

import (
	"fmt"

	"github.com/exploser/sam/config"
	"github.com/exploser/sam/render"
)

type Sam struct {
	stress        [256]byte //numbers from 0 to 8
	phonemeLength [256]byte //tab40160
	phonemeindex  [256]byte

	Input [256]byte //tab39445
	//standard sam sound
	Config *config.Config
}

// contains the final soundbuffer

func (s *Sam) SetInput(input [256]byte) {
	copy(s.Input[:], input[:])
}

func (s *Sam) Init() {
	var i int
	render.SetMouthThroat(s.Config.Mouth, s.Config.Throat)

	for i = 0; i < 256; i++ {
		s.stress[i] = 0
		s.phonemeLength[i] = 0
	}

	s.phonemeindex[255] = render.PhonemeEnd //to prevent buffer overflow // ML : changed from 32 to 255 to stop freezing with long inputs
}

func (s *Sam) SAMMain() bool {
	s.Init()
	/* FIXME: At odds with assignment in Init() */
	s.phonemeindex[255] = 32 //to prevent buffer overflow

	if !s.Parser1() {
		return false
	}
	if s.Config.Debug {
		PrintPhonemes(s.phonemeindex[:], s.phonemeLength[:], s.stress[:])
	}
	s.Parser2()
	s.CopyStress()
	s.SetPhonemeLength()
	s.AdjustLengths()
	s.Code41240()

	var X byte
	for ok := true; ok; ok = (X != 0) {
		if s.phonemeindex[X] > 80 {
			s.phonemeindex[X] = render.PhonemeEnd
			break // error: delete all behind it
		}
		X++
	}
	s.InsertBreath(0)

	if s.Config.Debug {
		PrintPhonemes(s.phonemeindex[:], s.phonemeLength[:], s.stress[:])
	}

	// s.PrepareOutput(&s.r)
	return true
}

func (s *Sam) PrepareOutput(r *render.Render) {
	var srcpos byte = 0  // Position in source
	var destpos byte = 0 // Position in output

	for {
		A := s.phonemeindex[srcpos]
		r.PhonemeIndexOutput[destpos] = A
		switch A {
		case render.PhonemeEnd:
			r.Render(s.Config)
			return
		case BREAK:
			r.PhonemeIndexOutput[destpos] = render.PhonemeEnd
			r.Render(s.Config)
			destpos = 0
			break
		case 0:
			break
		default:
			r.PhonemeLengthOutput[destpos] = s.phonemeLength[srcpos]
			r.StressOutput[destpos] = s.stress[srcpos]
			destpos++
		}
		srcpos++
	}
}

func (s *Sam) InsertBreath(mem59 byte) {
	var mem54 byte = 255
	var len byte
	var index byte //variable Y

	var pos byte

	index = s.phonemeindex[pos]
	for s.phonemeindex[pos] != render.PhonemeEnd {
		index = s.phonemeindex[pos]
		len += s.phonemeLength[pos]
		if len < 232 {
			if index == BREAK {
			} else if !(flags[index]&FLAG_PUNCT != 0) {
				if index == 0 {
					mem54 = pos
				}
			} else {
				len = 0
				pos++
				s.Insert(pos, BREAK, mem59, 0)
			}
		} else {
			pos = mem54
			s.phonemeindex[pos] = 31 // 'Q*' glottal stop
			s.phonemeLength[pos] = 4
			s.stress[pos] = 0

			len = 0
			pos++
			s.Insert(pos, BREAK, mem59, 0)
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

func (s *Sam) CopyStress() {
	// loop thought all the phonemes to be output
	var pos byte //mem66
	var Y byte
	Y = s.phonemeindex[pos]
	for s.phonemeindex[pos] != render.PhonemeEnd {
		Y = s.phonemeindex[pos]
		// if CONSONANT_FLAG set, skip - only vowels get stress
		if flags[Y]&64 != 0 {
			Y = s.phonemeindex[pos+1]

			// if the following phoneme is the end, or a vowel, skip
			if Y != render.PhonemeEnd && (flags[Y]&128) != 0 {
				// get the stress value at the next position
				Y = s.stress[pos+1]
				if Y != 0 && !(Y&128 != 0) {
					// if next phoneme is stressed, and a VOWEL OR ER
					// copy stress from next phoneme to this one
					s.stress[pos] = Y + 1
				}
			}
		}

		pos++
	}
}

func (s *Sam) Insert(position /*var57*/, mem60, mem59, mem58 byte) {
	var i int
	for i = 253; i >= int(position); i-- { // ML : always keep last safe-guarding 255
		s.phonemeindex[i+1] = s.phonemeindex[i]
		s.phonemeLength[i+1] = s.phonemeLength[i]
		s.stress[i+1] = s.stress[i]
	}

	s.phonemeindex[position] = mem60
	s.phonemeLength[position] = mem59
	s.stress[position] = mem58
	return
}

func full_match(sign1, sign2 byte) int {
	var Y byte
	for ok := true; ok; ok = (Y != 81) {
		// GET FIRST CHARACTER AT POSITION Y IN signInputTable
		// --> should change name to PhonemeNameTable1
		A := SignInputTable1[Y]

		if A == sign1 {
			A = SignInputTable2[Y]
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
	var Y int
	for ok := true; ok; ok = (Y != 81) {
		if SignInputTable2[Y] == '*' {
			if SignInputTable1[Y] == sign1 {
				return Y
			}
		}
		Y++
	}
	return -1
}

func PrintPhonemes(phonemeindex []byte, phonemeLength []byte, stress []byte) {
	i := 0
	fmt.Printf("===========================================\n")

	fmt.Printf("Internal Phoneme presentation:\n\n")
	fmt.Printf(" idx    phoneme  length  stress\n")
	fmt.Printf("------------------------------\n")

	for (phonemeindex[i] != render.PhonemeEnd) && (i < 255) {
		if phonemeindex[i] < 81 {
			fmt.Printf(" %3v      %c%c      %3v       %v\n",
				phonemeindex[i],
				SignInputTable1[phonemeindex[i]],
				SignInputTable2[phonemeindex[i]],
				phonemeLength[i],
				stress[i],
			)
		} else {
			fmt.Printf(" %3v      ??      %3v       %v\n", phonemeindex[i], phonemeLength[i], stress[i])
		}
		i++
	}
	fmt.Printf("===========================================\n")
	fmt.Printf("\n")
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
// The first character of the phonemes are stored in the table SignInputTable1[].
// The second character of the phonemes are stored in the table SignInputTable2[].
// The stress characters are arranged in low to high stress order in stressInputTable[].
//
// The following process is used to parse the input[] buffer:
//
// Repeat until the <0x9B> character is reached:
//
//        First, a search is made for a 2 character match for phonemes that do not
//        end with the '*' (wildcard) character. On a match, the index of the phoneme
//        is added tos.phonemeindex[] and the buffer position is advanced 2 bytes.
//
//        If this fails, a search is made for a 1 character match against all
//        phoneme names ending with a '*' (wildcard). If this succeeds, the
//        phoneme is added tos.phonemeindex[] and the buffer position is advanced
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
//    1.s.phonemeindex[] will contain the index of all the phonemes.
//    2. The last index ins.phonemeindex[] will be 255.
//    3. stress[] will contain the stress value for each phoneme

// input[] holds the string of phonemes, each two bytes wide
// SignInputTable1[] holds the first character of each phoneme
// SignInputTable2[] holds te second character of each phoneme
//s.phonemeindex[] holds the indexes of the phonemes after parsing input[]
//
// The parser scans through the input[], finding the names of the phonemes
// by searching SignInputTable1[] and SignInputTable2[]. On a match, it
// copies the index of the phoneme into thes.phonemeindexTable[].
//
// The character <0x9B> marks the end of text in input[]. When it is reached,
// the index 255 is placed at the end of thes.phonemeindexTable[], and the
// function returns with a 1 indicating success.
func (s *Sam) Parser1() bool {
	for i := 0; i < 256; i++ {
		s.stress[i] = 0
	} // Clear the stress table.

	var position byte
	var srcpos byte
	sign1 := s.Input[srcpos]
	var sign2 byte
	for s.Input[srcpos] != 155 { // 155 (\233) is end of line marker
		sign1 = s.Input[srcpos]
		srcpos++
		sign2 = s.Input[srcpos]
		match := full_match(sign1, sign2)
		if match != -1 {
			// Matched both characters (no wildcards)
			s.phonemeindex[position] = byte(match)
			position++
			srcpos++ // Skip the second character of the input as we've matched it
			continue
		}
		match = wild_match(sign1, sign2)
		if match != -1 {
			// Matched just the first character (with second character matching '*'
			s.phonemeindex[position] = byte(match)
			position++
		} else {
			// Should be a stress character. Search through the
			// stress table backwards.
			match = 8 // End of stress table. FIXME: Don't hardcode.
			for (sign1 != StressInputTable[match]) && (match > 0) {
				match--
			}

			if match == 0 {
				return false
			} // failure

			s.stress[position-1] = byte(match) // Set stress for prior phoneme
		}
	} //while

	s.phonemeindex[position] = render.PhonemeEnd
	return true
}

//change phonemelength depedendent on stress
func (s *Sam) SetPhonemeLength() {
	position := 0
	for s.phonemeindex[position] != 255 {
		var A = s.stress[position]
		if (A == 0) || ((A & 128) != 0) {
			s.phonemeLength[position] = phonemeLengthTable[s.phonemeindex[position]]
		} else {
			s.phonemeLength[position] = phonemeStressedLengthTable[s.phonemeindex[position]]
		}
		position++
	}
}

func (s *Sam) Code41240() {
	var pos byte = 0

	for s.phonemeindex[pos] != render.PhonemeEnd {
		var index = s.phonemeindex[pos]

		if flags[index]&FLAG_STOPCONS != 0 {
			if flags[index]&FLAG_PLOSIVE != 0 {
				var A byte
				X := pos + 1
				for s.phonemeindex[X] == 0 {
					X++
					A = s.phonemeindex[X]
				} /* Skip pause */

				if A != render.PhonemeEnd {
					if (flags[A]&8 != 0) || (A == 36) || (A == 37) {
						pos++
						continue
					} // '/H' '/X'
				}

			}
			s.Insert(pos+1, index+1, phonemeLengthTable[index+1], s.stress[pos])
			s.Insert(pos+2, index+2, phonemeLengthTable[index+2], s.stress[pos])
			pos += 2
		}
		pos++
	}
}

func (s *Sam) ChangeRule(position, rule, mem60, mem59, stress byte, descr string) {
	if s.Config.Debug {
		fmt.Printf("RULE: %s\n", descr)
	}
	s.phonemeindex[position] = rule
	s.Insert(position+1, mem60, mem59, stress)
}

func (s *Sam) drule(str string) {
	if s.Config.Debug {
		fmt.Printf("RULE: %s\n", str)
	}
}

func (s *Sam) drule_pre(descr string, X byte) {
	s.drule(descr)
	if s.Config.Debug {
		fmt.Printf("PRE\n")
	}
	if s.Config.Debug {
		fmt.Printf("phoneme %d (%c%c) length %d\n", X, SignInputTable1[s.phonemeindex[X]], SignInputTable2[s.phonemeindex[X]], s.phonemeLength[X])
	}
}

func (s *Sam) drule_post(X byte) {
	if s.Config.Debug {
		fmt.Printf("POST\n")
	}
	if s.Config.Debug {
		fmt.Printf("phoneme %d (%c%c) length %d\n", X, SignInputTable1[s.phonemeindex[X]], SignInputTable2[s.phonemeindex[X]], s.phonemeLength[X])
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

func (s *Sam) rule_alveolar_uw(X byte) {
	// ALVEOLAR flag set?
	if flags[s.phonemeindex[X-1]]&FLAG_ALVEOLAR != 0 {
		s.drule("<ALVEOLAR> UW -> <ALVEOLAR> UX")
		s.phonemeindex[X] = 16
	}
}

func (s *Sam) rule_ch(X, mem59 byte) {
	s.drule("CH -> CH CH+1")
	s.Insert(X+1, 43, mem59, s.stress[X])
}

func (s *Sam) rule_j(X, mem59 byte) {
	s.drule("J -> J J+1")
	s.Insert(X+1, 45, mem59, s.stress[X])
}

func (s *Sam) rule_g(pos byte) {
	// G <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> GX <VOWEL OR DIPTHONG NOT ENDING WITH IY>
	// Example: GO

	var index = s.phonemeindex[pos+1]

	// If dipthong ending with YX, move continue processing next phoneme
	if (index != 255) && ((flags[index] & FLAG_DIP_YX) == 0) {
		// replace G with GX and continue processing next phoneme
		s.drule("G <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> GX <VOWEL OR DIPTHONG NOT ENDING WITH IY>")
		s.phonemeindex[pos] = 63 // 'GX'
	}
}

func (s *Sam) change(pos, val byte, rule string) {
	s.drule(rule)
	s.phonemeindex[pos] = val
}

func (s *Sam) rule_dipthong(p, pf, pos, mem59 byte) {
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
		s.drule("insert WX following dipthong NOT ending in IY sound")
	}
	if A == 21 {
		s.drule("insert YX following dipthong ending in IY sound")
	}
	s.Insert(pos+1, byte(A), mem59, s.stress[pos])

	if p == 53 || p == 42 || p == 44 {
		if p == 53 {
			s.rule_alveolar_uw(pos) // Example: NEW, DEW, SUE, ZOO, THOO, TOO
		} else if p == 42 {
			s.rule_ch(pos, mem59) // Example: CHEW
		} else if p == 44 {
			s.rule_j(pos, mem59)
		} // Example: JAY
	}
}

func (s *Sam) Parser2() {
	var pos byte = 0 //mem66;
	var p byte

	if s.Config.Debug {
		fmt.Printf("Parser2\n")
	}
	p = s.phonemeindex[pos]
	for s.phonemeindex[pos] != render.PhonemeEnd {
		p = s.phonemeindex[pos]
		if s.Config.Debug {
			fmt.Printf("%d: %c%c\n", pos, SignInputTable1[p], SignInputTable2[p])
		}

		if p == 0 { // Is phoneme pause?
			pos++
			continue
		}

		var pf = flags[p]
		var prior = s.phonemeindex[pos-1]

		if pf&FLAG_DIPTHONG != 0 {
			s.rule_dipthong(p, byte(pf), pos, 0)
		} else if p == 78 {
			s.ChangeRule(pos, 13, 24, 0, s.stress[pos], "UL -> AX L") // Example: MEDDLE
		} else if p == 79 {
			s.ChangeRule(pos, 13, 27, 0, s.stress[pos], "UM -> AX M") // Example: ASTRONOMY
		} else if p == 80 {
			s.ChangeRule(pos, 13, 28, 0, s.stress[pos], "UN -> AX N") // Example: FUNCTION
		} else if (pf&FLAG_VOWEL != 0) && s.stress[pos] != 0 {
			// RULE:
			//       <STRESSED VOWEL> <SILENCE> <STRESSED VOWEL> -> <STRESSED VOWEL> <SILENCE> Q <VOWEL>
			// EXAMPLE: AWAY EIGHT
			if s.phonemeindex[pos+1] == 0 { // If following phoneme is a pause, get next
				p = s.phonemeindex[pos+2]
				if p != render.PhonemeEnd && (flags[p]&FLAG_VOWEL != 0) && s.stress[pos+2] != 0 {
					s.drule("Insert glottal stop between two stressed vowels with space between them")
					s.Insert(pos+2, 31, 0, 0) // 31 = 'Q'
				}
			}
		} else if p == pR { // RULES FOR PHONEMES BEFORE R
			if prior == pT {
				s.change(pos-1, 42, "T R -> CH R") // Example: TRACK
			} else if prior == pD {
				s.change(pos-1, 44, "D R -> J R") // Example: DRY
			} else if flags[prior]&FLAG_VOWEL != 0 {
				s.change(pos, 18, "<VOWEL> R -> <VOWEL> RX")
			} // Example: ART
		} else if p == 24 && (flags[prior]&FLAG_VOWEL != 0) {
			s.change(pos, 19, "<VOWEL> L -> <VOWEL> LX") // Example: ALL
		} else if prior == 60 && p == 32 { // 'G' 'S'
			// Can't get to fire -
			//       1. The G -> GX rule intervenes
			//       2. Reciter already replaces GS -> GZ
			s.change(pos, 38, "G S -> G Z")
		} else if p == 60 {
			s.rule_g(pos)
		} else {
			if p == 72 { // 'K'
				// K <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> KX <VOWEL OR DIPTHONG NOT ENDING WITH IY>
				// Example: COW
				var Y = s.phonemeindex[pos+1]
				// If at end, replace current phoneme with KX
				if Y == render.PhonemeEnd || (flags[Y]&FLAG_DIP_YX) == 0 { // VOWELS AND DIPTHONGS ENDING WITH IY SOUND flag set?
					s.change(pos, 75, "K <VOWEL OR DIPTHONG NOT ENDING WITH IY> -> KX <VOWEL OR DIPTHONG NOT ENDING WITH IY>")
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

				if s.Config.Debug {
					fmt.Printf("RULE: S* %c%c -> S* %c%c\n", SignInputTable1[p], SignInputTable2[p], SignInputTable1[p-12], SignInputTable2[p-12])
				}
				s.phonemeindex[pos] = p - 12
			} else if !(pf&FLAG_PLOSIVE != 0) {
				p = s.phonemeindex[pos]
				if p == 53 {
					s.rule_alveolar_uw(pos) // Example: NEW, DEW, SUE, ZOO, THOO, TOO
				} else if p == 42 {
					s.rule_ch(pos, 0) // Example: CHEW
				} else if p == 44 {
					s.rule_j(pos, 0) // Example: JAY
				}
			}

			if p == 69 || p == 57 { // 'T', 'D'
				// RULE: Soften T following vowel
				// NOTE: This rule fails for cases such as "ODD"
				//       <UNSTRESSED VOWEL> T <PAUSE> -> <UNSTRESSED VOWEL> DX <PAUSE>
				//       <UNSTRESSED VOWEL> D <PAUSE>  -> <UNSTRESSED VOWEL> DX <PAUSE>
				// Example: PARTY, TARDY
				if flags[s.phonemeindex[pos-1]]&FLAG_VOWEL != 0 {
					p = s.phonemeindex[pos+1]
					if p == 0 {
						p = s.phonemeindex[pos+2]
					}
					if p != render.PhonemeEnd && (flags[p]&FLAG_VOWEL != 0) && s.stress[pos+1] == 0 {
						s.change(pos, 30, "Soften T or D following vowel or ER and preceding a pause -> DX")
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
func (s *Sam) AdjustLengths() {
	// LENGTHEN VOWELS PRECEDING PUNCTUATION
	//
	// Search for punctuation. If found, back up to the first vowel, then
	// process all phonemes between there and up to (but not including) the punctuation.
	// If any phoneme is found that is a either a fricative or voiced, the duration is
	// increased by (length * 1.5) + 1

	// loop index
	var X byte = 0
	var index byte
	index = s.phonemeindex[X]
	for s.phonemeindex[X] != render.PhonemeEnd {
		index = s.phonemeindex[X]
		// not punctuation?
		if (flags[index] & FLAG_PUNCT) == 0 {
			X++
			continue
		}

		var loopIndex byte = X

		X--
		for X != 0 && !(flags[s.phonemeindex[X]]&FLAG_VOWEL != 0) {
			X--
		} // back up while not a vowel
		if X == 0 {
			break
		}

		for ok := true; ok; ok = (X != loopIndex) {
			// test for vowel
			index = s.phonemeindex[X]

			// test for fricative/unvoiced or not voiced
			if !(flags[index]&FLAG_FRICATIVE != 0) || (flags[index]&FLAG_VOICED != 0) { //nochmal �berpr�fen
				var A = s.phonemeLength[X]
				// change phoneme length to (length * 1.5) + 1
				s.drule_pre("Lengthen <FRICATIVE> or <VOICED> between <VOWEL> and <PUNCTUATION> by 1.5", X)
				s.phonemeLength[X] = (A >> 1) + A + 1
				s.drule_post(X)
			}
			X++
		}
		X++
	} // while

	// Similar to the above routine, but shorten vowels under some circumstances

	// Loop throught all phonemes
	var loopIndex byte = 0
	index = s.phonemeindex[loopIndex]
	for s.phonemeindex[loopIndex] != render.PhonemeEnd {
		index = s.phonemeindex[loopIndex]
		X := loopIndex

		if flags[index]&FLAG_VOWEL != 0 {
			index = s.phonemeindex[loopIndex+1]
			if index == render.PhonemeEnd {
				break
			}
			if !(flags[index]&FLAG_CONSONANT != 0) {
				if (index == 18) || (index == 19) { // 'RX', 'LX'
					index = s.phonemeindex[loopIndex+2]
					if index != render.PhonemeEnd && flags[index]&FLAG_CONSONANT != 0 {
						s.drule_pre("<VOWEL> <RX | LX> <CONSONANT> - decrease length of vowel by 1\n", loopIndex)
						s.phonemeLength[loopIndex]--
						s.drule_post(loopIndex)
					}
				}
			} else { // Got here if not <VOWEL>
				var flag = flags[index] // 65 if end marker
				if index == render.PhonemeEnd {
					flag = 65
				}

				if !(flag&FLAG_VOICED != 0) { // Unvoiced
					// *, .*, ?*, ,*, -*, DX, S*, SH, F*, TH, /H, /X, CH, P*, T*, K*, KX
					if flag&FLAG_PLOSIVE != 0 { // unvoiced plosive
						// RULE: <VOWEL> <UNVOICED PLOSIVE>
						// <VOWEL> <P*, T*, K*, KX>
						s.drule_pre("<VOWEL> <UNVOICED PLOSIVE> - decrease vowel by 1/8th", loopIndex)
						s.phonemeLength[loopIndex] -= (s.phonemeLength[loopIndex] >> 3)
						s.drule_post(loopIndex)
					}
				} else {
					s.drule_pre("<VOWEL> <VOICED CONSONANT> - increase vowel by 1/2 + 1\n", X-1)
					// decrease length
					var A = s.phonemeLength[loopIndex]
					s.phonemeLength[loopIndex] = (A >> 2) + A + 1 // 5/4*A + 1
					s.drule_post(loopIndex)
				}
			}
		} else if (flags[index] & FLAG_NASAL) != 0 { // nasal?
			// RULE: <NASAL> <STOP CONSONANT>
			//       Set punctuation length to 6
			//       Set stop consonant length to 5
			X++
			index = s.phonemeindex[X]
			if index != render.PhonemeEnd && (flags[index]&FLAG_STOPCONS != 0) {
				s.drule("<NASAL> <STOP CONSONANT> - set nasal = 5, consonant = 6")
				s.phonemeLength[X] = 6   // set stop consonant length to 6
				s.phonemeLength[X-1] = 5 // set nasal length to 5
			}
		} else if flags[index]&FLAG_STOPCONS != 0 { // (voiced) stop consonant?
			// RULE: <VOICED STOP CONSONANT> {optional silence} <STOP CONSONANT>
			//       Shorten both to (length/2 + 1)

			// move past silence
			X++
			index = s.phonemeindex[X]
			for s.phonemeindex[X] == 0 {
				X++
				index = s.phonemeindex[X]
			}

			if index != render.PhonemeEnd && (flags[index]&FLAG_STOPCONS != 0) {
				// FIXME, this looks wrong?
				// RULE: <UNVOICED STOP CONSONANT> {optional silence} <STOP CONSONANT>
				s.drule("<UNVOICED STOP CONSONANT> {optional silence} <STOP CONSONANT> - shorten both to 1/2 + 1")
				s.phonemeLength[X] = (s.phonemeLength[X] >> 1) + 1
				s.phonemeLength[loopIndex] = (s.phonemeLength[loopIndex] >> 1) + 1
				X = loopIndex
			}
		} else if flags[index]&FLAG_LIQUIC != 0 { // liquic consonant?
			// RULE: <VOICED NON-VOWEL> <DIPTHONG>
			//       Decrease <DIPTHONG> by 2
			index = s.phonemeindex[X-1] // prior phoneme;

			// FIXME: The global.Debug code here breaks the rule.
			// prior phoneme a stop consonant>
			if (flags[index] & FLAG_STOPCONS) != 0 {
				s.drule_pre("<LIQUID CONSONANT> <DIPTHONG> - decrease by 2", X)
				s.phonemeLength[X] -= 2 // 20ms
				s.drule_post(X)
			}
		}

		loopIndex++
	}
}
