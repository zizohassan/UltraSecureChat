package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"
	"strings"
)

// Helper function to convert color to hex
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func main() {
	// Define and parse command-line flags
	fileName := flag.String("image", "path_to_your_image.jpg", "Path to the image file")
	coords := flag.String("coords", "10,10;20,20;30,30", "Coordinates in x,y format separated by semicolons")
	flag.Parse()

	// Open the image file
	file, err := os.Open(*fileName)
	if err != nil {
		fmt.Println("Error: Unable to open image file")
		return
	}
	defer file.Close()

	// Decode the image
	var img image.Image
	switch {
	case strings.HasSuffix(*fileName, ".jpg"), strings.HasSuffix(*fileName, ".jpeg"):
		img, err = jpeg.Decode(file)
	case strings.HasSuffix(*fileName, ".png"):
		img, err = png.Decode(file)
	default:
		fmt.Println("Error: Unsupported image format")
		return
	}

	if err != nil {
		fmt.Println("Error: Unable to decode image")
		return
	}

	// Get the image bounds
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Calculate the number of pixels
	numPixels := width * height
	//fmt.Printf("The image has %d pixels.\n", numPixels)

	// Parse the coordinates from the command-line argument
	coordStrings := strings.Split(*coords, ";")
	var coordinates []struct {
		X int
		Y int
	}

	for _, coordStr := range coordStrings {
		xy := strings.Split(coordStr, ",")
		if len(xy) != 2 {
			fmt.Println("Error: Invalid coordinate format")
			return
		}
		x, err1 := strconv.Atoi(xy[0])
		y, err2 := strconv.Atoi(xy[1])
		if err1 != nil || err2 != nil {
			fmt.Println("Error: Invalid coordinate values")
			return
		}
		coordinates = append(coordinates, struct {
			X int
			Y int
		}{X: x, Y: y})
	}

	// Create a slice to store the results
	type Pixel struct {
		X   int
		Y   int
		Hex string
	}
	var pixels []Pixel

	// Get the hex color of each pixel at the specified coordinates
	for _, coord := range coordinates {
		if coord.X >= img.Bounds().Min.X && coord.X < img.Bounds().Max.X && coord.Y >= img.Bounds().Min.Y && coord.Y < img.Bounds().Max.Y {
			color := img.At(coord.X, coord.Y)
			hex := colorToHex(color)
			pixels = append(pixels, Pixel{X: coord.X, Y: coord.Y, Hex: hex})
		} else {
			fmt.Printf("Coordinates (%d, %d) are out of bounds.\n", coord.X, coord.Y)
		}
	}

	// Print the pixel data
	//for _, pixel := range pixels {
	//	fmt.Printf("Pixel at (%d, %d): %s\n", pixel.X, pixel.Y, pixel.Hex)
	//}

	// Combine the data to generate a hash
	hashInput := fmt.Sprintf("FileName: %s, PixelCount: %d", *fileName, numPixels)
	for _, pixel := range pixels {
		hashInput += fmt.Sprintf("(%d,%d):%s,", pixel.X, pixel.Y, pixel.Hex)
	}

	// Generate the SHA-256 hash
	hash := sha256.Sum256([]byte(hashInput))
	hashString := fmt.Sprintf("%x", hash)

	// Print the hash
	fmt.Printf("Generated Hash: %s\n", hashString)

	// Optionally, you can do something else with the hash, like saving it to a file or further processing.
}
