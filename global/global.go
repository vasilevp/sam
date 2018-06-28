package global

import "fmt"

const (
	Debug              = true
	PHONEME_PERIOD     = 1
	PHONEME_QUESTION   = 2
	RISING_INFLECTION  = 1
	FALLING_INFLECTION = 255
)

var SignInputTable1 []byte
var SignInputTable2 []byte

var Bufferpos int
var Buffer []byte
var PhonemeIndexOutput []byte
var PhonemeLengthOutput []byte
var StressOutput []byte

const (
	PR    = 23
	PD    = 57
	PT    = 69
	BREAK = 254
	END   = 255
)

var Input [256]byte //tab39445
//standard sam sound
var Speed byte = 72
var Pitch byte = 64
var Mouth byte = 128
var Throat byte = 128
var Singmode = false

func PrintPhonemes(phonemeindex [256]byte, phonemeLength [256]byte, stress [256]byte) {
	i := 0
	fmt.Printf("===========================================\n")

	fmt.Printf("Internal Phoneme presentation:\n\n")
	fmt.Printf(" idx    phoneme  length  stress\n")
	fmt.Printf("------------------------------\n")

	for (phonemeindex[i] != 255) && (i < 255) {
		if phonemeindex[i] < 81 {
			fmt.Printf(" %3i      %c%c      %3i       %i\n",
				phonemeindex[i],
				SignInputTable1[phonemeindex[i]],
				SignInputTable2[phonemeindex[i]],
				phonemeLength[i],
				stress[i],
			)
		} else {
			fmt.Printf(" %3i      ??      %3i       %i\n", phonemeindex[i], phonemeLength[i], stress[i])
		}
		i++
	}
	fmt.Printf("===========================================\n")
	fmt.Printf("\n")
}

func PrintOutput(flag, f1, f2, f3, a1, a2, a3, p []byte) {
	fmt.Printf("===========================================\n")
	fmt.Printf("Final data for speech output:\n\n")
	i := 0
	fmt.Printf(" flags ampl1 freq1 ampl2 freq2 ampl3 freq3 pitch\n")
	fmt.Printf("------------------------------------------------\n")
	for i < 255 {
		fmt.Printf("%5i %5i %5i %5i %5i %5i %5i %5i\n", flag[i], a1[i], f1[i], a2[i], f2[i], a3[i], f3[i], p[i])
		i++
	}
	fmt.Printf("===========================================\n")

}

func PrintRule(offset int) {
	i := 1
	var A byte = 0
	fmt.Printf("Applying rule: ")
	for (A & 128) == 0 {
		A = GetRuleByte(offset, byte(i))
		if (A & 127) == '=' {
			fmt.Printf(" -> ")
		} else {
			fmt.Printf("%c", A&127)
		}
		i++
	}
	fmt.Printf("\n")
}

func GetRuleByte(mem62 int, Y byte) byte {
	address := mem62
	if mem62 >= 37541 {
		address -= 37541
		return Rules2[address+int(Y)]
	}
	address -= 32000
	return Rules[address+int(Y)]
}
