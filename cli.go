package main

import (
	"image"
	"os"
	"log"
	"image/png"
	_ "image/jpeg"
	"fmt"
	"github.com/esimov/legoizer/drawer"
	"path/filepath"
)

var quant drawer.Quantizer = drawer.Quantizer{}

func main() {
	absPath, _ := filepath.Abs("./gopher2.jpg")
	img, err := loadImage(absPath)
	if err != nil {
		fmt.Printf("Failed to open image '%v'\n", img)
		os.Exit(1)
	}
	res := quant.Process(img, 32)
	generateImage(res)
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
func generateImage(input image.Image) error {
	fq, err := os.Create("output.png")
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
