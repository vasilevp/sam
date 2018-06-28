package render

import (
	"github.com/exploser/sam/config"
)

func (r *Render) CombineGlottalAndFormants(phase1, phase2, phase3, Y byte) {
	var tmp uint

	tmp = uint(multtable[sinus[phase1]|r.amplitude1[Y]])
	tmp += uint(multtable[sinus[phase2]|r.amplitude2[Y]])
	// tmp  += tmp > 255 ? 1 : 0; // if addition above overflows, we for some reason add one;
	if tmp > 255 {
		tmp++
	}
	tmp += uint(multtable[rectangle[phase3]|r.amplitude3[Y]])
	tmp += 136
	tmp >>= 4 // Scale down to 0..15 range of C64 audio.

	r.Output(0, byte(tmp&0xf))
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
func (r *Render) ProcessFrames(count byte, cfg *config.Config) {

	var speedcounter byte = 72
	var phase1 byte
	var phase2 byte
	var phase3 byte
	var mem66 byte

	var Y byte

	var glottal_pulse = r.pitches[0]
	var mem38 = glottal_pulse - (glottal_pulse >> 2) // mem44 * 0.75

	for count != 0 {
		var flags = r.sampledConsonantFlag[Y]

		// unvoiced sampled phoneme?
		if flags&248 != 0 {
			r.RenderSample(&mem66, flags, Y)
			// skip ahead two in the phoneme buffer
			Y += 2
			count -= 2
			speedcounter = cfg.Speed
		} else {
			r.CombineGlottalAndFormants(phase1, phase2, phase3, Y)

			speedcounter--
			if speedcounter == 0 {
				Y++ //go to next amplitude
				// decrement the frame count
				count--
				if count == 0 {
					return
				}
				speedcounter = cfg.Speed
			}

			glottal_pulse--

			if glottal_pulse != 0 {
				// not finished with a glottal pulse

				mem38--
				// within the first 75% of the glottal pulse?
				// is the count non-zero and the sampled flag is zero?
				if (mem38 != 0) || (flags == 0) {
					// reset the phase of the formants to match the pulse
					phase1 += r.frequency1[Y]
					phase2 += r.frequency2[Y]
					phase3 += r.frequency3[Y]
					continue
				}

				// voiced sampled phonemes interleave the sample with the
				// glottal pulse. The sample flag is non-zero, so render
				// the sample for the phoneme.
				r.RenderSample(&mem66, flags, Y)
			}
		}

		glottal_pulse = r.pitches[Y]
		mem38 = glottal_pulse - (glottal_pulse >> 2) // mem44 * 0.75

		// reset the formant wave generators to keep them in
		// sync with the glottal pulse
		phase1 = 0
		phase2 = 0
		phase3 = 0
	}
}
