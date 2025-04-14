package main

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const imageDir = "/mnt/All Disks/Files/cats"

func main() {
	app := fiber.New()
	rand.Seed(time.Now().UnixNano())

	// Root path — generate random slug and redirect
	app.Get("/", func(c *fiber.Ctx) error {
		slug := strconv.FormatInt(time.Now().UnixNano(), 36) // unique slug like "lqx1uc0puv"
		return c.Redirect("/view/"+slug, fiber.StatusTemporaryRedirect)
	})

	// View a random cat (slug is unused, just ensures a fresh URL)
	app.Get("/view/:slug", func(c *fiber.Ctx) error {
		imagePath, err := getRandomImage(imageDir)
		if err != nil {
			return c.Status(500).SendString("No cat found 😿")
		}

		filename := filepath.Base(imagePath)
		imageURL := c.BaseURL() + "/image/" + filename + "?t=" + strconv.FormatInt(time.Now().Unix(), 10)

		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta property="og:title" content="Random Cat!" />
  <meta property="og:description" content="Here’s a cat just for you 🐱" />
  <meta property="og:image" content="%s" />
  <meta property="twitter:card" content="summary_large_image" />
  <title>Random Cat</title>
</head>
<body>
  <h1>Random Cat 🐾</h1>
  <img src="%s" alt="Random Cat" style="max-width: 100%%; border-radius: 12px;" />
  <p><a href="/">Next Cat</a></p>
</body>
</html>`, imageURL, imageURL)

		c.Set("Cache-Control", "no-store")
		c.Set("Content-Type", "text/html")

		return c.SendString(html)
	})

	// Serve specific image files
	app.Get("/image/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
		}

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
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
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
