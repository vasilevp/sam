package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/exploser/sam/config"
)

func legacyMain() {
	var phonetic = false

	var input string

	if len(os.Args) <= 1 {
		printUsage()
		os.Exit(1)
	}

	wavfilename := ""

	cfg := config.DefaultConfig()

	i := 1
	for i < len(os.Args) {
		if os.Args[i][0] != '-' {
			input += os.Args[i] + " "
		} else {
			switch os.Args[i][1:] {
			case "wav":
				wavfilename = os.Args[i+1]
				i++
			case "sing":
				cfg.EnableSingmode()
			case "phonetic":
				phonetic = true
			case "debug":
				cfg.Debug = true
			case "pitch":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				cfg.SetPitch(byte(val))
				i++
			case "speed":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				cfg.SetSpeed(byte(val))
				i++
			case "mouth":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				cfg.SetMouth(byte(val))
				i++
			case "throat":
				val, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}
				cfg.SetThroat(byte(val))
				i++
			default:
				printUsage()
				os.Exit(1)
			}
		}

		i++
	}

	input = strings.ToUpper(input)

	r := generateSpeech(input, cfg, phonetic)

	outputSpeech(r, wavfilename)
}

func printUsage() {
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
	printPhoneticGuide()
}
