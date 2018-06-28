package render

import (
	"fmt"
	"test/sam/global"
)

// CREATE TRANSITIONS
//
// Linear transitions are now created to smoothly connect each
// phoeneme. This transition is spread between the ending frames
// of the old phoneme (outBlendLength), and the beginning frames
// of the new phoneme (inBlendLength).
//
// To determine how many frames to use, the two phonemes are
// compared using the blendRank[] table. The phoneme with the
// smaller score is used. In case of a tie, a blend of each is used:
//
//      if blendRank[phoneme1] ==  blendRank[phomneme2]
//          // use lengths from each phoneme
//          outBlendFrames = outBlend[phoneme1]
//          inBlendFrames = outBlend[phoneme2]
//      else if blendRank[phoneme1] < blendRank[phoneme2]
//          // use lengths from first phoneme
//          outBlendFrames = outBlendLength[phoneme1]
//          inBlendFrames = inBlendLength[phoneme1]
//      else
//          // use lengths from the second phoneme
//          // note that in and out are swapped around!
//          outBlendFrames = inBlendLength[phoneme2]
//          inBlendFrames = outBlendLength[phoneme2]
//
//  Blend lengths can't be less than zero.
//
// For most of the parameters, SAM interpolates over the range of the last
// outBlendFrames-1 and the first inBlendFrames.
//
// The exception to this is the Pitch[] parameter, which is interpolates the
// pitch from the center of the current phoneme to the center of the next
// phoneme.

//written by me because of different table positions.
// mem[47] = ...
// 168=pitches
// 169=frequency1
// 170=frequency2
// 171=frequency3
// 172=amplitude1
// 173=amplitude2
// 174=amplitude3
func Read(p, Y byte) byte {
	switch p {
	case 168:
		return pitches[Y]
	case 169:
		return frequency1[Y]
	case 170:
		return frequency2[Y]
	case 171:
		return frequency3[Y]
	case 172:
		return amplitude1[Y]
	case 173:
		return amplitude2[Y]
	case 174:
		return amplitude3[Y]
	}
	fmt.Printf("Error reading to tables")
	return 0
}

func Write(p, Y, value byte) {

	switch p {
	case 168:
		pitches[Y] = value
		return
	case 169:
		frequency1[Y] = value
		return
	case 170:
		frequency2[Y] = value
		return
	case 171:
		frequency3[Y] = value
		return
	case 172:
		amplitude1[Y] = value
		fmt.Println("wrote", value)
		return
	case 173:
		amplitude2[Y] = value
		return
	case 174:
		amplitude3[Y] = value
		return
	}
	fmt.Printf("Error writing to tables\n")
}

func abs(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

// linearly interpolate values
func interpolate(width, table, frame, mem53 byte) {
	sign := (int8(mem53) < 0)
	remainder := byte(int(abs(int8(mem53))) % int(width))
	div := byte(int(int8(mem53)) / int(width))

	var intError byte = 0
	var pos = width
	var val = Read(table, frame) + div

	pos--
	for pos > 0 {
		intError += remainder
		if intError >= width { // accumulated a whole integer error, so adjust output
			intError -= width
			if sign {
				val--
			} else if val != 0 {
				val++
			} // if input is 0, we always leave it alone
		}
		frame++
		Write(table, frame, val) // Write updated value back to next frame.
		val += div
		pos--
	}
}

func interpolate_pitch(width, pos, mem49, phase3 byte) {
	// unlike the other values, the pitches[] interpolates from
	// the middle of the current phoneme to the middle of the
	// next phoneme

	// half the width of the current and next phoneme
	cur_width := global.PhonemeLengthOutput[pos] / 2
	next_width := global.PhonemeLengthOutput[pos+1] / 2
	// sum the values
	width = cur_width + next_width
	pitch := pitches[next_width+mem49] - pitches[mem49-cur_width]
	interpolate(width, 168, phase3, pitch)
}

func CreateTransitions() byte {
	var phase1, phase2, mem49, pos byte
	for {
		phoneme := global.PhonemeIndexOutput[pos]
		next_phoneme := global.PhonemeIndexOutput[pos+1]

		if next_phoneme == 255 {
			break
		} // 255 == end_token

		// get the ranking of each phoneme
		next_rank := blendRank[next_phoneme]
		rank := blendRank[phoneme]

		// compare the rank - lower rank value is stronger
		if rank == next_rank {
			// same rank, so use out blend lengths from each phoneme
			phase1 = outBlendLength[phoneme]
			phase2 = outBlendLength[next_phoneme]
		} else if rank < next_rank {
			// next phoneme is stronger, so us its blend lengths
			phase1 = inBlendLength[next_phoneme]
			phase2 = outBlendLength[next_phoneme]
		} else {
			// current phoneme is stronger, so use its blend lengths
			// note the out/in are swapped
			phase1 = outBlendLength[phoneme]
			phase2 = inBlendLength[phoneme]
		}

		mem49 += global.PhonemeLengthOutput[pos]

		speedcounter := mem49 + phase2
		phase3 := mem49 - phase1
		transition := phase1 + phase2 // total transition?

		if ((transition - 2) & 128) == 0 {

			interpolate_pitch(transition, pos, mem49, phase3)
			var table byte = 169
			for table < 175 {
				// tables:
				// 168  pitches[]
				// 169  frequency1
				// 170  frequency2
				// 171  frequency3
				// 172  amplitude1
				// 173  amplitude2
				// 174  amplitude3

				value := Read(table, speedcounter) - Read(table, phase3)
				interpolate(transition, table, phase3, value)
				table++
			}
		}
		pos++
	}

	// add the length of this phoneme
	return mem49 + global.PhonemeLengthOutput[pos]
}
