package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/exploser/sam/config"
	"github.com/exploser/sam/reciter"
	"github.com/exploser/sam/render"
	"github.com/exploser/sam/sammain"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	wavplayer "github.com/faiface/beep/wav"
	wav "github.com/youpy/go-wav"
)

func main() {
	var args struct {
		config.Config
		Wav          string   `arg:"-w" help:"output to wav instead of libsdl"`
		Input        []string `arg:"positional"`
		Phonetic     bool     `help:"enters phonetic mode (use -P to show phonetic guide)"`
		PhoneticHelp bool     `arg:"-P" help:"show phonetic guide"`
	}

	args.Config = *config.DefaultConfig()

	arg.MustParse(&args)

	if args.PhoneticHelp {
		printPhoneticGuide()
		return
	}

	r := generateSpeech(strings.Join(args.Input, " "), &args.Config, args.Phonetic)
	outputSpeech(r, args.Wav)
}

func generateSpeech(input string, cfg *config.Config, phonetic bool) *render.Render {
	var data [256]byte
	var i int
	for i = 0; i < len(input); i++ {
		data[i] = byte(input[i])
	}

	if cfg.Debug {
		if phonetic {
			fmt.Printf("phonetic input: %s\n", string(data[:]))
		} else {
			fmt.Printf("text input: %s\n", string(data[:]))
		}
	}

	if !phonetic {
		data[i] = '['

		rec := reciter.Reciter{}

		if !rec.TextToPhonemes(data[:], cfg) {
			os.Exit(1)
		}
		if cfg.Debug {
			fmt.Printf("phonetic input: %s\n", data)
		}
	} else {
		data[i] = '\x9b'
	}

	sam := sammain.Sam{
		Config: cfg,
	}
	sam.SetInput(data)
	if !sam.SAMMain() {
		// PrintUsage()
		os.Exit(2)
	}

	r := render.Render{
		Buffer: make([]byte, 22050*10),
	}

	sam.PrepareOutput(&r)

	return &r
}

func outputSpeech(r *render.Render, destination string) {
	if destination != "" {
		file, err := os.Create(destination)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(3)
		}
		_, err = wav.NewWriter(file, uint32(r.GetBufferLength()), 1, 22050, 8).Write(r.GetBuffer())
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(4)
		}
	} else {
		buf := &bytes.Buffer{}
		_, err := wav.NewWriter(buf, uint32(r.GetBufferLength()), 1, 22050, 8).Write(r.GetBuffer())
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

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/20))
		done := make(chan int)
		speaker.Play(beep.Seq(s, beep.Callback(func() {
			done <- 0
		})))
		<-done
	}
}

func printPhoneticGuide() {
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 1, '\t', 0)
	w.Write([]byte(
		`VOWELS		VOICED CONSONANTS
IY	f(ee)t	R	red
IH	p(i)n	L	allow
EH	beg	W	away
AE	Sam	W	whale
AA	pot	Y	you
AH	b(u)dget	M	Sam
AO	t(al)k	N	man
OH	cone	NX	so(ng)
UH	book	B	bad
UX	l(oo)t	D	dog
ER	bird	G	again
AX	gall(o)n	J	judge
IX	dig(i)t	Z	zoo
		ZH	plea(s)ure
		V	seven
		DH	(th)en

DIPHTHONGS		UNVOICED CONSONANTS
EY	m(a)de	S	Sam
AY	h(igh)	Sh	fish
OY	boy	F	fish
AW	h(ow)	TH	thin
OW	slow	P	poke
UW	crew	T	talk
		K	cake
		CH	speech
		/H	(h)ead

SPECIAL PHONEMES
UL	sett(le) (=AXL)
UM	astron(omy) (=AXM)
UN	functi(on) (=AXN)
Q	kitt-en (glottal stop)
`))
	w.Flush()
}
