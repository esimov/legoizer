package drawer

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	proc "github.com/esimov/legoizer/processor"
	"github.com/fogleman/gg"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	_1x1 = iota
	_2x1
	_3x1
	_4x1
	_6x1
	_2x2
	_3x2
	_4x2
	_6x2
)

type point struct {
	x, y float64
}

type lego struct {
	point
	cellSize  float64
	cellColor color.NRGBA64
}

type context struct {
	*gg.Context
}

type Quantizer struct {
	proc.Quant
}

type legoIndexes struct {
	idx int
	idy int
}

var legos []legoIndexes

var (
	threshold   uint16 = 127
	legoMaxRows        = 3
	legoMaxCols        = 2
	idx, idy           = 1, 1
)

// Process is the main function responsible to generate the lego bricks based on the provided source image.
func (quant *Quantizer) Process(input image.Image, nq int, cs int) image.Image {
	rand.Seed(time.Now().UTC().Unix())
	var (
		legoType                 int
		cellSize                 int
		current, total, progress float64
	)

	dx, dy := input.Bounds().Dx(), input.Bounds().Dy()
	imgRatio := func(w, h int) float64 {
		var ratio float64
		if w > h {
			ratio = float64((w / h) * w)
		} else {
			ratio = float64((h / w) * h)
		}
		return ratio
	}

	if cs == 0 {
		cellSize = int(round(float64(imgRatio(dx, dy)) * 0.015))
	} else {
		cellSize = cs
	}
	quantified := quant.Quantize(input, nq)
	nrgbaImg := convertToNRGBA64(quantified)

	dc := &context{gg.NewContext(dx, dy)}
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	total = math.Floor(float64(dx * dy / (cellSize * cellSize)))
	for x := 0; x < dx; x += cellSize {
		// Reset Y index after each row
		idy = 1
		for y := 0; y < dy; y += cellSize {
			xx := x + (cellSize / 2)
			yy := y + (cellSize / 2)
			current = math.Floor(float64(idx * dy / cellSize))

			if xx < dx && yy < dy {
				subImg := nrgbaImg.SubImage(image.Rect(x, y, x+cellSize, y+cellSize)).(*image.NRGBA64)
				cellColor := getAvgColor(subImg)

				lego := dc.getCurrentLego(nrgbaImg, float64(x), float64(y), float64(cellSize))
				rows, cols := dc.checkNeighbors(lego, nrgbaImg)

				switch {
				case rows == 1 && cols == 1:
					legoType = _1x1
				case rows == 2 && cols == 1:
					legoType = _2x1
				case rows == 3 && cols == 1:
					legoType = _3x1
				case rows == 4 && cols == 1:
					legoType = _4x1
				case rows == 6 && cols == 1:
					legoType = _6x1
				case rows == 2 && cols == 2:
					legoType = _2x2
				case rows == 3 && cols == 2:
					legoType = _3x2
				case rows == 4 && cols == 2:
					legoType = _4x2
				case rows == 6 && cols == 2:
					legoType = _6x2
				}
				dc.generateLegoSet(float64(x), float64(y), float64(xx), float64(yy), float64(cellSize), idx, idy, cellColor, legoType)
			}
			idy++
		}
		if current < total {
			progress = math.Floor(float64(current/total) * 100.0)
			showProgress(progress)
		}
		idx++
	}
	if progress < 100 {
		showProgress(100)
	}
	img := dc.Image()
	noisyImg := noise(10, img, img.Bounds().Dx(), img.Bounds().Dy())

	return noisyImg
}

// createLegoPiece creates the lego piece
func (dc *context) createLegoPiece(x, y, xx, yy, cellSize float64, c color.NRGBA64) *lego {
	// Brightness factor
	var bf = 1.0005
	// Background
	dc.DrawRectangle(x, y, x+cellSize, y+cellSize)
	dc.SetRGBA(float64(c.R/255^0xff)*bf, float64(c.G/255^0xff)*bf, float64(c.B/255^0xff)*bf, 1)
	dc.Fill()
	// Create a shadow effect
	dc.Push()
	// Top circle
	grad := gg.NewRadialGradient(xx, yy, cellSize/2, x, y, 0)
	grad.AddColorStop(0, color.RGBA{177, 177, 177, 0})
	grad.AddColorStop(1, color.RGBA{255, 255, 255, 177})

	dc.SetFillStyle(grad)
	dc.DrawCircle(float64(xx-1), float64(yy-1), cellSize/2-math.Sqrt(cellSize))
	dc.Fill()

	// Bottom circle
	grad = gg.NewRadialGradient(xx, yy, cellSize/2, x, y, 0)
	grad.AddColorStop(0, color.RGBA{0, 0, 0, 177})

	r, g, b := c.R/255^0xff, c.G/255^0xff, c.B/255^0xff
	if r > threshold || g > threshold || b > threshold {
		grad.AddColorStop(1, color.RGBA{0, 0, 0, 255})
	} else {
		grad.AddColorStop(1, color.RGBA{177, 177, 177, 255})
	}

	dc.SetFillStyle(grad)
	dc.StrokePreserve()
	dc.DrawCircle(float64(xx+1), float64(yy+1), cellSize/2-math.Sqrt(cellSize))
	dc.Fill()
	dc.Pop()

	// Draw the main circle
	dc.DrawCircle(xx, yy, float64(cellSize/2)-math.Sqrt(float64(cellSize)))
	dc.SetRGBA(float64(c.R/255^0xff), float64(c.G/255^0xff), float64(c.B/255^0xff), float64(c.A/255))
	dc.Fill()

	return &lego{
		point{x: x, y: y},
		cellSize,
		c,
	}
}

// generateLegoSet creates the lego block constituted by the lego pieces.
// This function traces the lego borders on the intersection of columns and rows.
func (dc *context) generateLegoSet(x, y, xx, yy, cellSize float64, idx, idy int, c color.NRGBA64, legoType int) *lego {
	var rows, cols int
	switch legoType {
	case _1x1:
		rows, cols = 1, 1
	case _2x1:
		rows, cols = 2, 1
	case _3x1:
		rows, cols = 3, 1
	case _4x1:
		rows, cols = 4, 1
	case _6x1:
		rows, cols = 6, 1
	case _2x2:
		rows, cols = 2, 2
	case _3x2:
		rows, cols = 3, 2
	case _4x2:
		rows, cols = 4, 2
	case _6x2:
		rows, cols = 6, 2
	}

	drawLeftBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{177, 177, 177, 177})
		dc.SetLineWidth(0.10)
		dc.MoveTo(x, y)
		dc.LineTo(x, y+cellSize)
		dc.ClosePath()
		dc.Stroke()
	}
	drawTopBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{177, 177, 177, 177})
		dc.SetLineWidth(0.05)
		dc.MoveTo(x, y)
		dc.LineTo(x+cellSize, y)
		dc.ClosePath()
		dc.Stroke()
	}
	drawRightBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{0, 0, 0, 177})
		dc.SetLineWidth(0.15)
		dc.MoveTo(x+cellSize, y)
		dc.LineTo(x+cellSize, y+cellSize)
		dc.ClosePath()
		dc.Stroke()
	}
	drawBottomBorderLine := func(x, y float64) {
		dc.SetColor(color.RGBA{0, 0, 0, 177})
		dc.SetLineWidth(0.15)
		dc.MoveTo(x, y+cellSize)
		dc.LineTo(x+cellSize, y+cellSize)
		dc.ClosePath()
		dc.Stroke()
	}

	// Create the lego piece then trace the borders.
	dc.createLegoPiece(x, y, float64(xx), float64(yy), float64(cellSize), c)

	legoExists := findLegoIndex(legos, int(x), int(y))
	// Draw the borders only if index does not exists in the index table.
	if !legoExists {
		if idx%rows == 0 {
			drawLeftBorderLine(x-(cellSize*float64(rows))+cellSize+1, y)
			drawRightBorderLine(x, y)
		}
		if idy%cols == 0 {
			drawTopBorderLine(x, y-(cellSize*float64(cols))+cellSize+1)
			drawBottomBorderLine(x, y)
		}
	}

	return &lego{
		point{x: x, y: y},
		cellSize,
		c,
	}
}

// getCurrentLego returns the current lego's first pixel color.
// We don't need to get all the colors of the cell, since we are averaging the cell color.
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
	var lastIdx, lastIdy int = 1, 1
	var (
		cellSize            = lego.cellSize
		cellColor           = lego.cellColor
		x                   = lego.x
		y                   = lego.y
		ct                  = 7.0
		currentRowCellColor color.NRGBA64
	)

	rows, cols := 1, 1
	legoWidth := random(rows, legoMaxRows)
	legoHeight := random(cols, legoMaxCols)

	xi := int(x)
	yi := int(y)

	// Rows
	for i := 1; ; i++ {
		if i > legoWidth {
			break
		}
		if xi*i < dc.Width() && yi*i < dc.Height() {
			nextCell := neighborCell.SubImage(image.Rect(xi*i, yi, xi*i+int(cellSize), yi+int(cellSize))).(*image.NRGBA64)
			nextCellColor := getAvgColor(nextCell)

			// Because the next cell average color might differ from the current cell color even with a small amount,
			// we have to check if the current cell color is approximately identical with the neighboring cells.
			c1 := colorful.Color{
				R: float64(cellColor.R >> 8),
				G: float64(cellColor.G >> 8),
				B: float64(cellColor.B >> 8),
			}
			c2 := colorful.Color{
				R: float64(nextCellColor.R >> 8),
				G: float64(nextCellColor.G >> 8),
				B: float64(nextCellColor.B >> 8),
			}

			colorThreshold := c1.DistanceCIE94(c2)
			if colorThreshold > ct {
				currentRowCellColor = cellColor
				lastIdx = xi * i
				break
			}
		}
		rows++
	}

	// Columns
	for i := 1; ; i++ {
		if i > legoHeight {
			break
		}
		if xi*i < dc.Width() && yi*i < dc.Height() {
			nextCell := neighborCell.SubImage(image.Rect(xi, yi*i, xi+int(cellSize), yi*i+int(cellSize))).(*image.NRGBA64)
			nextCellColor := getAvgColor(nextCell)

			c1 := colorful.Color{
				R: float64(cellColor.R >> 8),
				G: float64(cellColor.G >> 8),
				B: float64(cellColor.B >> 8),
			}
			c2 := colorful.Color{
				R: float64(nextCellColor.R >> 8),
				G: float64(nextCellColor.G >> 8),
				B: float64(nextCellColor.B >> 8),
			}

			colorThreshold := c1.DistanceCIE94(c2)
			if colorThreshold > ct || currentRowCellColor.R != cellColor.R {
				lastIdy = yi * i
				break
			}
		}
		cols++
	}
	// No lego piece with 5 rows
	if rows == 5 {
		rows = 4
	}
	// Save the generated lego indexes into the index table.
	// We need verify if the lego borders have been traced based on the index value.
	if !findLegoIndex(legos, lastIdx, lastIdy) {
		legos = append(legos, legoIndexes{lastIdx, lastIdy})
	}

	return rows, cols
}

// getAvgColor get the average color of a cell
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
		R: maxUint16(0, minUint16(65535, uint16(r/(bounds.Dx()*bounds.Dy())))),
		G: maxUint16(0, minUint16(65535, uint16(g/(bounds.Dx()*bounds.Dy())))),
		B: maxUint16(0, minUint16(65535, uint16(b/(bounds.Dx()*bounds.Dy())))),
		A: 255,
	}
}

// convertToNRGBA64 converts an image.Image into an image.NRGBA64.
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

// findLegoIndex check if the processed lego index exists in the lego table.
func findLegoIndex(legos []legoIndexes, ix, iy int) bool {
	for i := 0; i < len(legos); i++ {
		idx, idy := legos[i].idx, legos[i].idy
		if idx == ix && idy == iy {
			return true
		}
	}
	return false
}

// round number down.
func round(x float64) float64 {
	return math.Floor(x)
}

// random generates a random number between min & max.
func random(min, max int) int {
	return rand.Intn(max-min) + min
}

// minUint16 returns the smallest number between two uint16 numbers.
func minUint16(x, y uint16) uint16 {
	if x < y {
		return x
	}
	return y
}

// maxUint16 returns the biggest number between two uint16 numbers.
func maxUint16(x, y uint16) uint16 {
	if x > y {
		return x
	}
	return y
}

// showProgress show the progress status.
func showProgress(progress float64) {
	fmt.Printf("  \r  %v%% [", progress)
	for p := 0; p < 100; p += 3 {
		if progress > float64(p) {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] \r")
}
