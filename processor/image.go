package processor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sort"
	"strconv"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
)

const (
	CanvasMargin               = 10
	GridSize                   = 14
	GridShapePadding           = 4
	GridShapeThickness         = 2
	GridBorder                 = 1
	ColorDescriptionLeftMargin = 20
	ColorDescriptionMargin     = 10
	ColorDescriptionWidth      = 200
)

type ColorMap map[color.Color]ColorMapInfo

func drawSymbolAtPosition(colorMap ColorMapInfo, x int, y int, dst *image.RGBA) {
	// The symbol is rendered in black or white depending on the color luminance
	blackColor := color.RGBA{0, 0, 0, 255}
	whiteColor := color.RGBA{255, 255, 255, 255}
	r := colorMap.MappedColor.R
	g := colorMap.MappedColor.G
	b := colorMap.MappedColor.B
	col := color.RGBA{r, g, b, 255}
	luminance := 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b)
	highlightColor := blackColor
	if luminance < 128 {
		highlightColor = whiteColor
	}
	// Grid square rectangle
	rect := image.Rect(x, y, x+GridSize, y+GridSize)
	// Fill in the square
	draw.Draw(dst, rect, image.NewUniform(col), image.ZP, draw.Src)
	rect = rect.Inset(GridShapePadding)
	middleMargin := (GridSize-GridShapeThickness)/2 - GridShapePadding
	// I can't possibly prepare infinite amount of symbols... Just repeat them.
	sym := colorMap.Pattern % 6
	switch sym {
	case 0: // Filled square
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
	case 1: // Cross
		rect.Min.X += middleMargin
		rect.Max.X = rect.Min.X + GridShapeThickness
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
		rect.Min.X -= middleMargin
		rect.Min.Y += middleMargin
		rect.Max.X += middleMargin
		rect.Max.Y -= middleMargin
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
	case 2: // Open box
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
		rect = rect.Inset(GridShapeThickness)
		rect.Max.X += GridShapeThickness
		draw.Draw(dst, rect, image.NewUniform(col), image.ZP, draw.Src)
	case 3: // Vertical line
		rect.Min.X += middleMargin
		rect.Max.X = rect.Min.X + GridShapeThickness
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
	case 4: // Horizontal line
		rect.Min.Y += middleMargin
		rect.Max.Y = rect.Min.Y + GridShapeThickness
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
	case 5: // Corner
		draw.Draw(dst, rect, image.NewUniform(highlightColor), image.ZP, draw.Src)
		rect = rect.Inset(GridShapeThickness)
		rect.Max.X += GridShapeThickness
		rect.Max.Y += GridShapeThickness
		draw.Draw(dst, rect, image.NewUniform(col), image.ZP, draw.Src)
	}
}

func drawGrid(width int, height int, dst *image.RGBA) {
	gridHeight := (GridSize+GridBorder)*height + GridBorder
	gridWidth := (GridSize+GridBorder)*width + GridBorder
	blackColor := color.RGBA{0, 0, 0, 255}
	whiteColor := color.RGBA{255, 255, 255, 255}
	// Background fill
	draw.Draw(dst, dst.Bounds(), image.NewUniform(whiteColor), image.ZP, draw.Src)
	// Vertical lines
	for x := 0; x <= width; x++ {
		rX := CanvasMargin + (GridSize+GridBorder)*x
		rect := image.Rect(rX, CanvasMargin, rX+GridBorder, CanvasMargin+gridHeight)
		draw.Draw(dst, rect, image.NewUniform(blackColor), image.ZP, draw.Src)
	}
	// Horizontal lines
	for y := 0; y <= height; y++ {
		rY := CanvasMargin + (GridSize+GridBorder)*y
		rect := image.Rect(CanvasMargin, rY, CanvasMargin+gridWidth, rY+GridBorder)
		draw.Draw(dst, rect, image.NewUniform(blackColor), image.ZP, draw.Src)
	}
}

type SortNumberString []string

func (s SortNumberString) Len() int {
	return len(s)
}

func (s SortNumberString) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortNumberString) Less(i, j int) bool {
	iN, _ := strconv.Atoi(s[i])
	jN, _ := strconv.Atoi(s[j])
	return iN < jN
}

func drawColorLegends(width int, height int, dst *image.RGBA, colorMap ColorMap, fnt *truetype.Font) {
	leftOffset := CanvasMargin + (GridSize+GridBorder)*width + GridBorder + ColorDescriptionLeftMargin
	yOffset := 0
	// Prepare FreeType renderer
	ftCtx := freetype.NewContext()
	ftCtx.SetDst(dst)
	ftCtx.SetSrc(image.Black)
	ftCtx.SetFont(fnt)
	ftCtx.SetFontSize(16)
	ftCtx.SetHinting(freetype.NoHinting)
	ftCtx.SetDPI(72)
	ftCtx.SetClip(dst.Bounds())
	codeMap := make(map[string]ColorMapInfo)
	codeArray := make(SortNumberString, len(colorMap))
	idx := 0
	for _, colMap := range colorMap {
		codeMap[colMap.MappedColor.Code] = colMap
		codeArray[idx] = colMap.MappedColor.Code
		idx++
	}
	sort.Sort(codeArray)
	for _, code := range codeArray {
		colMap := codeMap[code]
		topOffset := CanvasMargin + (GridSize+ColorDescriptionMargin)*yOffset
		// Draw color representation symbol
		drawSymbolAtPosition(colMap, leftOffset, topOffset, dst)
		// Draw
		drawPos := freetype.Pt(leftOffset+GridSize+ColorDescriptionMargin, topOffset+GridSize-2)
		_, err := ftCtx.DrawString(colMap.MappedColor.Code+": "+colMap.MappedColor.Name, drawPos)
		if err != nil {
			fmt.Println("Rendering text failed: " + err.Error())
		}
		yOffset++
	}
}

func createCanvasImage(width int, height int, colorCount int) (result *image.RGBA) {
	canvasWidth := GridSize*width + GridBorder*(width+1)
	colorDescWidth := GridSize + ColorDescriptionLeftMargin + ColorDescriptionMargin + ColorDescriptionWidth
	pixelWidth := CanvasMargin*2 + canvasWidth + colorDescWidth
	// Height of the grid
	pixelHeight1 := CanvasMargin*2 + GridSize*height + GridBorder*(height+1)
	// Height of the legends
	pixelHeight2 := CanvasMargin*2 + (colorCount+1)*GridSize + colorCount*GridSize
	// Pick the taller height
	var pixelHeight int
	if pixelHeight1 > pixelHeight2 {
		pixelHeight = pixelHeight1
	} else {
		pixelHeight = pixelHeight2
	}
	// Create the image canvas
	result = image.NewRGBA(image.Rect(0, 0, pixelWidth, pixelHeight))
	drawGrid(width, height, result)
	return
}

func generateColorMap(img image.Image, colorList []ColorInfo) (result ColorMap) {
	result = make(ColorMap)
	idx := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			color := img.At(x, y)
			if _, _, _, alpha := color.RGBA(); alpha == 0 {
				continue
			}
			if _, exists := result[color]; !exists {
				mappedCol := CreateMappedColor(color, colorList, idx)
				for _, currentMappedCol := range result {
					if mappedCol.MappedColor.Equals(currentMappedCol.MappedColor) {
						mappedCol = currentMappedCol
						idx--
						break
					}
				}
				result[color] = mappedCol
				idx++
			}
		}
	}
	return
}

func GenerateCrossStitchPlanImage(img image.Image, colorList []ColorInfo, fnt *truetype.Font) image.Image {
	colorMap := generateColorMap(img, colorList)
	result := createCanvasImage(img.Bounds().Dx(), img.Bounds().Dy(), len(colorMap))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			color := img.At(x, y)
			if _, _, _, alpha := color.RGBA(); alpha != 0 {
				xPos := CanvasMargin + GridBorder + (GridSize+GridBorder)*x
				yPos := CanvasMargin + GridBorder + (GridSize+GridBorder)*y
				drawSymbolAtPosition(colorMap[color], xPos, yPos, result)
			}
		}
	}
	drawColorLegends(img.Bounds().Dx(), img.Bounds().Dy(), result, colorMap, fnt)
	return result
}
