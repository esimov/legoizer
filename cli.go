package main

import (
	"image"
	"os"
	"log"
	"image/png"
	_ "image/jpeg"
	"fmt"
	"flag"
	"github.com/esimov/legoizer/drawer"
	"time"
)

var (
	quant drawer.Quantizer = drawer.Quantizer{}

	inPath    = flag.String("in", "", "Input path")
	outPath   = flag.String("out", "", "Output path")
	legoSize  = flag.Int("size", 0, "Lego size")
	colors	  = flag.Int("colors", 128, "Number of colors")
)

func main() {
	// Parse the command-line arguments
	flag.Parse()

	img, err := loadImage(*inPath)
	if err != nil {
		fmt.Printf("Failed to open image '%v'\n", img)
		os.Exit(1)
	}

	fmt.Println("Generating legoized image...")
	now := time.Now()

	res := quant.Process(img, *colors, *legoSize)
	generateImage(res, *outPath)

	since := time.Since(now)
	fmt.Println("\n Done✓")
	fmt.Printf("Generated in: %.2fs\n", since.Seconds())
}

// Loads an image from a file path.
func loadImage(path string) (image.Image, error) {
	sf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer sf.Close()
	img, _, err := image.Decode(sf)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Generate the resulted image.
func generateImage(input image.Image, outPath string) error {
	fq, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer fq.Close()

	if err = png.Encode(fq, input); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}