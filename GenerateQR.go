package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/rsc/qr"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"regexp"
	"strings"
)

var (
	seperator *string
	label     *string
	filename  *string
)

func init() {
	seperator = flag.String("s", ";", "seperator used in the CSV")

	label = flag.String("l", "Serie: ", "the label to remove in the first column to QR")

	filename = flag.String("i", "input.csv", "input file to process")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Generates a QR code for each row in the specified CSV file.\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	// read parameters
	flag.Parse()

	// open input file
	inFile, err := os.Open(*filename)
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
	re, err := regexp.Compile(*label + "(.*)")

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.Replace(text, *seperator, "\n", -1)

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
