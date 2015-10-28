package main

import (
	"bufio"
	"code.google.com/p/rsc/qr"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"
)

var (
	seperator *string
	label     *int
	filename  *string
	outdir    *string
	scale     int
	black     *image.RGBA
	white     *image.RGBA
)

func init() {
	seperator = flag.String("s", ";", "seperator used in the CSV")

	label = flag.Int("l", 0, "the column to use as filename")

	outdir = flag.String("o", ".", "directory for the generated PNG files")

	filename = flag.String("i", "input.csv", "input file to process")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Generates a QR code for each row in the specified CSV file.\n\n")
		flag.PrintDefaults()
	}

	flag.IntVar(&scale, "sc", 10, "pixel size of a dot in the QR code")

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

	r := csv.NewReader(bufio.NewReader(inFile))
	r.Comma = ';'
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	// prepare images to draw white an black rectangles
	black = image.NewRGBA(image.Rect(0, 0, scale, scale))
	white = image.NewRGBA(image.Rect(0, 0, scale, scale))

	cB := color.RGBA{0x0, 0x0, 0x0, 0xFF}
	cW := color.RGBA{0xFF, 0xFF, 0xFF, 0x0}

	draw.Draw(black, black.Bounds(), &image.Uniform{cB}, image.ZP, draw.Src)
	draw.Draw(white, white.Bounds(), &image.Uniform{cW}, image.ZP, draw.Src)

	var headers []string
	pool := make(chan bool, 200/scale)

	var wg sync.WaitGroup
	for i, v := range records {
		if i == 0 {
			// first row are the headings
			headers = v
		} else {
			wg.Add(1)
			pool <- true
			go generateQr(headers, v, pool, &wg)
		}
	}
	wg.Wait()
}

func generateQr(headers []string, row []string, pool chan bool, wg *sync.WaitGroup) {
	filename := row[*label]
	var text string

	//build the output string
	for index, value := range row {
		text += headers[index] + ": " + value + "\n"
	}

	// Encode string to QR codes
	code, err := qr.Encode(text, qr.M)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// convert code to rgba
	b := code.Size
	img := image.NewRGBA(image.Rect(0, 0, b*scale, b*scale))

	for x := 0; x < b; x++ {
		for y := 0; y < b; y++ {
			if code.Black(x, y) {
				draw.Draw(img, image.Rect(x*scale, y*scale, (x+1)*scale, (y+1)*scale), black, black.Bounds().Min, draw.Src)
			} else {
				draw.Draw(img, image.Rect(x*scale, y*scale, (x+1)*scale, (y+1)*scale), white, white.Bounds().Min, draw.Src)
			}
		}
	}

	// verify output dir
	if _, err := os.Stat(*outdir); err != nil {
		os.Mkdir(*outdir, 0666)
	}
	//save the imgByte to file
	filename = *outdir + "/" + filename + ".png"
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
	<-pool
	(*wg).Done()
}
