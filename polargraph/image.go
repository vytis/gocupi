package polargraph

// Draws a series of coordinates to an image

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
)

func DrawToImageExact(imageName string, widthMM float64, heightMM float64, plotCoords <-chan Coordinate) {
	paddingMM := 20.0

	dpi := 150.0
	mmToPixels := (dpi / 25.4)

	paddingPx := paddingMM * mmToPixels

	canvasWidth := (widthMM + paddingMM*2) * mmToPixels
	canvasHeight := (heightMM + paddingMM*2) * mmToPixels

	fmt.Println("Paper ", widthMM, "mm x ", heightMM, "mm. Image ", canvasWidth, "px x", canvasHeight, "px")

	image := image.NewRGBA(image.Rect(0, 0, int(canvasWidth), int(canvasHeight)))

	// draw border
	drawLineExact(Coordinate{X: paddingPx, Y: paddingPx}, Coordinate{X: canvasWidth - paddingPx, Y: paddingPx}, image)
	drawLineExact(Coordinate{X: canvasWidth - paddingPx, Y: paddingPx}, Coordinate{X: canvasWidth - paddingPx, Y: canvasHeight - paddingPx}, image)
	drawLineExact(Coordinate{X: canvasWidth - paddingPx, Y: canvasHeight - paddingPx}, Coordinate{X: paddingPx, Y: canvasHeight - paddingPx}, image)
	drawLineExact(Coordinate{X: paddingPx, Y: canvasHeight - paddingPx}, Coordinate{X: paddingPx, Y: paddingPx}, image)

	// plot each point in the image
	previousPoint := Coordinate{X: paddingPx, Y: paddingPx}
	for point := range plotCoords {
		//image.Set(int(point.X-minPoint.X), int(-(point.Y-minPoint.Y)+2*maxPoint.Y), color.RGBA{0, 0, 0, 255})
		next := Coordinate{X: point.X*mmToPixels + paddingPx, Y: point.Y*mmToPixels + paddingPx, PenUp: point.PenUp}
		drawLineExact(previousPoint, next, image)

		previousPoint = next
	}

	file, err := os.OpenFile(imageName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err = png.Encode(file, image); err != nil {
		panic(err)
	}
}

// Draw a line, from http://41j.com/blog/2012/09/bresenhams-line-drawing-algorithm-implemetations-in-go-and-c/
func drawLineExact(start Coordinate, end Coordinate, image *image.RGBA) {
	start_x := int(start.X)
	start_y := int(start.Y)
	end_x := int(end.X)
	end_y := int(end.Y)
	var lineColor color.RGBA
	if end.PenUp {
		lineColor = color.RGBA{0, 255, 0, 255}
	} else {
		lineColor = color.RGBA{0, 0, 0, 255}
	}

	/*
		image.Set(end_x+1, end_y+1, color.RGBA{255, 0, 0, 128})
		image.Set(end_x+1, end_y-1, color.RGBA{255, 0, 0, 128})
		image.Set(end_x-1, end_y+1, color.RGBA{255, 0, 0, 128})
		image.Set(end_x-1, end_y-1, color.RGBA{255, 0, 0, 128})
	*/
	// Bresenham's
	cx := start_x
	cy := start_y

	dx := end_x - cx
	dy := end_y - cy
	if dx < 0 {
		dx = 0 - dx
	}
	if dy < 0 {
		dy = 0 - dy
	}

	var sx int
	var sy int
	if cx < end_x {
		sx = 1
	} else {
		sx = -1
	}
	if cy < end_y {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	var n int
	for n = 0; n < 10000; n++ {

		image.Set(cx, cy, lineColor)
		if (cx == end_x) && (cy == end_y) {
			return
		}
		e2 := 2 * err
		if e2 > (0 - dy) {
			err = err - dy
			cx = cx + sx
		}
		if e2 < dx {
			err = err + dx
			cy = cy + sy
		}
	}
}
