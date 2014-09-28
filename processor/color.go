package processor

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"math"
)

type ColorInfo struct {
	Code string `json="code"`
	Name string `json="name"`
	R    uint8  `json="r"`
	G    uint8  `json="g"`
	B    uint8  `json="b"`
}

func (c ColorInfo) Equals(col ColorInfo) bool {
	return c.R == col.R && c.G == col.G && c.B == col.B
}

func LoadColorInfo(path string) (result []ColorInfo) {
	jsonBytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Missing color mapping JSON")
		return
	}
	json.Unmarshal(jsonBytes, &result)
	return
}

type ColorMapInfo struct {
	MappedColor ColorInfo
	Pattern     int
}

func CreateMappedColor(col color.Color, colorList []ColorInfo, pattern int) (result ColorMapInfo) {
	result.Pattern = pattern
	var minColorDistance float64 = math.MaxFloat64
	for _, colInfo := range colorList {
		currentCol := color.RGBA{colInfo.R, colInfo.G, colInfo.B, 255}
		currentColDist := GetColorDistance(col, currentCol)
		if currentColDist < minColorDistance {
			result.MappedColor = colInfo
			minColorDistance = currentColDist
		}
	}
	return
}

func GetColorDistance(c1 color.Color, c2 color.Color) float64 {
	c1r, c1g, c1b, _ := c1.RGBA()
	c2r, c2g, c2b, _ := c2.RGBA()
	var rmean, r, g, b int
	rmean = int((c1r>>8 + c2r>>8) / 2)
	r = int(c1r>>8) - int(c2r>>8)
	g = int(c1g>>8) - int(c2g>>8)
	b = int(c1b>>8) - int(c2b>>8)
	return math.Sqrt(float64((((512 + rmean) * r * r) >> 8) + 4*g*g + (((767 - rmean) * b * b) >> 8)))
}
