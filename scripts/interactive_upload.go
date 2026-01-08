package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	baseURL = "http://localhost:8080/api/v1"
)

type ImageRecord struct {
	ImageID      string
	OriginalPath string
	Title        string
	Category     string // "original", "product-simulation-ai", "design-ai"
	UploadedAt   time.Time
}

type SearchRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Query   string         `json:"query"`
}

type SearchResult struct {
	ImageID        string  `json:"image_id"`
	RelevanceScore float64 `json:"relevance_score"`
	Reason         string  `json:"reason"`
}

var uploadedImages []ImageRecord

func main() {
	fmt.Println("🎨 Interactive Artwork Upload & Knowledge Management Agent")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()

	// Step 1: Check server
	fmt.Println("📡 Checking server connection...")
	if err := checkHealth(); err != nil {
		fmt.Printf("❌ Server not running: %v\n", err)
		fmt.Println("💡 Start the server first: make run")
		os.Exit(1)
	}
	fmt.Println("✅ Server is ready")
	fmt.Println()

	// Step 2: Get base path
	homeDir, _ := os.UserHomeDir()
	basePath := filepath.Join(homeDir, "Downloads", "artwork_images")

	fmt.Printf("📁 Base path: %s\n", basePath)
	fmt.Println()

	// Check if path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		fmt.Printf("❌ Path not found: %s\n", basePath)
		fmt.Print("Enter the correct path: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		basePath = strings.TrimSpace(input)
	}

	// Step 3: Scan all folders automatically
	fmt.Println("🔍 Scanning all folders in artwork_images directory:")

	entries, err := os.ReadDir(basePath)
	if err != nil {
		fmt.Printf("❌ Cannot read directory: %v\n", err)
		return
	}

	// Build folder list from actual directories
	folders := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			folderName := entry.Name()
			// Create category name from folder name (sanitize for storage)
			category := strings.ToLower(folderName)
			category = strings.ReplaceAll(category, " ", "-")
			folders[folderName] = category
		}
	}

	if len(folders) == 0 {
		fmt.Println("❌ No subdirectories found!")
		return
	}

	totalImages := 0
	for folderName, category := range folders {
		folderPath := filepath.Join(basePath, folderName)
		count := countImages(folderPath)
		fmt.Printf("  • %s → %d images (category: %s)\n", folderName, count, category)
		totalImages += count
	}
	fmt.Printf("\n📊 Total: %d images found across %d folders\n\n", totalImages, len(folders))

	if totalImages == 0 {
		fmt.Println("❌ No images found!")
		return
	}

	// Step 4: Confirm upload
	fmt.Print("🚀 Ready to upload. Continue? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		fmt.Println("Cancelled.")
		return
	}

	// Step 5: Upload all images
	fmt.Println("\n📤 Uploading images...")
	fmt.Println()

	for folderName, category := range folders {
		folderPath := filepath.Join(basePath, folderName)
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("📁 Processing: %s\n", folderName)
		uploadFolder(folderPath, folderName, category)
		fmt.Println()
	}

	fmt.Printf("✅ Upload complete! Uploaded %d images\n", len(uploadedImages))
	fmt.Println()

	// Step 6: Wait for processing
	fmt.Println("⏳ Waiting for AI processing...")
	fmt.Println("   (Background workers analyzing images with Gemini)")
	waitTime := time.Duration(len(uploadedImages)*10) * time.Second
	if waitTime > 120*time.Second {
		waitTime = 120 * time.Second
	}

	for i := 0; i < int(waitTime.Seconds()); i++ {
		fmt.Printf("\r   Progress: %d/%d seconds", i+1, int(waitTime.Seconds()))
		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n✅ Processing should be complete")
	fmt.Println()

	// Step 7: Enter interactive mode
	fmt.Println("🤖 Entering Interactive Knowledge Management Mode")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()
	showHelp()
	interactiveMode()
}

func countImages(folderPath string) int {
	count := 0
	files, err := os.ReadDir(folderPath)
	if err != nil {
		fmt.Printf("    ⚠️  Error reading folder: %v\n", err)
		return 0
	}
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
				count++
			}
		}
	}
	return count
}

func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

func uploadFolder(folderPath, folderName, category string) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		fmt.Printf("  ❌ Error reading folder: %v\n", err)
		return
	}

	for i, file := range files {
		if file.IsDir() || !isImage(file.Name()) {
			continue
		}

		imagePath := filepath.Join(folderPath, file.Name())
		title := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		fmt.Printf("  [%d] Uploading: %s... ", i+1, file.Name())

		imageID, err := uploadImage(imagePath, title, category, folderName)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			continue
		}

		if imageID == "" {
			fmt.Printf("❌ No image ID returned (upload may have failed)\n")
			continue
		}

		uploadedImages = append(uploadedImages, ImageRecord{
			ImageID:      imageID,
			OriginalPath: imagePath,
			Title:        title,
			Category:     category,
			UploadedAt:   time.Now(),
		})

		// Only print first 8 chars if ID is long enough
		displayID := imageID
		if len(imageID) > 8 {
			displayID = imageID[:8] + "..."
		}
		fmt.Printf("✅ %s\n", displayID)
		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

func uploadImage(imagePath, title, category, source string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	writer.WriteField("title", title)
	writer.WriteField("artist", "Colleague Artwork")

	// Add tags based on category
	tags := []string{category, source}
	if strings.Contains(category, "ai") {
		tags = append(tags, "ai-generated")
	} else {
		tags = append(tags, "original")
	}
	tagsJSON, _ := json.Marshal(tags)
	writer.WriteField("tags", string(tagsJSON))

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/images/upload", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Try "id" first, then "image_id" for compatibility
	imageID, ok := result["id"].(string)
	if !ok || imageID == "" {
		imageID, ok = result["image_id"].(string)
	}

	if !ok || imageID == "" {
		// If still no ID, check if there's an actual error
		if msg, exists := result["error"].(string); exists {
			return "", fmt.Errorf("server error: %s", msg)
		}
		return "", fmt.Errorf("no id in response: %v", result)
	}

	return imageID, nil
}

func interactiveMode() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n💬 You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// Parse command
		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		switch command {
		case "search", "s":
			query := strings.Join(parts[1:], " ")
			if query == "" {
				fmt.Println("❌ Usage: search <query>")
				continue
			}
			handleSearch(query)

		case "stats", "stat":
			handleStats()

		case "list", "ls":
			handleList()

		case "help", "h", "?":
			showHelp()

		case "quit", "exit", "q":
			fmt.Println("👋 Goodbye!")
			return

		default:
			// Treat unknown commands as search queries
			handleSearch(input)
		}
	}
}

func handleSearch(query string) {
	fmt.Printf("🔍 Searching for: \"%s\"\n", query)

	results, err := searchImages(query, 10)
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return
	}

	if results.Total == 0 {
		fmt.Println("📭 No results found")
		fmt.Println("💡 Try different keywords or wait longer for processing")
		return
	}

	fmt.Printf("✅ Found %d result(s):\n\n", results.Total)
	for i, result := range results.Results {
		// Find the record
		var record *ImageRecord
		for j := range uploadedImages {
			if uploadedImages[j].ImageID == result.ImageID {
				record = &uploadedImages[j]
				break
			}
		}

		fmt.Printf("  %d. [%s] Score: %.2f\n", i+1, result.ImageID[:8], result.RelevanceScore)
		if record != nil {
			fmt.Printf("     Title: %s\n", record.Title)
			fmt.Printf("     Category: %s\n", record.Category)
			fmt.Printf("     Source: %s\n", filepath.Base(filepath.Dir(record.OriginalPath)))
		}
		fmt.Printf("     Reason: %s\n", result.Reason)
		fmt.Println()
	}
}

func handleStats() {
	fmt.Println("📊 Statistics:")
	fmt.Printf("  Total uploaded: %d images\n", len(uploadedImages))

	// Count by category
	categories := make(map[string]int)
	for _, img := range uploadedImages {
		categories[img.Category]++
	}

	fmt.Println("\n  By category:")
	for cat, count := range categories {
		fmt.Printf("    • %s: %d images\n", cat, count)
	}

	// Time span
	if len(uploadedImages) > 0 {
		first := uploadedImages[0].UploadedAt
		last := uploadedImages[len(uploadedImages)-1].UploadedAt
		duration := last.Sub(first)
		fmt.Printf("\n  Upload duration: %s\n", duration.Round(time.Second))
	}
}

func handleList() {
	fmt.Printf("📋 Uploaded Images (%d total):\n\n", len(uploadedImages))

	for i, img := range uploadedImages {
		fmt.Printf("  %d. [%s] %s\n", i+1, img.ImageID[:8], img.Title)
		fmt.Printf("     Category: %s | Time: %s\n",
			img.Category,
			img.UploadedAt.Format("15:04:05"))
	}
}

func showHelp() {
	fmt.Println("📖 Available commands:")
	fmt.Println("  search <query>  - Search for images (or just type your query)")
	fmt.Println("  stats           - Show upload statistics")
	fmt.Println("  list            - List all uploaded images")
	fmt.Println("  help            - Show this help")
	fmt.Println("  quit            - Exit the agent")
	fmt.Println()
	fmt.Println("💡 Examples:")
	fmt.Println("  search product design")
	fmt.Println("  search AI generated images")
	fmt.Println("  original reference images")
}

func checkHealth() error {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func searchImages(query string, limit int) (*SearchResponse, error) {
	searchReq := SearchRequest{
		Query: query,
		Limit: limit,
	}

	jsonData, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}
