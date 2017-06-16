package drawer

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	proc "github.com/esimov/legoizer/processor"
	"github.com/fogleman/gg"
	//"fmt"
)
const (
	_1x1 = iota
	_1x2
	_1x3
	_1x4
	_1x6
	_2x2
	_2x3
	_2x4
	_2x6
)
type point struct {
	x, y float64
}

type lego struct {
	point
	cellSize float64
	cellColor color.NRGBA64
}

type context struct {
	*gg.Context
}

type Quantizer struct {
	proc.Quant
}

var (
	threshold uint16 = 127
	legoMaxPiece, legoMaxHeight int = 4, 3
	idx, idy = 1, 1
)

func (quant *Quantizer) Process(input image.Image, nq int) image.Image {
	rand.Seed(time.Now().UTC().Unix())
	var legoType int

	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()
	dc := &context{gg.NewContext(dx, dy)}
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	cellSize := 30
	quantified := quant.Quantize(input, nq)
	nrgbaImg := convertToNRGBA64(quantified)

	for x := 0; x < dx; x += cellSize {
		idy = 1
		for y := 0; y < dy; y += cellSize {
			xx := x + (cellSize / 2)
			yy := y + (cellSize / 2)
			if xx < dx && yy < dy {
				var subImg = convertToNRGBA64(nrgbaImg.SubImage(image.Rect(x, y, x + cellSize, y + cellSize)))
				cellColor := getAvgColor(subImg)
				lego := dc.getCurrentLego(nrgbaImg, float64(x), float64(y), float64(cellSize))
				rows, cols := dc.checkNeighbors(lego, nrgbaImg)
				//fmt.Println(rows, " : ",  cols)
				switch {
				case rows == 1 && cols == 1 :
					legoType = _1x1
				case rows == 1 && cols == 2 :
					legoType = _1x2
				case rows == 1 && cols == 3 :
					legoType = _1x3
				case rows == 1 && cols == 4 :
					legoType = _1x4
				case rows == 1 && cols == 6 :
					legoType = _1x6
				case rows == 2 && cols == 2 :
					legoType = _2x2
				case rows == 2 && cols == 3 :
					legoType = _2x3
				case rows == 2 && cols == 4 :
					legoType = _2x4
				case rows == 2 && cols == 6 :
					legoType = _2x6
				}
				dc.generateLegoSet(float64(x), float64(y), float64(xx), float64(yy), float64(cellSize), idx, idy, cellColor, legoType)
			}
			idy++
		}
		idx++
	}

	/*for x := 0; x < dx; x += cellSize {
		for y := 0; y < dy; y += cellSize {
			xx := x + (cellSize / 2)
			yy := y + (cellSize / 2)
			if xx < dx && yy < dy {
				lego := getCurrentLego(dc, nrgbaImg, float64(x), float64(y), float64(cellSize))
				checkNeighbors(dc, lego, nrgbaImg)
			}
		}
	}*/
	newImg := dc.Image()
	return newImg
}

// Create the lego piece
func (dc *context) createLegoPiece(x, y, xx, yy, cellSize float64, c color.NRGBA64) *lego {
	// Brightness factor
	var bf float64 = 1.0005

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

	// Draw the main circle
	dc.DrawCircle(xx, yy, float64(cellSize / 2) - math.Sqrt(float64(cellSize)))
	dc.SetRGBA(float64(c.R/255 ^ 0xff), float64(c.G/255 ^ 0xff), float64(c.B/255 ^ 0xff), float64(c.A/255))
	dc.Fill()

	return &lego {
		point{x: x, y: y},
		cellSize,
		c,
	}
}

// Generate the lego block compounded by lego pieces
// This function trace the lego borders on the intersection of columns and rows.
func (dc *context) generateLegoSet(x, y, xx, yy, cellSize float64, idx, idy int, c color.NRGBA64, legoType int) *lego {
	var rows, cols int
	switch legoType {
	case _1x1 :
		rows, cols = 1, 1
	case _1x2 :
		rows, cols = 1, 2
	case _1x3 :
		rows, cols = 1, 3
	case _1x4 :
		rows, cols = 1, 4
	case _1x6 :
		rows, cols = 1, 6
	case _2x2 :
		rows, cols = 2, 2
	case _2x3 :
		rows, cols = 2, 3
	case _2x4 :
		rows, cols = 2, 4
	case _2x6 :
		rows, cols = 2, 6
	}

	drawLeftBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{177, 177, 177, 177})
		dc.SetLineWidth(0.10)
		dc.MoveTo(x, y)
		dc.LineTo(x, y + cellSize)
		dc.ClosePath()
		dc.Stroke()
	}

	drawTopBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{177, 177, 177, 177})
		dc.SetLineWidth(0.05)
		dc.MoveTo(x, y)
		dc.LineTo(x + cellSize, y)
		dc.ClosePath()
		dc.Stroke()
	}

	drawRightBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{0, 0, 0, 177})
		dc.SetLineWidth(0.15)
		dc.MoveTo(x + cellSize, y)
		dc.LineTo(x + cellSize, y + cellSize)
		dc.ClosePath()
		dc.Stroke()
	}

	drawBottomBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{0, 0, 0, 177})
		dc.SetLineWidth(0.15)
		dc.MoveTo(x, y + cellSize)
		dc.LineTo(x + cellSize, y + cellSize)
		dc.ClosePath()
		dc.Stroke()
	}

	// Create the lego piece, then trace the borders.
	dc.createLegoPiece(x, y, float64(xx), float64(yy), float64(cellSize), c)
	if idx % rows == 0 {
		drawLeftBorderLine(x - (cellSize * float64(rows)) + cellSize + 1, y)
		drawRightBorderLine(x, y)
	}
	if idy % cols == 0 {
		drawTopBorderLine(x, y - (cellSize * float64(cols)) + cellSize + 1)
		drawBottomBorderLine(x, y)
	}

	return &lego{
		point{x: x, y: y},
		cellSize,
		c,
	}
}

// Return the current lego's first pixel color.
// We don't need to get all the colors of the cell, since we averaged the cell color.
func (dc *context) getCurrentLego(cell *image.NRGBA64, x, y, cellSize float64) *lego {
	// Get the first pixel color
	var c = cell.NRGBA64At(int(x), int(y))

	return &lego{
		point{x: x, y: y},
		cellSize,
		c,
	}
}

// Check if the current lego color is identical with the neighbors color.
// Returns the number of rows and columns identical with the current lego, with the rows & columns representing the lego type.
func (dc *context) checkNeighbors(lego *lego, neighborCell *image.NRGBA64) (int, int) {
	var (
		cellSize = lego.cellSize
		cellColor = lego.cellColor
		x = lego.x
		y = lego.y
		//bf float64 = 1.0003
	)

	rows, cols := 1, 1
	legoWidth := random(1, legoMaxPiece)
	legoHeight := random(1, legoMaxHeight)

	xi := int(x)
	yi := int(y)

	// Rows
	for i := 1; ; i++ {
		if i > legoWidth - 1 {
			break
		}
		if xi*i < dc.Width() && yi*i < dc.Height() {
			nextCell := convertToNRGBA64(neighborCell.SubImage(image.Rect(xi*i, yi, xi*i + int(cellSize), yi + int(cellSize))))
			nextCellColor := getAvgColor(nextCell)

			if cellColor.R != nextCellColor.R &&
				cellColor.G != nextCellColor.G &&
				cellColor.B != nextCellColor.B {
				break
			}
		}
		rows++
	}

	// Columns
	for i := 1; ; i++ {
		if i > legoHeight - 1 {
			break
		}
		if xi*i < dc.Width() && yi*i < dc.Height() {
			nextCell := convertToNRGBA64(neighborCell.SubImage(image.Rect(xi, yi*i, xi + int(cellSize), yi*i + int(cellSize))))
			nextCellColor := getAvgColor(nextCell)

			if cellColor.R != nextCellColor.R &&
				cellColor.G != nextCellColor.G &&
				cellColor.B != nextCellColor.B {
				break
			}
		}
		cols++
	}
	return rows, cols
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
	return rand.Intn(max - min) + min
}