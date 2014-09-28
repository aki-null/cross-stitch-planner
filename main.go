package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"net/http"

	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/aki-null/cross-stitch-planner/processor"
)

type processAPIResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
	Body    string `json:"body"`
}

func (p processAPIResponse) GetResponse() (result []byte) {
	result, err := json.Marshal(p)
	if err != nil {
		result = []byte("")
	}
	return
}

func main() {
	// Load available thread colors
	dmcColor := processor.LoadColorInfo("dmc.json")
	// Load font for legends rendering
	fontBytes, err := ioutil.ReadFile("DroidSans.ttf")
	if err != nil {
		fmt.Println("Missing font: " + err.Error())
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		fmt.Println("Invalid font: " + err.Error())
	}

	// Register static site
	http.Handle("/", http.FileServer(http.Dir("./static/")))
	// Register API endpoint
	http.HandleFunc("/api/process", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > 1024*256+1024 {
			w.Write(processAPIResponse{false, "The file is too large. It must be smaller than 256KB.", ""}.GetResponse())
			return
		}
		// Only allow maximum of 1MB of memory to be used
		r.ParseMultipartForm(1024 * 1024)
		if r.MultipartForm == nil {
			w.Write(processAPIResponse{false, "No image provided", ""}.GetResponse())
			return
		}
		// Get image file in request
		files := r.MultipartForm.File["input"]
		if len(files) == 0 {
			w.Write(processAPIResponse{false, "No image provided", ""}.GetResponse())
			return
		}
		imageFileHeader := files[0]
		imageFile, err := imageFileHeader.Open()
		if err != nil {
			w.Write(processAPIResponse{false, "Error opening image file", ""}.GetResponse())
			return
		}
		// Decode the PNG image
		parsedImage, err := png.Decode(imageFile)
		if err != nil {
			w.Write(processAPIResponse{false, "Error decoding image file. The file must be PNG.", ""}.GetResponse())
			return
		}
		if parsedImage.Bounds().Dx() > 128 || parsedImage.Bounds().Dy() > 128 {
			w.Write(processAPIResponse{false, "The image is too big. The maximum dimension is 128x128.", ""}.GetResponse())
			return
		}
		// Generate cross-stich plan image
		resultImage := processor.GenerateCrossStitchPlanImage(parsedImage, dmcColor, font)
		imageBuf := new(bytes.Buffer)
		// Encode it into PNG
		err = png.Encode(imageBuf, resultImage)
		if err != nil {
			w.Write(processAPIResponse{false, "Error encoding output PNG", ""}.GetResponse())
			return
		}
		// Encode the image into base64 string
		encodedImageBuf := new(bytes.Buffer)
		encoder := base64.NewEncoder(base64.StdEncoding, encodedImageBuf)
		encoder.Write(imageBuf.Bytes())
		w.Write(processAPIResponse{true, "Success", encodedImageBuf.String()}.GetResponse())
		return
	})
	fmt.Println("Ready to handle requests")
	http.ListenAndServe(":3434", nil)
}
