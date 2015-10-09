package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/rsc/qr"
	"fmt"
	"image"
	"image/png"
	"os"
	"regexp"
	"runtime"
	"strings"
)

func main() {

	// maximize CPU usage for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	// open input file
	inFile, err := os.Open("input.csv")
	if err != nil {
		panic(err)
	}
	// close file on exit and check for its returned error
	defer func() {
		if err := inFile.Close(); err != nil {
			panic(err)
		}
	}()

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	// prepare regexp to determine filename outside the loop
	re, err := regexp.Compile(`Serie: (.*)`)

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.Replace(text, ";", "\n", -1)

		// Parse filename
		filename := re.FindStringSubmatch(text)[1]

		// Encode string to QR codes
		// qr.H = 65% redundant level
		// see https://godoc.org/code.google.com/p/rsc/qr#Level
		code, err := qr.Encode(text, qr.H)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		imgByte := code.PNG()

		// convert byte to image for saving to file
		img, _, _ := image.Decode(bytes.NewReader(imgByte))

		//save the imgByte to file
		filename = filename + ".png"
		out, err := os.Create(filename)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = png.Encode(out, img)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// everything ok
		fmt.Println("QR code generated and saved to " + filename)

	}
}
