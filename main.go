package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func main() {
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--disable-gpu",
			"--no-sandbox",
			// "--headless", // uncomment to run in headless mode
		},
	}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, "http://localhost:9515/wd/hub")
	if err != nil {
		log.Fatalf("Failed to connect to ChromeDriver: %v", err)
	}
	defer wd.Quit()

	url := "https://yandex.uz/maps/org/175039152082"
	fmt.Println("Navigating to:", url)
	if err := wd.Get(url); err != nil {
		log.Fatalf("Failed to open URL: %v", err)
	}

	time.Sleep(5 * time.Second)

	shareBtn, err := wd.FindElement(selenium.ByCSSSelector, "button[aria-label='Baham ko‘rish']")
	if err != nil {
		log.Fatalf("Share button not found: %v", err)
	}

	if err := shareBtn.Click(); err != nil {
		log.Fatalf("Failed to click share button: %v", err)
	}

	time.Sleep(2 * time.Second)

	elems, err := wd.FindElements(selenium.ByCSSSelector, "div.card-share-view__text")
	if err != nil {
		log.Fatalf("No elements found: %v", err)
	}

	if len(elems) < 2 {
		log.Fatalf("Coordinates not found")
	}

	coords, err := elems[1].Text()
	if err != nil {
		log.Fatalf("Failed to extract coordinates text: %v", err)
	}

	fmt.Println("Extracted coordinates:", coords)

	parts := strings.Split(coords, ",")
	if len(parts) == 2 {
		lat := strings.TrimSpace(parts[0])
		lng := strings.TrimSpace(parts[1])
		fmt.Printf("Latitude: %s\nLongitude: %s\n", lat, lng)
	} else {
		log.Fatalf("Invalid coordinates format: %s", coords)
	}

}
