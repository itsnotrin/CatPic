package main

import (
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	imageDir = "/mnt/nas/"
)

func main() {
	app := fiber.New()

	// Initialize random with a time-based seed
	rand.Seed(time.Now().UnixNano())

	app.Get("/", func(c *fiber.Ctx) error {
		filename := c.Query("filename")

		var imagePath string
		var err error

		if filename == "" {
			// Get a random image
			imagePath, err = getRandomImage(imageDir)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error finding images: " + err.Error())
			}
		} else {
			// Prevent directory traversal by checking for suspicious characters
			if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
			}

			// Only use basename to prevent any path manipulation
			cleanFilename := filepath.Base(filename)

			// Get specific image by filename
			imagePath = filepath.Join(imageDir, cleanFilename)

			// Verify the resulting path is still within the imageDir
			if !strings.HasPrefix(imagePath, imageDir) {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
			}

			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				return c.Status(fiber.StatusNotFound).SendString("Image not found")
			}
		}

		// Get just the filename part for the response
		filename = filepath.Base(imagePath)

		// Set content disposition for downloads
		c.Set("Content-Disposition", "inline; filename="+filename)

		// Set headers for embeds in social platforms
		c.Set("Content-Type", getContentType(imagePath))
		c.Set("og:image", c.BaseURL()+"/"+filename)
		c.Set("og:title", filename)
		c.Set("og:description", "Random image: "+filename)

		// Return the image file
		return c.SendFile(imagePath)
	})

	log.Fatal(app.Listen(":3000"))
}

func getRandomImage(dir string) (string, error) {
	var images []string

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isImageFile(path) {
			images = append(images, path)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(images) == 0 {
		return "", fiber.NewError(fiber.StatusNotFound, "No images found")
	}

	return images[rand.Intn(len(images))], nil
}

func isImageFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}
