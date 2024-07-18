package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	MathRand "math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	serverAddr    = "0.0.0.0"
	disconnectMsg = "!DISCONNECT"
)

var (
	clients     = make(map[net.Conn]string) // Store session key per client
	usernames   = make(map[net.Conn]string) // Store username per client
	sessionHash = ""
)

const (
	baseURL   = "https://api.open-meteo.com/v1/forecast"
	latitude  = 30.0444 // Cairo's latitude
	longitude = 31.2357 // Cairo's longitude
)

// WeatherResponse represents the structure of the weather API response
type WeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		WeatherCode int     `json:"weathercode"`
	} `json:"current_weather"`
}

var shuffelNumber int64

func main() {
	ImagesHashLoaded()
	MathRand.Seed(time.Now().UnixNano())

	// Generate the combined number
	combinedNumber := GenerateCombinedNumber()

	fmt.Println("Unique 50-digit number:", combinedNumber)

	intValue, err := strconv.ParseInt(combinedNumber, 10, 64)
	if err != nil {
		fmt.Println("Error converting to int64:", err)
		return
	}

	shuffelNumber = intValue

	cert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		fmt.Println("Error loading server certificate:", err)
		return
	}

	caCert, err := ioutil.ReadFile("cert/root-cert.pem")
	if err != nil {
		fmt.Println("Error loading CA certificate:", err)
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	listener, err := tls.Listen("tcp", serverAddr+":443", config)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server started on", serverAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Prompt for username first
	payload := receiveMessage(conn, "")
	if payload == "" {
		return
	}
	payloadSplit := strings.Split(payload, "|")
	username := payloadSplit[0]
	userHash := payloadSplit[1]

	for _, s := range usernames {
		if s == username {
			fmt.Println("This Username Already online:", userHash, " | ", username)
			return
		}
	}

	if userHash != sessionHash {
		fmt.Println("This Username Already online:", userHash, " | ", username)
		return
	}

	usernames[conn] = username

	sessionKey := generateSessionKey()
	fmt.Println("Session Key for", username, ":", sessionKey)
	// Provide this session key to the client manually

	// Validate session key from the client
	if valid, err := validateSessionKey(conn, sessionKey); !valid || err != nil {
		fmt.Println("Invalid session key from client:", err)
		return
	}

	clients[conn] = sessionKey

	// Announce new user joining
	joinMessage := ShuffleString(" has joined the chat .", shuffelNumber)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	joinMessage = fmt.Sprintf("( %s - %s ) - %s", timestamp, username, joinMessage)

	broadcast(joinMessage, conn)

	for {
		msg := receiveMessage(conn, sessionKey)
		if msg == "" {
			delete(clients, conn)
			delete(usernames, conn)
			return
		}
		if msg == disconnectMsg {
			leftMessage := ShuffleString(" has left the chat .", shuffelNumber)
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			leftMessage = fmt.Sprintf("( %s - %s ) - %s", timestamp, username, leftMessage)

			broadcast(leftMessage, conn)
			delete(clients, conn)
			delete(usernames, conn)
			return
		}
		broadcast(msg, conn)
	}
}

func generateSessionKey() string {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println("Error generating session key:", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(key)
}

func validateSessionKey(conn net.Conn, sessionKey string) (bool, error) {
	key, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	key = strings.TrimSpace(key)
	return key == sessionKey, nil
}

func receiveMessage(conn net.Conn, sessionKey string) string {
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading message:", err, msg)
		}
		return ""
	}
	msg = strings.TrimSpace(msg)

	if sessionKey != "" {
		return decryptMessage(msg, sessionKey)
	}
	return msg
}

func broadcast(message string, excludeConn net.Conn) {
	for client := range clients {
		if client != excludeConn {
			sendMessage(client, message, clients[client])
		}
	}
}

func sendMessage(conn net.Conn, message string, sessionKey string) {
	encryptedMessage := encryptMessage(message, sessionKey)
	conn.Write([]byte(encryptedMessage + "\n"))
}

func encryptMessage(message string, sessionKey string) string {
	keyBytes, _ := base64.StdEncoding.DecodeString(sessionKey)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Error creating GCM:", err)
		return ""
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		fmt.Println("Error creating nonce:", err)
		return ""
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(message), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func decryptMessage(encryptedMessage string, sessionKey string) string {
	keyBytes, _ := base64.StdEncoding.DecodeString(sessionKey)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Error creating GCM:", err)
		return ""
	}
	decodedMessage, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		fmt.Println("Error decoding message:", err)
		return ""
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := decodedMessage[:nonceSize], decodedMessage[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Println("Error decrypting message:", err)
		return ""
	}
	return string(plaintext)
}

func ImagesHashLoaded() {
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

	// Combine the data to generate a hash
	hashInput := fmt.Sprintf("FileName: %s, PixelCount: %d", *fileName, numPixels)
	for _, pixel := range pixels {
		hashInput += fmt.Sprintf("(%d,%d):%s,", pixel.X, pixel.Y, pixel.Hex)
	}

	// Generate the SHA-256 hash
	hash := sha256.Sum256([]byte(hashInput))
	hashString := fmt.Sprintf("%x", hash)

	sessionHash = hashString
	// Print the hash
	fmt.Printf("Generated Hash: %s\n", hashString)

}

// Helper function to convert color to hex
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// GetCurrentWeatherInCairo fetches the current weather in Cairo using Open-Meteo API
func GetCurrentWeatherInCairo() string {
	url := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&current_weather=true", baseURL, latitude, longitude)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching weather data:", err)
		return "Unable to fetch weather data"
	}
	defer resp.Body.Close()

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		fmt.Println("Error decoding weather data:", err)
		return "Unable to decode weather data"
	}

	temperature := weatherResponse.CurrentWeather.Temperature
	weatherCode := weatherResponse.CurrentWeather.WeatherCode

	// Convert weather code to a human-readable format
	var condition string
	switch weatherCode {
	case 0:
		condition = "Clear sky"
	case 1, 2, 3:
		condition = "Partly cloudy"
	case 45, 48:
		condition = "Foggy"
	case 51, 53, 55:
		condition = "Drizzle"
	case 61, 63, 65:
		condition = "Rain"
	case 71, 73, 75:
		condition = "Snow"
	case 80, 81, 82:
		condition = "Rain showers"
	case 95:
		condition = "Thunderstorm"
	default:
		condition = "Unknown"
	}

	return fmt.Sprintf("%s, %.1f°C", condition, temperature)
}

// RandomLocation generates a random latitude and longitude
func RandomLocation() (float64, float64) {
	latitude := MathRand.Float64()*180 - 90   // Latitude: -90 to +90
	longitude := MathRand.Float64()*360 - 180 // Longitude: -180 to +180
	return latitude, longitude
}

// HashToNumericString converts a hash to a numeric string
func HashToNumericString(hash [32]byte) string {
	var num uint64
	var numericString string

	for i := 0; i < len(hash); i += 8 {
		num = binary.BigEndian.Uint64(hash[i : i+8])
		numericString += fmt.Sprintf("%020d", num)
	}

	return numericString
}

// GenerateCombinedNumber generates a unique 50-digit number
func GenerateCombinedNumber() string {
	for {
		// Generate the current timestamp
		timestamp := time.Now().Format(time.RFC3339)

		// Get the current weather in Cairo
		weather := GetCurrentWeatherInCairo()

		// Generate a random location
		latitude, longitude := RandomLocation()

		// Combine all information
		combined := fmt.Sprintf("%s|%s|%.6f|%.6f", timestamp, weather, latitude, longitude)

		// Generate a SHA-256 hash of the combined string
		hash := sha256.Sum256([]byte(combined))

		// Convert the hash to a numeric string
		numericString := HashToNumericString(hash)

		// Ensure the string is 50 digits long and does not start with a zero
		if numericString[0] != '0' {
			return numericString[:15]
		}
	}
}

func ShuffleString(s string, seed int64) string {
	// Seed the random number generator with the seed
	MathRand.Seed(seed)

	// Convert the string to a slice of runes
	runes := []rune(s)

	// Shuffle the slice of runes
	MathRand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})

	// Convert the slice of runes back to a string
	return string(runes)
}
