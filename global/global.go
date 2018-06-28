package global

import (
	"fmt"

	"github.com/exploser/sam/global"
)

const (
	PHONEME_PERIOD     = 1
	PHONEME_QUESTION   = 2
	RISING_INFLECTION  = 1
	FALLING_INFLECTION = 255
)

var (
	StressInputTable = []byte{
		'*', '1', '2', '3', '4', '5', '6', '7', '8',
	}

	//tab40682
	SignInputTable1 = []byte{
		' ', '.', '?', ',', '-', 'I', 'I', 'E',
		'A', 'A', 'A', 'A', 'U', 'A', 'I', 'E',
		'U', 'O', 'R', 'L', 'W', 'Y', 'W', 'R',
		'L', 'W', 'Y', 'M', 'N', 'N', 'D', 'Q',
		'S', 'S', 'F', 'T', '/', '/', 'Z', 'Z',
		'V', 'D', 'C', '*', 'J', '*', '*', '*',
		'E', 'A', 'O', 'A', 'O', 'U', 'B', '*',
		'*', 'D', '*', '*', 'G', '*', '*', 'G',
		'*', '*', 'P', '*', '*', 'T', '*', '*',
		'K', '*', '*', 'K', '*', '*', 'U', 'U',
		'U',
	}

	//tab40763
	SignInputTable2 = []byte{
		'*', '*', '*', '*', '*', 'Y', 'H', 'H',
		'E', 'A', 'H', 'O', 'H', 'X', 'X', 'R',
		'X', 'H', 'X', 'X', 'X', 'X', 'H', '*',
		'*', '*', '*', '*', '*', 'X', 'X', '*',
		'*', 'H', '*', 'H', 'H', 'X', '*', 'H',
		'*', 'H', 'H', '*', '*', '*', '*', '*',
		'Y', 'Y', 'Y', 'W', 'W', 'W', '*', '*',
		'*', '*', '*', '*', '*', '*', '*', 'X',
		'*', '*', '*', '*', '*', '*', '*', '*',
		'*', '*', '*', 'X', '*', '*', 'L', 'M',
		'N',
	}
)

var PhonemeIndexOutput [60]byte //tab47296
var PhonemeLengthOutput [60]byte
var StressOutput [60]byte

const (
	PR    = 23
	PD    = 57
	PT    = 69
	BREAK = 254
	END   = 255
)

func PrintPhonemes(phonemeindex []byte, phonemeLength []byte, stress []byte) {
	i := 0
	fmt.Printf("===========================================\n")

	fmt.Printf("Internal Phoneme presentation:\n\n")
	fmt.Printf(" idx    phoneme  length  stress\n")
	fmt.Printf("------------------------------\n")

	for (phonemeindex[i] != global.END) && (i < 255) {
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

func PrintOutput(flag, f1, f2, f3, a1, a2, a3, p []byte) {
	fmt.Printf("===========================================\n")
	fmt.Printf("Final data for speech output:\n\n")
	i := 0
	fmt.Printf(" flags ampl1 freq1 ampl2 freq2 ampl3 freq3 pitch\n")
	fmt.Printf("------------------------------------------------\n")
	for i < 255 {
		fmt.Printf("%5d %5d %5d %5d %5d %5d %5d %5d\n", flag[i], a1[i], f1[i], a2[i], f2[i], a3[i], f3[i], p[i])
		i++
	}
	fmt.Printf("===========================================\n")

}

func PrintRule(offset int) {
	i := 1
	var A byte
	fmt.Printf("Applying rule: ")
	for (A & 128) == 0 {
		A = GetRuleByte(uint16(offset), byte(i))
		if (A & 127) == '=' {
			fmt.Printf(" -> ")
		} else {
			fmt.Printf("%c", A&127)
		}
		i++
	}
	fmt.Printf("\n")
}

func GetRuleByte(mem62 uint16, Y byte) byte {
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
