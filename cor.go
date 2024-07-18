package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Open the image file
	fileName := "images/secret.jpeg" // Replace this with your actual image file path
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error: Unable to open image file: %s\n", err)
		return
	}
	defer file.Close()

	// Decode the image
	var img image.Image
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		fmt.Println("Error: Unsupported image format")
		return
	}

	if err != nil {
		fmt.Printf("Error: Unable to decode image: %s\n", err)
		return
	}

	// Get the image bounds
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Generate 10 random coordinates within the image bounds
	coordinates := make([]struct {
		X int
		Y int
	}, 10)
	for i := 0; i < 10; i++ {
		coordinates[i].X = rand.Intn(width)
		coordinates[i].Y = rand.Intn(height)
	}

	// Check if the coordinates exist within the image bounds
	validCoordinates := []struct {
		X int
		Y int
	}{}
	for _, coord := range coordinates {
		if coord.X >= bounds.Min.X && coord.X < bounds.Max.X && coord.Y >= bounds.Min.Y && coord.Y < bounds.Max.Y {
			validCoordinates = append(validCoordinates, coord)
		} else {
			fmt.Printf("Coordinates (%d, %d) are out of bounds.\n", coord.X, coord.Y)
		}
	}

	// Format the coordinates as "x,y;x,y;..."
	var formattedCoords strings.Builder
	for i, coord := range validCoordinates {
		if i > 0 {
			formattedCoords.WriteString(";")
		}
		formattedCoords.WriteString(fmt.Sprintf("%d,%d", coord.X, coord.Y))
	}

	// Print the formatted string
	fmt.Println(formattedCoords.String())

	// Optionally, you can use the formatted string for further processing
}
