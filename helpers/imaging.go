package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/image/draw"

	//"math/rand"

	"fyne.io/fyne/v2"
)

func DefaultImg() *image.RGBA {
	width := 200
	height := 50

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img1 := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img1.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img1.Set(x, y, cyan)
			default:
				// Use zero value.
			}
		}
	}
	//fmt.Println("Generated image type:", reflect.TypeOf(img1))

	return img1

}

func ResizeImg(r fyne.URIReadCloser) *image.RGBA { //
	//fmt.Println("Selected File : ", string()
	//fmt.Println("Resizing.. type:", reflect.TypeOf(r))
	//fmt.Println("URI:", URI)

	input, _ := os.Open(r.URI().Path())
	defer input.Close()
	var src image.Image
	if r.URI().MimeType() == "image/png" {
		// Decode the image (from PNG to image.Image):
		src, _ = png.Decode(input)
	} else if r.URI().MimeType() == "image/jpeg" {
		src, _ = jpeg.Decode(input)
	} else {
		fmt.Println("Resized Image type:", r.URI().MimeType())
		return DefaultImg()
	}
	W := src.Bounds().Max.X
	H := src.Bounds().Max.Y
	//	fmt.Println("Width X:", W)
	//	fmt.Println("Height Y:", H)
	to_width := 500
	to_height := 500
	var ratio float64
	if W > H {
		ratio = float64(500) / float64(W)
		to_height = int(float64(H) * float64(ratio))

	} else {
		ratio = float64(500) / float64(H)
		to_width = int(float64(W) * float64(ratio))
	}

	//	fmt.Println("to_width Width X:", to_width)
	//	fmt.Println("to_heightHeight Y:", to_height)

	// Set the expected size that you want:
	dst := image.NewRGBA(image.Rect(0, 0, to_width, to_height))

	// Resize:NearestNeighbor
	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	// Encode to `output`:
	//png.Encode(output, dst)
	return dst

}

func GetImgBase64(img *image.RGBA) string {

	base64Encoding := "data:image/png;base64,"
	/*switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	}
	*/
	var buf bytes.Buffer
	imgBase64Str := ""
	if err := png.Encode(&buf, img); err != nil {
		fmt.Errorf("unable to encode png: %w", err)
	}
	data := buf.Bytes()
	// Append the base64 encoded output
	imgBase64Str = base64.StdEncoding.EncodeToString(data)

	return base64Encoding + imgBase64Str
}

func GetImgFromBase64(img_str string) *image.RGBA {
	//fmt.Println("IMGSTR:", img_str)

	before, after, found := strings.Cut(img_str, ",")
	if before != "data:image/png;base64" || !found {
		return DefaultImg()
	}

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(after))

	src, _ := png.Decode(reader)
	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X, src.Bounds().Max.Y))
	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	//fmt.Println("MAX X:", src.Bounds().Max.X)
	//fmt.Println("MAX Y:", src.Bounds().Max.Y)
	return dst
}
