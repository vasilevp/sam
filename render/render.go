package render

import (
	"github.com/exploser/sam/global"
)

var wait1 byte = 7
var wait2 byte = 6

//extern byte A, X, Y;
//extern byte mem44;

var pitches [256]byte // tab43008

var frequency1 [256]byte
var frequency2 [256]byte
var frequency3 [256]byte

var amplitude1 [256]byte
var amplitude2 [256]byte
var amplitude3 [256]byte

var sampledConsonantFlag [256]byte // tab44800

//return = hibyte(mem39212*mem39213) <<  1
func trans(a, b byte) byte {
	return byte(((int(a) * int(b)) >> 8) << 1)
}

//timetable for more accurate c64 simulation
var timetable = [5][5]int{
	{162, 167, 167, 127, 128},
	{226, 60, 60, 0, 0},
	{225, 60, 59, 0, 0},
	{200, 0, 0, 54, 55},
	{199, 0, 0, 54, 54},
}

var oldtimetableindex int = 0

func Output(index int, A byte) {
	var k int
	global.Bufferpos += timetable[oldtimetableindex][index]
	oldtimetableindex = index
	// write a little bit in advance
	for k = 0; k < 5; k++ {
		global.Buffer[global.Bufferpos/50+k] = (A & 15) * 16
	}
}

func RenderVoicedSample(hi uint, off, phase1 byte) byte {
	for ok := true; ok; ok = (phase1 != 0) {
		sample := sampleTable[hi+uint(off)]
		var bit byte = 8
		for ok := true; ok; ok = (bit != 0) {
			if (sample & 128) != 0 {
				Output(3, 26)
			} else {
				Output(4, 6)
			}
			sample <<= 1
			bit--
		}
		off++
		phase1++
	}
	return off
}

func RenderUnvoicedSample(hi uint, off, mem53 byte) {
	for ok := true; ok; ok = (off != 0) {
		var bit byte = 8
		sample := sampleTable[hi+uint(off)]
		for ok := true; ok; ok = (bit != 0) {
			if (sample & 128) != 0 {
				Output(2, 5)
			} else {
				Output(1, mem53)
			}
			sample <<= 1
			bit--
		}
		off++
	}
}

// -------------------------------------------------------------------------
//Code48227
// Render a sampled sound from the sampleTable.
//
//   Phoneme   Sample Start   Sample End
//   32: S*    15             255
//   33: SH    257            511
//   34: F*    559            767
//   35: TH    583            767
//   36: /H    903            1023
//   37: /X    1135           1279
//   38: Z*    84             119
//   39: ZH    340            375
//   40: V*    596            639
//   41: DH    596            631
//
//   42: CH
//   43: **    399            511
//
//   44: J*
//   45: **    257            276
//   46: **
//
//   66: P*
//   67: **    743            767
//   68: **
//
//   69: T*
//   70: **    231            255
//   71: **
//
// The SampledPhonemesTable[] holds flags indicating if a phoneme is
// voiced or not. If the upper 5 bits are zero, the sample is voiced.
//
// Samples in the sampleTable are compressed, with bits being converted to
// bytes from high bit to low, as follows:
//
//   unvoiced 0 bit   -> X
//   unvoiced 1 bit   -> 5
//
//   voiced 0 bit     -> 6
//   voiced 1 bit     -> 24
//
// Where X is a value from the table:
//
//   { 0x18, 0x1A, 0x17, 0x17, 0x17 };
//
// The index into this table is determined by masking off the lower
// 3 bits from the SampledPhonemesTable:
//
//        index = (SampledPhonemesTable[i] & 7) - 1;
//
// For voices samples, samples are interleaved between voiced output.

func RenderSample(mem66 *byte, consonantFlag, mem49 byte) {
	// mem49 == current phoneme's index

	// mask low three bits and subtract 1 get value to
	// convert 0 bits on unvoiced samples.
	hibyte := (consonantFlag & 7) - 1

	// determine which offset to use from table { 0x18, 0x1A, 0x17, 0x17, 0x17 }
	// T, S, Z                0          0x18
	// CH, J, SH, ZH          1          0x1A
	// P, F*, V, TH, DH       2          0x17
	// /H                     3          0x17
	// /X                     4          0x17

	hi := uint(hibyte) * 256
	// voiced sample?
	pitch := consonantFlag & 248
	if pitch == 0 {
		// voiced phoneme: Z*, ZH, V*, DH
		pitch = pitches[mem49] >> 4
		*mem66 = RenderVoicedSample(hi, *mem66, pitch^255)
		return
	}
	RenderUnvoicedSample(hi, pitch^255, tab48426[hibyte])
}

// CREATE FRAMES
//
// The length parameter in the list corresponds to the number of frames
// to expand the phoneme to. Each frame represents 10 milliseconds of time.
// So a phoneme with a length of 7 = 7 frames = 70 milliseconds duration.
//
// The parameters are copied from the phoneme to the frame verbatim.
//
func CreateFrames() {
	var phase1, X byte

	i := 0
	for i < 256 {
		// get the phoneme at the index
		phoneme := global.PhonemeIndexOutput[i]

		// if terminal phoneme, exit the loop
		if phoneme == 255 {
			break
		}

		if phoneme == global.PHONEME_PERIOD {
			AddInflection(global.RISING_INFLECTION, phase1, X)
		} else if phoneme == global.PHONEME_QUESTION {
			AddInflection(global.FALLING_INFLECTION, phase1, X)
		}

		// get the stress amount (more stress = higher pitch)
		phase1 = tab47492[global.StressOutput[i]+1]

		// get number of frames to write
		phase2 := global.PhonemeLengthOutput[i]

		// copy from the source to the frames list
		for ok := true; ok; ok = (phase2 != 0) {
			frequency1[X] = freq1data[phoneme]                       // F1 frequency
			frequency2[X] = freq2data[phoneme]                       // F2 frequency
			frequency3[X] = freq3data[phoneme]                       // F3 frequency
			amplitude1[X] = ampl1data[phoneme]                       // F1 amplitude
			amplitude2[X] = ampl2data[phoneme]                       // F2 amplitude
			amplitude3[X] = ampl3data[phoneme]                       // F3 amplitude
			sampledConsonantFlag[X] = sampledConsonantFlags[phoneme] // phoneme data for sampled consonants
			pitches[X] = global.Pitch + phase1                       // pitch
			X++
			phase2--
		}

		i++
	}
}

// RESCALE AMPLITUDE
//
// Rescale volume from a linear scale to decibels.
//
func RescaleAmplitude() {
	for i := 255; i >= 0; i-- {
		amplitude1[i] = amplitudeRescale[amplitude1[i]]
		amplitude2[i] = amplitudeRescale[amplitude2[i]]
		amplitude3[i] = amplitudeRescale[amplitude3[i]]
	}
}

// ASSIGN PITCH CONTOUR
//
// This subtracts the F1 frequency from the pitch to create a
// pitch contour. Without this, the output would be at a single
// pitch level (monotone).

func AssignPitchContour() {
	for i := 0; i < 256; i++ {
		// subtract half the frequency of the formant 1.
		// this adds variety to the voice
		pitches[i] -= (frequency1[i] >> 1)
	}
}

// RENDER THE PHONEMES IN THE LIST
//
// The phoneme list is converted into sound through the steps:
//
// 1. Copy each phoneme <length> number of times into the frames list,
//    where each frame represents 10 milliseconds of sound.
//
// 2. Determine the transitions lengths between phonemes, and linearly
//    interpolate the values across the frames.
//
// 3. Offset the pitches by the fundamental frequency.
//
// 4. Render the each frame.
func Render() {
	if global.PhonemeIndexOutput[0] == 255 {
		return
	} //exit if no data

	CreateFrames()
	t := CreateTransitions()

	if !global.Singmode {
		AssignPitchContour()
	}
	RescaleAmplitude()

	if global.Debug {
		global.PrintOutput(sampledConsonantFlag[:], frequency1[:], frequency2[:], frequency3[:], amplitude1[:], amplitude2[:], amplitude3[:], pitches[:])
	}

	ProcessFrames(t)
}

// Create a rising or falling inflection 30 frames prior to
// index X. A rising inflection is used for questions, and
// a falling inflection is used for statements.

func AddInflection(inflection, phase1, pos byte) {
	var A byte
	// store the location of the punctuation
	end := pos

	if pos < 30 {
		pos = 0
	} else {
		pos -= 30
	}

	// FIXME: Explain this fix better, it's not obvious
	// ML : A =, fixes a problem with invalid pitch with '.'
	for pitches[pos] == 127 {
		pos++
	}
	A = pitches[pos-1]

	for pos != end {
		// add the inflection direction
		A += inflection

		// set the inflection
		pitches[pos] = A

		pos++
		for (pos != end) && pitches[pos] == 255 {
			pos++
		}
	}
}

/*
   SAM's voice can be altered by changing the frequencies of the
   mouth formant (F1) and the throat formant (F2). Only the voiced
   phonemes (5-29 and 48-53) are altered.
*/
func SetMouthThroat(mouth, throat byte) {
	var initialFrequency byte
	var newFrequency byte = 0

	// mouth formants (F1) 5..29
	mouthFormants5_29 := [30]byte{
		0, 0, 0, 0, 0, 10,
		14, 19, 24, 27, 23, 21, 16, 20, 14, 18, 14, 18, 18,
		16, 13, 15, 11, 18, 14, 11, 9, 6, 6, 6}

	// throat formants (F2) 5..29
	throatFormants5_29 := [30]byte{
		255, 255,
		255, 255, 255, 84, 73, 67, 63, 40, 44, 31, 37, 45, 73, 49,
		36, 30, 51, 37, 29, 69, 24, 50, 30, 24, 83, 46, 54, 86,
	}

	// there must be no zeros in this 2 tables
	// formant 1 frequencies (mouth) 48..53
	mouthFormants48_53 := [6]byte{19, 27, 21, 27, 18, 13}

	// formant 2 frequencies (throat) 48..53
	throatFormants48_53 := [6]byte{72, 39, 31, 43, 30, 34}

	var pos byte = 5

	// recalculate formant frequencies 5..29 for the mouth (F1) and throat (F2)
	for pos < 30 {
		// recalculate mouth frequency
		initialFrequency = mouthFormants5_29[pos]
		if initialFrequency != 0 {
			newFrequency = trans(mouth, initialFrequency)
		}
		freq1data[pos] = newFrequency

		// recalculate throat frequency
		initialFrequency = throatFormants5_29[pos]
		if initialFrequency != 0 {
			newFrequency = trans(throat, initialFrequency)
		}
		freq2data[pos] = newFrequency
		pos++
	}

	// recalculate formant frequencies 48..53
	pos = 0
	for pos < 6 {
		// recalculate F1 (mouth formant)
		initialFrequency = mouthFormants48_53[pos]
		newFrequency = trans(mouth, initialFrequency)
		freq1data[pos+48] = newFrequency

		// recalculate F2 (throat formant)
		initialFrequency = throatFormants48_53[pos]
		newFrequency = trans(throat, initialFrequency)
		freq2data[pos+48] = newFrequency
		pos++
	}
}
