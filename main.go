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
	imageDir = "/mnt/All Disks/Files/cats"
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
			imagePath, err = getRandomImage(imageDir)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error finding image: " + err.Error())
			}
			filename = filepath.Base(imagePath)
		} else {
			if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
			}
			imagePath = filepath.Join(imageDir, filepath.Base(filename))
			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				return c.Status(fiber.StatusNotFound).SendString("Image not found")
			}
		}

		// Create a cache-busted image URL
		imageURL := c.BaseURL() + "/image/" + filename + "?t=" + time.Now().Format("20060102150405")

		// HTML page with Open Graph tags
		html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Random Cat</title>
  <meta property="og:title" content="Random Cat: ` + filename + `" />
  <meta property="og:description" content="Another cute cat picture 🐱" />
  <meta property="og:image" content="` + imageURL + `" />
  <meta property="og:type" content="website" />
  <meta property="twitter:card" content="summary_large_image" />
</head>
<body>
  <img src="` + imageURL + `" alt="Cat Image" style="max-width: 100%; height: auto;" />
</body>
</html>`

		// Prevent caching of the preview page
		c.Set("Content-Type", "text/html")
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Surrogate-Control", "no-store")

		return c.SendString(html)
	})

	app.Get("/image/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		imagePath := filepath.Join(imageDir, filepath.Base(filename))

		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("Image not found")
		}

		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Surrogate-Control", "no-store")
		c.Set("Content-Type", getContentType(imagePath))
		c.Set("Content-Disposition", "inline; filename="+filename)

		return c.SendFile(imagePath)
	})

	app.Get("/image/:ts/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")

		// Same checks as before
		if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
		}

		imagePath := filepath.Join(imageDir, filepath.Base(filename))
		if !strings.HasPrefix(imagePath, imageDir) {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
		}
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("Image not found")
		}

		// Set the same headers
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Surrogate-Control", "no-store")
		c.Set("Content-Type", getContentType(imagePath))
		c.Set("Content-Disposition", "inline; filename="+filepath.Base(filename))

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
