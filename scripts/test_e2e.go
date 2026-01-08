package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	baseURL     = "http://localhost:8080/api/v1"
	testDataDir = "test_images"
)

type UploadResponse struct {
	ImageID string `json:"image_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
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

func main() {
	fmt.Println("🚀 Starting End-to-End Test Suite")
	fmt.Println("=" + repeat("=", 50))

	// Step 1: Check server health
	fmt.Println("\n📡 Step 1: Checking server health...")
	if err := checkHealth(); err != nil {
		fmt.Printf("❌ Server health check failed: %v\n", err)
		fmt.Println("\n💡 Make sure the server is running: make run")
		os.Exit(1)
	}
	fmt.Println("✅ Server is healthy")

	// Step 2: Prepare test images
	fmt.Println("\n🖼️  Step 2: Preparing test images...")
	testImages := prepareTestImages()
	fmt.Printf("✅ Prepared %d test images\n", len(testImages))

	// Step 3: Upload images
	fmt.Println("\n📤 Step 3: Uploading test images...")
	uploadedIDs := make([]string, 0)
	for i, img := range testImages {
		fmt.Printf("  [%d/%d] Uploading: %s...", i+1, len(testImages), img.Title)
		imageID, err := uploadImage(img)
		if err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			continue
		}
		uploadedIDs = append(uploadedIDs, imageID)
		fmt.Printf(" ✅ ID: %s\n", imageID[:8])
	}
	fmt.Printf("✅ Uploaded %d/%d images successfully\n", len(uploadedIDs), len(testImages))

	// Step 4: Wait for processing
	fmt.Println("\n⏳ Step 4: Waiting for AI processing...")
	fmt.Println("  (Background workers need time to analyze images)")
	waitTime := 30 * time.Second
	for i := 0; i < int(waitTime.Seconds()); i++ {
		fmt.Printf("\r  Waiting: %d/%d seconds...", i+1, int(waitTime.Seconds()))
		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n✅ Processing complete (hopefully!)")

	// Step 5: Test search queries
	fmt.Println("\n🔍 Step 5: Testing search functionality...")
	testQueries := []string{
		"beach sunset ocean",
		"mountain landscape",
		"city night lights",
		"forest nature",
		"animals cats",
	}

	for i, query := range testQueries {
		fmt.Printf("\n  [%d/%d] Query: \"%s\"\n", i+1, len(testQueries), query)
		results, err := searchImages(query, 5)
		if err != nil {
			fmt.Printf("    ❌ Search failed: %v\n", err)
			continue
		}

		if results.Total == 0 {
			fmt.Printf("    ℹ️  No results found\n")
		} else {
			fmt.Printf("    ✅ Found %d result(s):\n", results.Total)
			for j, result := range results.Results {
				fmt.Printf("      %d. ID: %s (Score: %.2f)\n", j+1, result.ImageID[:8], result.RelevanceScore)
				fmt.Printf("         Reason: %s\n", result.Reason)
			}
		}
	}

	// Step 6: Summary
	fmt.Println("\n" + repeat("=", 50))
	fmt.Println("📊 Test Summary:")
	fmt.Printf("  • Images uploaded: %d/%d\n", len(uploadedIDs), len(testImages))
	fmt.Printf("  • Search queries tested: %d\n", len(testQueries))
	fmt.Println("\n✅ End-to-End Test Complete!")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  • Check data/index.md to see indexed images")
	fmt.Println("  • Check data/categories/ to see organized files")
	fmt.Println("  • Try more search queries via curl or Postman")
}

type TestImage struct {
	FilePath string
	Title    string
	Artist   string
	Tags     []string
}

func prepareTestImages() []TestImage {
	// Create test data directory if it doesn't exist
	os.MkdirAll(testDataDir, 0755)

	// For this demo, we'll create placeholder instructions
	// In a real scenario, you'd download or use actual images
	images := []TestImage{
		{
			FilePath: "Please place test images in test_images/ folder",
			Title:    "Beach Sunset",
			Artist:   "Test Artist",
			Tags:     []string{"beach", "sunset", "ocean", "nature"},
		},
		{
			FilePath: "mountain.jpg",
			Title:    "Mountain Landscape",
			Artist:   "Test Artist",
			Tags:     []string{"mountain", "landscape", "nature", "outdoor"},
		},
		{
			FilePath: "city.jpg",
			Title:    "City at Night",
			Artist:   "Test Artist",
			Tags:     []string{"city", "urban", "night", "lights"},
		},
	}

	// Check if test_images directory has any jpg/png files
	files, err := os.ReadDir(testDataDir)
	if err == nil && len(files) > 0 {
		realImages := make([]TestImage, 0)
		for _, file := range files {
			if !file.IsDir() {
				ext := filepath.Ext(file.Name())
				if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
					realImages = append(realImages, TestImage{
						FilePath: filepath.Join(testDataDir, file.Name()),
						Title:    file.Name(),
						Artist:   "Test User",
						Tags:     []string{"test", "sample"},
					})
				}
			}
		}
		if len(realImages) > 0 {
			return realImages
		}
	}

	return images
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

func uploadImage(img TestImage) (string, error) {
	// Check if file exists
	if _, err := os.Stat(img.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", img.FilePath)
	}

	// Open file
	file, err := os.Open(img.FilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image file
	part, err := writer.CreateFormFile("image", filepath.Base(img.FilePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	// Add other fields
	writer.WriteField("title", img.Title)
	writer.WriteField("artist", img.Artist)

	// Add tags as JSON array
	tagsJSON, _ := json.Marshal(img.Tags)
	writer.WriteField("tags", string(tagsJSON))

	writer.Close()

	// Send request
	req, err := http.NewRequest("POST", baseURL+"/images/upload", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", err
	}

	return uploadResp.ImageID, nil
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

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
