package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"test/sam/global"
	"test/sam/reciter"
	"test/sam/sammain"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	wavplayer "github.com/faiface/beep/wav"
	wav "github.com/youpy/go-wav"
)

func PrintUsage() {
	fmt.Printf("usage: sam [options] Word1 Word2 ....\n")
	fmt.Printf("options\n")
	fmt.Printf("	-phonetic 		enters phonetic mode. (see below)\n")
	fmt.Printf("	-pitch number		set pitch value (default=64)\n")
	fmt.Printf("	-speed number		set speed value (default=72)\n")
	fmt.Printf("	-throat number		set throat value (default=128)\n")
	fmt.Printf("	-mouth number		set mouth value (default=128)\n")
	fmt.Printf("	-wav filename		output to wav instead of libsdl\n")
	fmt.Printf("	-sing			special treatment of pitch\n")
	fmt.Printf("	-debug			print additional debug messages\n")
	fmt.Printf("\n")

	fmt.Printf("     VOWELS                            VOICED CONSONANTS	\n")
	fmt.Printf("IY           f(ee)t                    R        red		\n")
	fmt.Printf("IH           p(i)n                     L        allow		\n")
	fmt.Printf("EH           beg                       W        away		\n")
	fmt.Printf("AE           Sam                       W        whale		\n")
	fmt.Printf("AA           pot                       Y        you		\n")
	fmt.Printf("AH           b(u)dget                  M        Sam		\n")
	fmt.Printf("AO           t(al)k                    N        man		\n")
	fmt.Printf("OH           cone                      NX       so(ng)		\n")
	fmt.Printf("UH           book                      B        bad		\n")
	fmt.Printf("UX           l(oo)t                    D        dog		\n")
	fmt.Printf("ER           bird                      G        again		\n")
	fmt.Printf("AX           gall(o)n                  J        judge		\n")
	fmt.Printf("IX           dig(i)t                   Z        zoo		\n")
	fmt.Printf("				       ZH       plea(s)ure	\n")
	fmt.Printf("   DIPHTHONGS                          V        seven		\n")
	fmt.Printf("EY           m(a)de                    DH       (th)en		\n")
	fmt.Printf("AY           h(igh)						\n")
	fmt.Printf("OY           boy						\n")
	fmt.Printf("AW           h(ow)                     UNVOICED CONSONANTS	\n")
	fmt.Printf("OW           slow                      S         Sam		\n")
	fmt.Printf("UW           crew                      Sh        fish		\n")
	fmt.Printf("                                       F         fish		\n")
	fmt.Printf("                                       TH        thin		\n")
	fmt.Printf(" SPECIAL PHONEMES                      P         poke		\n")
	fmt.Printf("UL           sett(le) (=AXL)           T         talk		\n")
	fmt.Printf("UM           astron(omy) (=AXM)        K         cake		\n")
	fmt.Printf("UN           functi(on) (=AXN)         CH        speech		\n")
	fmt.Printf("Q            kitt-en (glottal stop)    /H        a(h)ead	\n")
}

func main() {
	var i int
	var phonetic = false

	var input string

	if len(os.Args) <= 1 {
		PrintUsage()
		os.Exit(1)
	}

	wavfilename := ""

	i = 1
	for i < len(os.Args) {
		if os.Args[i][0] != '-' {
			input += os.Args[i] + " "
		} else {
			switch os.Args[i][1:] {
			case "wav":
				wavfilename = os.Args[i+1]
				i++
			case "sing":
				sammain.EnableSingmode()
			case "phonetic":
				phonetic = true
			case "debug":
				global.Debug = true
			case "pitch":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				sammain.SetPitch(byte(val))
				i++
			case "speed":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				sammain.SetSpeed(byte(val))
				i++
			case "mouth":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				sammain.SetMouth(byte(val))
				i++
			case "throat":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				sammain.SetThroat(byte(val))
				i++
			default:
				PrintUsage()
				os.Exit(1)
			}
		}

		i++
	} //while

	input = strings.ToUpper(input)

	var data [256]byte
	for i = 0; i < len(input); i++ {
		data[i] = byte(input[i])
	}

	if global.Debug {
		if phonetic {
			fmt.Printf("phonetic input: %s\n", string(data[:]))
		} else {
			fmt.Printf("text input: %s\n", string(data[:]))
		}
	}

	if !phonetic {
		data[i] = '['

		if reciter.TextToPhonemes(data[:]) == 0 {
			os.Exit(1)
		}
		if global.Debug {
			fmt.Printf("phonetic input: %s\n", data)
		}
	} else {
		input += "\x9b"
	}

	sammain.SetInput(data)
	if !sammain.SAMMain() {
		// PrintUsage()
		os.Exit(2)
	}

	if wavfilename != "" {
		file, err := os.Create(wavfilename)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(3)
		}
		_, err = wav.NewWriter(file, uint32(sammain.GetBufferLength()+22050), 1, 22050, 8).Write(sammain.GetBuffer())
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(4)
		}
	} else {
		buf := &bytes.Buffer{}
		_, err := wav.NewWriter(buf, uint32(sammain.GetBufferLength()+22050), 1, 22050, 8).Write(sammain.GetBuffer())
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(5)
		}
		reader := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
		s, format, err := wavplayer.Decode(reader)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(6)
		}

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		done := make(chan int)
		speaker.Play(beep.Seq(s, beep.Callback(func() {
			done <- 0
		})))
		<-done
	}
}
