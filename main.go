package main

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const imageDir = "/mnt/All Disks/Files/cats"

// Global cache of image paths relative to imageDir
var imageCache []string

func main() {
	// Initialize image cache at startup
	if err := buildImageCache(imageDir); err != nil {
		log.Fatalf("Fatal: Failed to build image cache: %v", err)
	}
	log.Printf("Loaded %d images", len(imageCache))

	app := fiber.New()

	// Root path — generate random slug and redirect
	app.Get("/", func(c *fiber.Ctx) error {
		// Force Discord/Browsers to not cache this redirect
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		slug := strconv.FormatInt(time.Now().UnixNano(), 36) // unique slug like "lqx1uc0puv"
		return c.Redirect("/view/"+slug, fiber.StatusTemporaryRedirect)
	})

	// Direct random image endpoint (useful for bots/embeds that just want an image)
	app.Get("/random", func(c *fiber.Ctx) error {
		relPath, err := getRandomImage()
		if err != nil {
			return c.Status(500).SendString("No cat found 😿")
		}

		imagePath := filepath.Join(imageDir, relPath)

		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Content-Type", getContentType(imagePath))

		return c.SendFile(imagePath)
	})

	// View a random cat (slug is unused, just ensures a fresh URL)
	app.Get("/view/:slug", func(c *fiber.Ctx) error {
		relPath, err := getRandomImage()
		if err != nil {
			return c.Status(500).SendString("No cat found 😿")
		}

		// URL encode the path parts to ensure valid URLs
		// We use forward slashes for URLs regardless of OS
		urlPath := filepath.ToSlash(relPath)
		// Simple encoding for the path
		urlPath = url.PathEscape(urlPath)

		imageURL := c.BaseURL() + "/image/" + urlPath + "?t=" + strconv.FormatInt(time.Now().Unix(), 10)
		pageURL := c.BaseURL() + "/view/" + c.Params("slug")

		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta property="og:title" content="Random Cat!" />
  <meta property="og:description" content="Here’s a cat just for you 🐱" />
  <meta property="og:image" content="%s" />
  <meta property="og:url" content="%s" />
  <meta property="twitter:card" content="summary_large_image" />
  <title>Random Cat</title>
  <script defer data-domain="cats.ryanwiecz.co.uk" src="https://analytics.ryanwiecz.co.uk/js/script.js"></script>
</head>
<body>
  <h1>Random Cat 🐾</h1>
  <img src="%s" alt="Random Cat" style="max-width: 100%%; border-radius: 12px;" />
  <p><a href="/">Next Cat</a></p>
</body>
</html>`, imageURL, pageURL, imageURL)

		c.Set("Cache-Control", "no-store")
		c.Set("Content-Type", "text/html")

		return c.SendString(html)
	})

	// Serve specific image files
	// Use * to capture the full path including slashes
	app.Get("/image/*", func(c *fiber.Ctx) error {
		// Get the wildcard parameter
		filename := c.Params("*")

		// Basic security check against directory traversal
		// cleanPath removes .. and resolves separators
		// Note: fiber's router might already decode the URL, but we should be careful.
		// We want to ensure the requested file is inside imageDir.

		// Decode URL path if needed (Fiber usually does this for params, but * might be raw-ish depending on config.
		// Default behavior is usually decoded).
		decodedFilename, err := url.PathUnescape(filename)
		if err == nil {
			filename = decodedFilename
		}

		if strings.Contains(filename, "..") {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid filename")
		}

		imagePath := filepath.Join(imageDir, filename)

		// Verify the resolved path is still within imageDir (extra safety)
		// This handles cases where filename might be absolute or sneaky
		cleanImageDir := filepath.Clean(imageDir)
		cleanPath := filepath.Clean(imagePath)
		if !strings.HasPrefix(cleanPath, cleanImageDir) {
			// On Windows this prefix check might be tricky with case sensitivity,
			// but the target is likely Linux based on the path.
			// We'll keep it simple.
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}

		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("Image not found")
		}

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

func buildImageCache(dir string) error {
	var images []string

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// If we can't access a specific file/dir, just log and continue?
			// Or fail? The original code returned err. Let's just return err.
			return err
		}
		if !info.IsDir() && isImageFile(path) {
			// Store path relative to base dir
			rel, err := filepath.Rel(dir, path)
			if err == nil {
				images = append(images, rel)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	imageCache = images
	return nil
}

func getRandomImage() (string, error) {
	for len(imageCache) > 0 {
		idx := rand.Intn(len(imageCache))
		relPath := imageCache[idx]
		fullPath := filepath.Join(imageDir, relPath)

		if _, err := os.Stat(fullPath); err == nil {
			return relPath, nil
		}

		// File missing, remove from cache (swap with last and shorten)
		log.Printf("Removing stale image from cache: %s", relPath)
		imageCache[idx] = imageCache[len(imageCache)-1]
		imageCache = imageCache[:len(imageCache)-1]
	}

	return "", fiber.NewError(fiber.StatusNotFound, "No images found")
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
