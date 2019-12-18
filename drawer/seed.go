package drawer

import (
	"image"
	"image/color"
	"math"
)

type prng struct {
	a    int
	m    int
	rand int
	div  float64
}

// noise apply a noise factor to the source image
func noise(amount int, pxl image.Image, w, h int) *image.NRGBA64 {
	noiseImg := image.NewNRGBA64(image.Rect(0, 0, w, h))
	prng := &prng{
		a:    16807,
		m:    0x7fffffff,
		rand: 1.0,
		div:  1.0 / 0x7fffffff,
	}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			noise := (prng.randomSeed() - 0.1) * float64(amount)
			r, g, b, a := pxl.At(x, y).RGBA()
			rf, gf, bf := float64(r>>8), float64(g>>8), float64(b>>8)

			// Check if color do not overflow the maximum limit after noise has been applied
			if math.Abs(rf+noise) < 255 && math.Abs(gf+noise) < 255 && math.Abs(bf+noise) < 255 {
				rf += noise
				gf += noise
				bf += noise
			}
			r2 := max(0, min(255, uint8(rf)))
			g2 := max(0, min(255, uint8(gf)))
			b2 := max(0, min(255, uint8(bf)))
			noiseImg.Set(x, y, color.RGBA{R: r2, G: g2, B: b2, A: uint8(a)})
		}
	}
	return noiseImg
}

// nextLongRand generates a new random number based on the provided seed.
func (prng *prng) nextLongRand(seed int) int {
	lo := prng.a * (seed & 0xffff)
	hi := prng.a * (seed >> 16)
	lo += (hi & 0x7fff) << 16

	if lo > prng.m {
		lo &= prng.m
		lo++
	}
	lo += hi >> 15
	if lo > prng.m {
		lo &= prng.m
		lo++
	}
	return lo
}

// randomSeed returns a new random number.
func (prng *prng) randomSeed() float64 {
	prng.rand = prng.nextLongRand(prng.rand)
	return float64(prng.rand) * prng.div
}

// min returns the smallest number between two numbers.
func min(x, y uint8) uint8 {
	if x < y {
		return x
	}
	return y
}

// max returns the biggest number between two numbers.
func max(x, y uint8) uint8 {
	if x > y {
		return x
	}
	return y
}
