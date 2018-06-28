package render

import (
	"test/sam/global"
)

func CombineGlottalAndFormants(phase1, phase2, phase3, Y byte) {
	var tmp uint

	tmp = uint(multtable[sinus[phase1]|amplitude1[Y]])
	tmp += uint(multtable[sinus[phase2]|amplitude2[Y]])
	// tmp  += tmp > 255 ? 1 : 0; // if addition above overflows, we for some reason add one;
	if tmp > 255 {
		tmp++
	}
	tmp += uint(multtable[rectangle[phase3]|amplitude3[Y]])
	tmp += 136
	tmp >>= 4 // Scale down to 0..15 range of C64 audio.

	Output(0, byte(tmp&0xf))
}

// PROCESS THE FRAMES
//
// In traditional vocal synthesis, the glottal pulse drives filters, which
// are attenuated to the frequencies of the formants.
//
// SAM generates these formants directly with sin and rectangular waves.
// To simulate them being driven by the glottal pulse, the waveforms are
// reset at the beginning of each glottal pulse.
//
func ProcessFrames(mem48 byte) {

	var speedcounter byte = 72
	var phase1 byte = 0
	var phase2 byte = 0
	var phase3 byte = 0
	var mem66 byte

	var Y byte = 0

	var glottal_pulse = pitches[0]
	var mem38 = glottal_pulse - (glottal_pulse >> 2) // mem44 * 0.75

	for mem48 != 0 {
		var flags = sampledConsonantFlag[Y]

		// unvoiced sampled phoneme?
		if flags&248 != 0 {
			RenderSample(&mem66, flags, Y)
			// skip ahead two in the phoneme buffer
			Y += 2
			mem48 -= 2
			speedcounter = global.Speed
		} else {
			CombineGlottalAndFormants(phase1, phase2, phase3, Y)

			speedcounter--
			if speedcounter == 0 {
				Y++ //go to next amplitude
				// decrement the frame count
				mem48--
				if mem48 == 0 {
					return
				}
				speedcounter = global.Speed
			}

			glottal_pulse--

			if glottal_pulse != 0 {
				// not finished with a glottal pulse

				mem38--
				// within the first 75% of the glottal pulse?
				// is the count non-zero and the sampled flag is zero?
				if (mem38 != 0) || (flags == 0) {
					// reset the phase of the formants to match the pulse
					phase1 += frequency1[Y]
					phase2 += frequency2[Y]
					phase3 += frequency3[Y]
					continue
				}

				// voiced sampled phonemes interleave the sample with the
				// glottal pulse. The sample flag is non-zero, so render
				// the sample for the phoneme.
				RenderSample(&mem66, flags, Y)
			}
		}

		glottal_pulse = pitches[Y]
		mem38 = glottal_pulse - (glottal_pulse >> 2) // mem44 * 0.75

		// reset the formant wave generators to keep them in
		// sync with the glottal pulse
		phase1 = 0
		phase2 = 0
		phase3 = 0
	}
}
