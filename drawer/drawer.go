package drawer

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	proc "github.com/esimov/legoizer/processor"
	"github.com/fogleman/gg"
)

type point struct {
	x, y float64
}

type lego struct {
	*gg.Context
	point
	cellSize float64
	cellColor color.NRGBA64
}

type Quantizer struct {
	proc.Quant
}

var (
	threshold uint16 = 127
	legoMaxWidth, legoMaxHeight int = 4, 2
)

func (quant *Quantizer) Init() {

}

func (quant *Quantizer) Process(input image.Image, nq int) image.Image {
	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()

	cellSize := 30

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

				var subImg = convertToNRGBA64(nrgbaImg.SubImage(image.Rect(x, y, x + cellSize, y + cellSize)))
				cellColor := getAvgColor(subImg)
				createLegoElement(dc, float64(x), float64(y), float64(xx), float64(yy), float64(cellSize), cellColor)
			}
		}
	}

	for x := 0; x < dx; x += cellSize {
		for y := 0; y < dy; y += cellSize {
			xx := x + (cellSize / 2)
			yy := y + (cellSize / 2)
			if xx < dx && yy < dy {
				lego := getCurrentLego(dc, nrgbaImg, float64(x), float64(y), float64(cellSize))
				checkNeighbors(dc, lego, nrgbaImg)
			}
		}
	}
	newImg := dc.Image()
	return newImg
}

// Create the lego piece
func createLegoElement(dc *gg.Context, x, y, xx, yy, cellSize float64, c color.NRGBA64) *lego {
	// Brightness factor
	var bf float64 = 1.0003

	// Background
	dc.DrawRectangle(x, y, x + cellSize, y + cellSize)
	dc.SetRGBA(float64(c.R/255 ^ 0xff) * bf, float64(c.G/255 ^ 0xff) * bf, float64(c.B/255 ^ 0xff) * bf, 1)
	dc.Fill()

	// Create the shadow effect
	dc.Push()
	// Top circle
	grad := gg.NewRadialGradient(xx, yy, cellSize/2, x, y, 0)
	grad.AddColorStop(0, color.RGBA{177, 177, 177, 0})
	grad.AddColorStop(1, color.RGBA{255, 255, 255, 177})

	dc.SetFillStyle(grad)
	dc.DrawCircle(float64(xx-1), float64(yy-1), cellSize / 2 - math.Sqrt(cellSize))
	dc.Fill()

	// Bottom circle
	grad = gg.NewRadialGradient(xx, yy, cellSize/2, x, y, 0)
	grad.AddColorStop(0, color.RGBA{0, 0, 0, 177})

	r, g, b := c.R/255 ^ 0xff, c.G/255 ^ 0xff, c.B/255 ^ 0xff
	if r > threshold || g > threshold || b > threshold {
		grad.AddColorStop(1, color.RGBA{0, 0, 0, 255})
	} else {
		grad.AddColorStop(1, color.RGBA{177, 177, 177, 255})
	}

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
		cellSize,
		c,
	}
}

func getCurrentLego(dc *gg.Context, cell *image.NRGBA64, x, y, cellSize float64) *lego {
	// Get the first pixel color
	var col = cell.NRGBA64At(int(x), int(y))

	return &lego{
		dc,
		point{x: x, y: y},
		cellSize,
		col,
	}
}

func drawBorders(lego *lego) {
	var (
		dc = lego
		cellSize = lego.cellSize
		x = lego.x
		y = lego.y
	)
	// Bottom line
	dc.SetColor(color.RGBA{177, 177, 177, 177})
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

func checkNeighbors(dc *gg.Context, lego *lego, neighborCell *image.NRGBA64) {
	var (
		cellSize = lego.cellSize
		cellColor = lego.cellColor
		x = lego.x
		y = lego.y
		//bf float64 = 1.0003
	)

	drawTopBorderLine := func(x, y float64, c color.NRGBA64) {
		dc.SetColor(color.RGBA{255, 255, 255, 255})
		dc.SetLineWidth(0.5)
		dc.MoveTo(x, y)
		dc.LineTo(x + cellSize, y)
		dc.ClosePath()
		dc.Stroke()
	}

	drawRightBorderLine := func(x, y float64, c color.NRGBA64) {
		//dc.SetColor(color.RGBA{uint8(float64(c.R) * bf), uint8(float64(c.G) * bf), uint8(float64(c.B) * bf), 255})
		dc.SetColor(color.RGBA{0, 0, 0, 255})
		dc.SetLineWidth(0.5)
		dc.MoveTo(x + cellSize, y)
		dc.LineTo(x + cellSize, y + cellSize)
		dc.ClosePath()
		dc.Stroke()
	}

	legoWidth := random(1, legoMaxWidth)
	//legoHeight := random(1, legoMaxHeight)

	xi := int(x)
	yi := int(y)
	for i := 1; ; i++ {
		if i > legoWidth {
			break
		}
		if xi*i < dc.Width() && yi*i < dc.Height() {
			rightCell := convertToNRGBA64(neighborCell.SubImage(image.Rect(xi*i, yi, xi*i + int(cellSize), yi + int(cellSize))))
			nextCellColor := getAvgColor(rightCell)

			if cellColor.R == nextCellColor.R &&
				cellColor.G == nextCellColor.G &&
				cellColor.B == nextCellColor.B {

				drawTopBorderLine(x * float64(i), y * float64(i), cellColor)
			} else {
				drawRightBorderLine(x * float64(i), y * float64(i), cellColor)
				break
			}
		}
	}

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

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - min) + min
}