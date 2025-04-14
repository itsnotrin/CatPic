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
		// Get the random image
		imagePath, err := getRandomImage(imageDir)
		if err != nil {
			return c.Status(500).SendString("No cat found 😿")
		}

		filename := filepath.Base(imagePath)
		imageURL := c.BaseURL() + "/image/" + filename + "?t=" + strconv.FormatInt(time.Now().Unix(), 10)

		// Get the user's settings from localStorage
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta property="og:title" content="Random Cat!" />
  <meta property="og:description" content="Here’s a cat just for you 🐱" />
  <meta property="og:image" content="%s" />
  <meta property="twitter:card" content="summary_large_image" />
  <title>Random Cat</title>
  <style>
    body {
      font-family: system-ui, sans-serif;
      background-color: #121212;
      color: #ffffff;
      display: flex;
      justify-content: center;
      align-items: center;
      flex-direction: column;
      padding: 2rem;
      margin: 0;
    }
    .card {
      background: #1e1e1e;
      border-radius: 12px;
      box-shadow: 0 4px 20px rgba(0,0,0,0.6);
      padding: 1rem;
      max-width: 400px;
      text-align: center;
      border: 1px solid #333;
    }
    .card img {
      max-width: 100%;
      border-radius: 8px;
    }
    .button {
      margin-top: 1rem;
      background: #6366f1;
      color: white;
      padding: 0.6rem 1.2rem;
      border: none;
      border-radius: 6px;
      font-size: 1rem;
      cursor: pointer;
      box-shadow: 0 0 10px rgba(99, 102, 241, 0.4);
      transition: background 0.2s ease;
    }
    .button:hover {
      background: #4f46e5;
    }
    body.dark {
      background-color: #121212;
      color: #ffffff;
    }
    body.light {
      background-color: #f7f7f7;
      color: #333;
    }
  </style>
</head>
<body class="dark">
  <div class="card">
    <h2>Random Cat 🐱</h2>
    <img src="%s" alt="A random cat" />
    <form method="GET" action="/">
      <button class="button">New Cat</button>
    </form>
  </div>

  <!-- Include confetti.js script -->
  <script src="https://cdn.jsdelivr.net/npm/canvas-confetti"></script>

  <script>
    // Get settings from localStorage
    const confettiEnabled = localStorage.getItem("confetti") === "true";
    const keyboardShortcutsEnabled = localStorage.getItem("keyboardShortcuts") === "true";
    const darkModeEnabled = localStorage.getItem("darkMode") === "true";

    // Apply Dark Mode
    if (darkModeEnabled) {
      document.body.classList.add('dark');
    } else {
      document.body.classList.add('light');
    }

    // Enable confetti if turned on in settings
    if (confettiEnabled) {
      document.addEventListener("DOMContentLoaded", function () {
        launchConfetti();
      });
    }

    // Keyboard Shortcuts (Enable 'N' to trigger New Cat)
    if (keyboardShortcutsEnabled) {
      document.addEventListener('keydown', function (e) {
        if (e.key === 'n' || e.key === 'N') {
          window.location.reload(); // Reload the page to show a new cat
        }
      });
    }

    // Function to trigger confetti animation
    function launchConfetti() {
      const confetti = window.confetti;
      confetti({
        particleCount: 200,
        spread: 70,
        origin: { y: 0.6 }
      });
    }

  </script>
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

	app.Get("/settings", func(c *fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Cat Settings</title>
  <style>
    body {
      font-family: system-ui, sans-serif;
      background-color: #f7f7f7;
      color: #333;
      display: flex;
      justify-content: center;
      align-items: center;
      flex-direction: column;
      padding: 2rem;
      margin: 0;
    }
    .settings-container {
      background-color: #ffffff;
      border-radius: 12px;
      box-shadow: 0 4px 20px rgba(0,0,0,0.1);
      padding: 1.5rem;
      width: 300px;
      text-align: center;
    }
    .button {
      background-color: #6366f1;
      color: white;
      padding: 0.8rem 1.2rem;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      margin-top: 1rem;
      font-size: 1rem;
    }
    .toggle {
      margin: 1rem 0;
    }
  </style>
</head>
<body>
  <div class="settings-container">
    <h2>Cat Settings</h2>
    <div class="toggle">
      <label>Enable Confetti</label>
      <input type="checkbox" id="confetti-toggle" />
    </div>
    <div class="toggle">
      <label>Enable Keyboard Shortcuts</label>
      <input type="checkbox" id="keyboard-toggle" />
    </div>
    <div class="toggle">
      <label>Enable Dark Mode</label>
      <input type="checkbox" id="darkmode-toggle" checked />
    </div>
    <button class="button" onclick="saveSettings()">Save Settings</button>
  </div>

  <!-- Include the confetti.js script -->
  <script src="https://cdn.jsdelivr.net/npm/canvas-confetti"></script>

  <script>
    // Check localStorage for user settings
    document.getElementById("confetti-toggle").checked = localStorage.getItem("confetti") === "true";
    document.getElementById("keyboard-toggle").checked = localStorage.getItem("keyboardShortcuts") === "true";
    document.getElementById("darkmode-toggle").checked = localStorage.getItem("darkMode") === "true";

    // Save user settings
    function saveSettings() {
      localStorage.setItem("confetti", document.getElementById("confetti-toggle").checked);
      localStorage.setItem("keyboardShortcuts", document.getElementById("keyboard-toggle").checked);
      localStorage.setItem("darkMode", document.getElementById("darkmode-toggle").checked);

      alert("Settings saved! Refreshing page...");
      window.location.reload(); // Reload to apply settings
    }
  </script>
</body>
</html>`

		c.Set("Content-Type", "text/html")
		return c.SendString(html)
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
