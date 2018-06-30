package render

import (
	"fmt"

	"github.com/exploser/sam/config"
)

type Render struct {
	Bufferpos int
	Buffer    []byte

	pitches [256]byte // tab43008

	frequency1 [256]byte
	frequency2 [256]byte
	frequency3 [256]byte

	amplitude1           [256]byte
	amplitude2           [256]byte
	amplitude3           [256]byte
	sampledConsonantFlag [256]byte // tab44800
	oldtimetableindex    int

	PhonemeIndexOutput  [256]byte //tab47296
	PhonemeLengthOutput [256]byte
	StressOutput        [256]byte
}

func (r *Render) GetBuffer() []byte {
	return r.Buffer
}
func (r *Render) GetBufferLength() int {
	return r.Bufferpos/50 + 5
}

//extern byte A, X, Y;
//extern byte mem44;

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

func (r *Render) Output8Bit(index int, A byte) {
	var k int
	r.Bufferpos += timetable[r.oldtimetableindex][index]
	r.oldtimetableindex = index
	// write a little bit in advance
	for k = 0; k < 5; k++ {
		i := r.Bufferpos/50 + k
		r.Buffer[i] = A
	}
}

func (r *Render) Output8BitArray(index int, A [5]byte) {
	var k int
	r.Bufferpos += timetable[r.oldtimetableindex][index]
	r.oldtimetableindex = index

	if len(A) != 5 {
		panic("A should contain exactly 5 elements")
	}

	// write a little bit in advance
	for k = 0; k < 5; k++ {
		r.Buffer[r.Bufferpos/50+k] = A[k]
	}
}

func (r *Render) RenderVoicedSample(hi uint, off, phase1 byte) byte {
	for ok := true; ok; ok = (phase1 != 0) {
		sample := sampleTable[hi+uint(off)]
		var bit byte = 8
		for ok := true; ok; ok = (bit != 0) {
			if (sample & 128) != 0 {
				r.Output8Bit(3, 0x70)
			} else {
				r.Output8Bit(4, 0x90)
			}
			sample <<= 1
			bit--
		}
		off++
		phase1++
	}
	return off
}

func (r *Render) RenderUnvoicedSample(hi uint, off, X byte) {
	for ok := true; ok; ok = (off != 0) {
		var bit byte = 8
		sample := sampleTable[hi+uint(off)]
		for ok := true; ok; ok = (bit != 0) {
			if (sample & 128) != 0 {
				r.Output8Bit(2, 5*16)
			} else {
				r.Output8Bit(1, X*16) // TODO: check for 0xf
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

func (r *Render) RenderSample(mem66 *byte, consonantFlag, mem49 byte) {
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
		pitch = r.pitches[mem49] >> 4
		*mem66 = r.RenderVoicedSample(hi, *mem66, pitch^255)
		return
	}
	r.RenderUnvoicedSample(hi, pitch^255, tab48426[hibyte])
}

// CREATE FRAMES
//
// The length parameter in the list corresponds to the number of frames
// to expand the phoneme to. Each frame represents 10 milliseconds of time.
// So a phoneme with a length of 7 = 7 frames = 70 milliseconds duration.
//
// The parameters are copied from the phoneme to the frame verbatim.
//
func (r *Render) CreateFrames(cfg *config.Config) {
	var phase1, X byte

	i := 0
	for i < 256 {
		// get the phoneme at the index
		phoneme := r.PhonemeIndexOutput[i]

		// if terminal phoneme, exit the loop
		if phoneme == PhonemeEnd {
			break
		}

		if phoneme == PhonemePeriod {
			r.AddInflection(inflectionRising, phase1, X)
		} else if phoneme == PhonemeQuestion {
			r.AddInflection(inflectionFalling, phase1, X)
		}

		// get the stress amount (more stress = higher pitch)
		phase1 = tab47492[r.StressOutput[i]+1]

		// get number of frames to write
		phase2 := r.PhonemeLengthOutput[i]

		// copy from the source to the frames list
		for ok := true; ok; ok = (phase2 != 0) {
			r.frequency1[X] = freq1data[phoneme]                       // F1 frequency
			r.frequency2[X] = freq2data[phoneme]                       // F2 frequency
			r.frequency3[X] = freq3data[phoneme]                       // F3 frequency
			r.amplitude1[X] = ampl1data[phoneme]                       // F1 amplitude
			r.amplitude2[X] = ampl2data[phoneme]                       // F2 amplitude
			r.amplitude3[X] = ampl3data[phoneme]                       // F3 amplitude
			r.sampledConsonantFlag[X] = sampledConsonantFlags[phoneme] // phoneme data for sampled consonants
			r.pitches[X] = cfg.Pitch + phase1                          // pitch
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
func (r *Render) RescaleAmplitude() {
	for i := 255; i >= 0; i-- {
		r.amplitude1[i] = amplitudeRescale[r.amplitude1[i]]
		r.amplitude2[i] = amplitudeRescale[r.amplitude2[i]]
		r.amplitude3[i] = amplitudeRescale[r.amplitude3[i]]
	}
}

// ASSIGN PITCH CONTOUR
//
// This subtracts the F1 frequency from the pitch to create a
// pitch contour. Without this, the output would be at a single
// pitch level (monotone).

func (r *Render) AssignPitchContour() {
	for i := 0; i < 256; i++ {
		// subtract half the frequency of the formant 1.
		// this adds variety to the voice
		r.pitches[i] -= (r.frequency1[i] >> 1)
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
func (r *Render) Render(cfg *config.Config) {
	if r.PhonemeIndexOutput[0] == PhonemeEnd {
		return
	} //exit if no data

	r.CreateFrames(cfg)
	t := r.CreateTransitions()

	if !cfg.Sing {
		r.AssignPitchContour()
	}
	r.RescaleAmplitude()

	if cfg.Debug {
		PrintOutput(r.sampledConsonantFlag[:], r.frequency1[:], r.frequency2[:], r.frequency3[:], r.amplitude1[:], r.amplitude2[:], r.amplitude3[:], r.pitches[:])
	}

	r.ProcessFrames(t, cfg)
}

// Create a rising or falling inflection 30 frames prior to
// index X. A rising inflection is used for questions, and
// a falling inflection is used for statements.

func (r *Render) AddInflection(inflection, phase1, pos byte) {
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
	for r.pitches[pos] == 127 {
		pos++
	}

	// pos-1 can become 255 and cause weird shit to happen
	if pos == 0 {
		A = r.pitches[pos]
	} else {
		A = r.pitches[pos-1]
	}

	for pos != end {
		// add the inflection direction
		A += inflection

		// set the inflection
		r.pitches[pos] = A

		pos++
		for (pos != end) && r.pitches[pos] == 255 {
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
	var newFrequency byte

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
