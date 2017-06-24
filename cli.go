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
	// Channel to signal the completion event
	done := make(chan struct{})

	// Parse the command-line arguments
	flag.Parse()

	img, err := loadImage(*inPath)
	if err != nil {
		fmt.Printf("Failed to open image '%v'\n", img)
		os.Exit(1)
	}

	fmt.Print("Rendering image...")
	now := time.Now()
	progress(done)

	func(done chan struct{}) {
		res := quant.Process(img, *colors, *legoSize)
		generateImage(res, *outPath)
		done <- struct{}{}
	}(done)

	since := time.Since(now)
	fmt.Println("\nDone✓")
	fmt.Printf("Rendered in: %.2fs\n", since.Seconds())
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

// Function to visualize the rendering progress
func progress(done chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 200)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Print(".")
			case <-done:
				ticker.Stop()
			}
		}
	}()
}