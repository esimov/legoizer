package drawer

import (
	"image"
	"image/color"
	"math"

	proc "github.com/esimov/legoizer/processor"
	"github.com/fogleman/gg"
)

type point struct {
	x, y float64
}

type lego struct {
	*gg.Context
	point
}

type Quantizer struct {
	proc.Quant
}

func (quant *Quantizer) Init() {

}

func (quant *Quantizer) Process(input image.Image, nq int) image.Image {
	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()

	cellSize := 20

	dc := gg.NewContext(dx, dy)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	quantified := quant.Quantize(input, nq)
	nrgbaImg := convertToNRGBA64(quantified)

	for x := 0; x < dx; x += cellSize {
		for y := 0; y < dy; y += cellSize {
			xx := x + (cellSize / 2)
			yy := y + (cellSize / 2)
			if xx < dx && yy < dy {
				//r, g, b, a := nrgbaImg.At(xx , yy).RGBA()

				var subNRGBA = convertToNRGBA64(nrgbaImg.SubImage(image.Rect(x, y, x + cellSize, y + cellSize)))
				cell := getAvgColor(subNRGBA)

				legoPiece := createLegoElement(dc, float64(x), float64(y), float64(xx), float64(yy), float64(cellSize), cell)
				drawBorders(legoPiece)
			}
		}
	}
	newImg := dc.Image()
	return newImg
}

// Create the lego piece
func createLegoElement(dc *gg.Context, x, y, xx, yy, cellSize float64, c color.NRGBA64) *lego {
	// Background
	dc.DrawRectangle(x, y, x + cellSize, y + cellSize)
	dc.SetRGBA(float64(c.R/255 ^ 0xff), float64(c.G/255 ^ 0xff), float64(c.B/255 ^ 0xff), 0.8)
	dc.Fill()

	// Create the shadow effect
	dc.Push()
	grad := gg.NewRadialGradient(x, y, cellSize/2, xx, yy, cellSize)
	grad.AddColorStop(0, color.RGBA{177, 177, 177, 177})
	grad.AddColorStop(1, color.RGBA{255, 255, 255, 255})

	dc.SetFillStyle(grad)
	dc.DrawCircle(float64(xx-1), float64(yy-1), cellSize / 2 - math.Sqrt(cellSize))
	dc.Fill()

	grad = gg.NewRadialGradient(x, y, cellSize/2, xx, yy, cellSize)
	grad.AddColorStop(0, color.RGBA{0, 0, 0, 255})
	grad.AddColorStop(1, color.RGBA{155, 155, 155, 255})

	dc.SetFillStyle(grad)
	dc.StrokePreserve()
	dc.DrawCircle(float64(xx+1), float64(yy+1), cellSize / 2 - math.Sqrt(cellSize))
	dc.Fill()
	dc.Pop()

	// Draw circles
	dc.DrawCircle(xx, yy, float64(cellSize / 2) - math.Sqrt(float64(cellSize)))
	dc.SetRGBA(float64(c.R/255 ^ 0xff), float64(c.G/255 ^ 0xff), float64(c.B/255 ^ 0xff), float64(c.A/255))
	dc.Fill()

	return &lego {
		dc,
		point{x: x, y: y},
	}
}

func drawBorders(lego *lego) {
	var (
		cellSize float64 = 20
		dc = lego
		x = lego.x
		y = lego.y
	)
	// Bottom line
	dc.SetColor(color.RGBA{177, 177, 177, 255})
	dc.SetLineWidth(1)
	dc.MoveTo(x, y + cellSize)
	dc.LineTo(x + cellSize, y + cellSize)
	dc.ClosePath()
	dc.Stroke()

	// Right line
	dc.MoveTo(x + cellSize, y + cellSize)
	dc.LineTo(x + cellSize, y - cellSize)
	dc.ClosePath()
	dc.Stroke()
}

func checkNeighbors(dc *gg.Context) {

}

// Get the average color of a cell
func getAvgColor(img *image.NRGBA64) color.NRGBA64 {
	var (
		bounds  = img.Bounds()
		r, g, b int
	)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			var c = img.NRGBA64At(x, y)
			r += int(c.R)
			g += int(c.G)
			b += int(c.B)
		}
	}

	return color.NRGBA64{
		R: uint16(r / (bounds.Dx() * bounds.Dy())),
		G: uint16(g / (bounds.Dx() * bounds.Dy())),
		B: uint16(b / (bounds.Dx() * bounds.Dy())),
		A: 255,
	}
}


// Converts an image.Image into an image.NRGBA64.
func convertToNRGBA64(img image.Image) *image.NRGBA64 {
	var (
		bounds = img.Bounds()
		nrgba  = image.NewNRGBA64(bounds)
	)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			nrgba.Set(x, y, img.At(x, y))
		}
	}
	return nrgba
}

// Returns the smallest number between two numbers.
func min(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

// Returns the biggest number between two numbers.
func max(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}