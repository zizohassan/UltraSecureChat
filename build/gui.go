package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	FarsiReshaper "github.com/javad-majidi/farsi-reshaper"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	MathRand "math/rand"
)

var conn net.Conn

var keyBytes []byte

var shufflenumber int64 = 10

var clearChat = "{CLEAR_CHAT}"

var userName string

func loadCustomFont() fyne.Resource {
	// Load custom font file
	fontPath := "./NotoSansArabic_Condensed-Regular.ttf" // Ensure this path is correct
	fontFile, err := os.ReadFile(fontPath)
	if err != nil {
		log.Fatalf("Unable to read font file: %v", err)
	}
	return fyne.NewStaticResource("NotoSansArabic_Condensed-Regular.ttf", fontFile)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Chat Client")

	// Set application icon
	resource, err := fyne.LoadResourceFromPath("icon.png")
	if err != nil {
		log.Fatalf("Failed to load icon: %v", err)
	}
	myApp.SetIcon(resource)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	usernameEntry.SetText("zizo")
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("IP Address")
	ipEntry.SetText("192.168.8.102")
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Port")
	portEntry.SetText("443")
	userSecret := widget.NewEntry()
	userSecret.SetPlaceHolder("User Secret")
	userSecret.SetText("348293b79f19bb6153e85bd26f3f8f1d97d2a0c7fee475b20a8050ab0fe00709")

	connectButton := widget.NewButton("Connect", func() {
		username := usernameEntry.Text
		ip := ipEntry.Text
		port := portEntry.Text
		secretKey := userSecret.Text
		userName = username

		if username == "" || ip == "" || port == "" || secretKey == "" {
			dialog.ShowError(fmt.Errorf("All fields are required"), myWindow)
			return
		}

		go connectToServer(myApp, username, ip, port, myWindow, secretKey)
	})

	myWindow.SetContent(container.NewVBox(
		usernameEntry,
		ipEntry,
		portEntry,
		userSecret,
		connectButton,
	))

	myWindow.Resize(fyne.NewSize(600, 190))
	myWindow.ShowAndRun()
}

func connectToServer(app fyne.App, username, ip, port string, mainWindow fyne.Window, userSecret string) {
	var err error

	cert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile("root-cert.pem")
	if err != nil {
		log.Fatal(err)
	}

	// Create a CA certificate pool and add the server's CA certificate to it
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Set up the TLS configuration with the client's certificate and the CA pool
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	conn, err := tls.Dial("tcp", ip+":"+port, config)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		dialog.ShowError(fmt.Errorf("Failed to connect to server: %v", err), mainWindow)
		return
	}

	payload := username + "|" + userSecret + "\n"
	_, err = conn.Write([]byte(payload))
	if err != nil {
		log.Printf("Failed to send username: %v", err)
		dialog.ShowError(fmt.Errorf("Failed to send username: %v", err), mainWindow)
		return
	}

	showKeyInputScreen(app, username, mainWindow, conn)
}

func showKeyInputScreen(app fyne.App, username string, mainWindow fyne.Window, conn net.Conn) {
	keyWindow := app.NewWindow("Enter Key")
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Key")

	connectButton := widget.NewButton("Connect", func() {
		var errKey error
		key := keyEntry.Text
		if key == "" {
			dialog.ShowError(fmt.Errorf("Key is required"), keyWindow)
			return
		}

		keyBytes = []byte(key)
		if len(keyBytes) != 32 {
			// Pad or truncate the key to 32 bytes
			keyBytes = padOrTruncateKey(keyBytes, 32)
		}

		keyBytes, errKey = base64.StdEncoding.DecodeString(key)
		if errKey != nil {
			fmt.Println("Invalid session key:", errKey)
			os.Exit(1)
		}

		_, errConn := conn.Write([]byte(key + "\n"))
		if errConn != nil {
			fmt.Println("Error sending session key to server:", errConn.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully connected to the server!")

		showChatScreen(app, username, keyWindow, conn)
	})

	keyWindow.SetContent(container.NewVBox(
		keyEntry,
		connectButton,
	))

	keyWindow.Resize(fyne.NewSize(600, 80))
	keyWindow.Show()
	mainWindow.Close()
}

func showChatScreen(app fyne.App, username string, keyWindow fyne.Window, conn net.Conn) {
	chatWindow := app.NewWindow("Chat")
	chatWindow.Resize(fyne.NewSize(800, 500))

	customFont := loadCustomFont()
	fyne.CurrentApp().Settings().SetTheme(&customTheme{font: customFont})

	chatContent := container.New(layout.NewVBoxLayout())
	scrollContainer := container.NewScroll(chatContent)

	// Create a data binding
	textBinding := binding.NewString()

	// Create an entry widget and bind it to the data binding
	shuffelText := widget.NewEntryWithData(textBinding)
	shuffelText.SetPlaceHolder("decrypt text ....")

	// Function to add custom logic on value change
	textBinding.AddListener(binding.NewDataListener(func() {
		value, err := textBinding.Get()
		if err == nil {
			if value == "" {
				shufflenumber = 10
				return
			}
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Println("Error converting to int64:", err)
				return
			}
			fmt.Println("*********************", shuffelText.Text, shufflenumber)
			// Add your custom logic here
			shufflenumber = intValue
			fmt.Println("*********************", shuffelText.Text, shufflenumber)

		}
	}))

	chatEntry := widget.NewMultiLineEntry()
	chatEntry.SetPlaceHolder("Type your message here...")

	sendButton := widget.NewButton("Send", func() {
		message := chatEntry.Text
		if message == "" {
			return
		}
		msgsrc := message
		if shuffelText.Text == "" || shufflenumber == 10 {
			dialog.ShowError(fmt.Errorf("decrypt text must have value"), chatWindow)
			return
		}
		message = ShuffleString(message, shufflenumber)
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		displayMessage := fmt.Sprintf("( %s - %s ) - %s", timestamp, userName, message)

		encryptedMessage := encryptMessage(displayMessage)
		_, err := conn.Write([]byte(encryptedMessage + "\n"))
		if err != nil {
			log.Printf("Failed to send message: %v", err)
			dialog.ShowError(fmt.Errorf("Failed to send message: %v", err), chatWindow)
			return
		}
		chatEntry.SetText("")
		displayMessage = strings.Replace(displayMessage, message, msgsrc, -1)
		displayMessage = handleRTL(displayMessage) // Handle RTL if necessary
		displayMessage = strings.TrimSpace(displayMessage)
		isRTL := containsRTL(displayMessage)
		var messageBox *fyne.Container
		if isRTL {
			messageBox = container.NewVBox(
				widget.NewLabelWithStyle(displayMessage, fyne.TextAlignTrailing, fyne.TextStyle{}),
			)
		} else {
			messageBox = container.NewVBox(
				widget.NewLabelWithStyle(displayMessage, fyne.TextAlignLeading, fyne.TextStyle{}),
			)
		}
		chatContent.Add(messageBox)
		scrollContainer.ScrollToBottom()
		chatWindow.Content().Refresh()
	})
	clearChatButton := widget.NewButton("Urgent Clean", func() {
		if chatContent.Objects == nil {
			return
		}
		clearChatShufful := ShuffleString(clearChat, shufflenumber)
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		displayMessage := fmt.Sprintf("( %s - %s ) - %s", timestamp, userName, clearChatShufful)

		encryptedMessage := encryptMessage(displayMessage)
		_, err := conn.Write([]byte(encryptedMessage + "\n"))
		if err != nil {
			log.Printf("Failed to send message: %v", err)
			dialog.ShowError(fmt.Errorf("Failed to send message: %v", err), chatWindow)
			return
		}
		chatContent.Objects = nil // Clear chat content in the client UI
		chatContent.Refresh()
	})
	clearChatButton.Importance = widget.DangerImportance

	buttonContainer := container.NewGridWithColumns(2,
		container.NewHScroll(sendButton),
		clearChatButton,
	)

	inputContainer := container.NewVBox(
		shuffelText,
		chatEntry,
		buttonContainer,
	)
	scrollContainer.SetMinSize(fyne.NewSize(800, 400)) // Set fixed height for scroll container

	chatWindow.SetContent(container.NewBorder(
		scrollContainer,
		inputContainer,
		nil, nil,
	))

	go receiveMessages(chatContent, chatWindow, conn, scrollContainer)
	chatWindow.Show()
	keyWindow.Close()
}

func receiveMessages(chatContent *fyne.Container, chatWindow fyne.Window, conn net.Conn, scrollContainer *container.Scroll) {
	if conn == nil {
		log.Println("Connection is nil")
		dialog.ShowError(fmt.Errorf("Connection is not established"), chatWindow)
		return
	}
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			if err == io.EOF {
				log.Println("Connection closed by server")
				dialog.ShowInformation("Disconnected", "Connection closed by server", chatWindow)
				return
			}
			dialog.ShowError(fmt.Errorf("Failed to receive message: %v", err), chatWindow)
			return
		}
		message := decryptMessage(strings.TrimSpace(msg))
		if message == "" {
			continue
		}
		splitMessage := strings.Split(message, " ) - ")
		displayMessage := UnshuffleString(splitMessage[1], shufflenumber)
		displayMessage = strings.Replace(message, splitMessage[1], displayMessage, -1)
		displayMessage = handleRTL(displayMessage) // Handle RTL if necessary
		displayMessage = strings.TrimSpace(displayMessage)
		isRTL := containsRTL(displayMessage)
		var messageBox *fyne.Container
		if isRTL {
			messageBox = container.NewVBox(
				widget.NewLabelWithStyle(displayMessage, fyne.TextAlignTrailing, fyne.TextStyle{}),
			)
		} else {
			messageBox = container.NewVBox(
				widget.NewLabelWithStyle(displayMessage, fyne.TextAlignLeading, fyne.TextStyle{}),
			)
		}
		if strings.Contains(displayMessage, clearChat) {
			chatContent.Objects = nil // Clear chat content in the client UI
			chatContent.Refresh()
		} else {
			chatContent.Add(messageBox)
			scrollContainer.ScrollToBottom()
			chatWindow.Content().Refresh()
		}

	}
}

func padOrTruncateKey(key []byte, length int) []byte {
	if len(key) > length {
		return key[:length]
	} else if len(key) < length {
		paddedKey := make([]byte, length)
		copy(paddedKey, key)
		return paddedKey
	}
	return key
}

func encryptMessage(message string) string {
	if len(keyBytes) == 0 {
		log.Println("Encryption key is not set")
		return ""
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		log.Printf("Error creating cipher: %v", err)
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("Error creating GCM: %v", err)
		return ""
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		log.Printf("Error creating nonce: %v", err)
		return ""
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(message), nil)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func decryptMessage(encryptedMessage string) string {
	if len(keyBytes) == 0 {
		log.Println("Decryption key is not set")
		return "Invalid session key"
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		log.Printf("Error creating cipher: %v", err)
		return "Error creating cipher"
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("Error creating GCM: %v", err)
		return "Error creating GCM"
	}
	decodedMessage, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		log.Printf("Error decoding message: %v", err)
		return "Error decoding message"
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := decodedMessage[:nonceSize], decodedMessage[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("Error decrypting message: %v", err)
		return "Error decrypting message"
	}

	return string(plaintext)
}

func containsRTL(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Arabic, r) || unicode.Is(unicode.Hebrew, r) {
			return true
		}
	}
	return false
}

func handleRTL(text string) string {
	splitMessage := strings.Split(text, " ) - ")
	fmt.Println(splitMessage, text)
	if containsRTL(splitMessage[1]) {
		msg := shapeArabicText(splitMessage[1])
		rigtoLeftSplit := strings.Split(splitMessage[0], " - ")
		return msg + " - ( " + rigtoLeftSplit[1] + " - " + strings.Replace(rigtoLeftSplit[0], "(", "", -1) + " )"
	}
	return text
}

func shapeArabicText(text string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isNonSpacingMark), norm.NFC)
	result, _, _ := transform.String(t, text)
	result = FarsiReshaper.Reshape(result)

	return result
}

func isNonSpacingMark(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

type customTheme struct {
	font fyne.Resource
}

func (c *customTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTextMonospaceFont()
	}
	if s.Bold {
		if s.Italic {
			return theme.DefaultTextBoldItalicFont()
		}
		return theme.DefaultTextBoldFont()
	}
	if s.Italic {
		return theme.DefaultTextItalicFont()
	}
	return c.font
}

func (c *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (c *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *customTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type customLayout struct {
	margin float32
}

func (c *customLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := c.margin
	for _, obj := range objects {
		obj.Resize(fyne.NewSize(size.Width-2*c.margin, obj.MinSize().Height))
		obj.Move(fyne.NewPos(c.margin, y))
		y += obj.MinSize().Height
	}
}

func (c *customLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := c.margin
	for _, obj := range objects {
		childSize := obj.MinSize()
		if childSize.Width > width {
			width = childSize.Width
		}
		height += childSize.Height + c.margin
	}
	return fyne.NewSize(width+2*c.margin, height)
}

func newCustomLayout(margin float32) fyne.Layout {
	return &customLayout{margin: 0}
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

// GenerateOriginalIndices generates the shuffle order based on the seed
func GenerateOriginalIndices(length int, seed int64) []int {
	MathRand.Seed(seed)
	indices := make([]int, length)
	for i := range indices {
		indices[i] = i
	}
	MathRand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices
}

// UnshuffleString reverses the shuffle based on the seed
func UnshuffleString(s string, seed int64) string {
	// Convert the string to a slice of runes
	runes := []rune(s)
	length := len(runes)

	// Get the shuffle order
	indices := GenerateOriginalIndices(length, seed)

	// Create a slice to hold the unshuffled runes
	unshuffledRunes := make([]rune, length)

	// Place each rune back in its original position
	for i, idx := range indices {
		unshuffledRunes[idx] = runes[i]
	}

	// Convert the slice of runes back to a string
	return string(unshuffledRunes)
}
