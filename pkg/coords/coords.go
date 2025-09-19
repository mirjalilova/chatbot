package coords

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

const (
	seleniumURL = "http://selenium:4444/wd/hub"
)

func ExtractCoordinates(url string) (float64, float64, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeArgs := []string{
		"--disable-dev-shm-usage",
		"--no-sandbox",
		"--lang=uz",
	}
	chromeCaps := map[string]interface{}{
		"args": chromeArgs,
	}
	caps["goog:chromeOptions"] = chromeCaps
	fmt.Println("Connecting to Selenium WebDriver at", seleniumURL)

	wd, err := selenium.NewRemote(caps, seleniumURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to connect to selenium: %w", err)
	}
	defer wd.Quit()

	if !strings.Contains(url, "lang=uz") {
		if strings.Contains(url, "?") {
			url += "&lang=uz"
		} else {
			url += "?lang=uz"
		}
	}

	fmt.Println("Navigating to:", url)

	if err := wd.Get(url); err != nil {
		return 0, 0, fmt.Errorf("failed to open url: %w", err)
	}

	fmt.Println("Waiting for page to load...")

	time.Sleep(5 * time.Second)

	shareBtn, err := wd.FindElement(selenium.ByCSSSelector, "button[aria-label='Baham ko‘rish']")
	fmt.Println("Share button found:", err)
	if err == nil {
		_ = shareBtn.Click()
		time.Sleep(2 * time.Second)
	}

	var elems []selenium.WebElement

	for i := 0; i < 10; i++ {
		elems, _ = wd.FindElements(selenium.ByCSSSelector, "div.card-share-view__text")
		if len(elems) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Coords elements found:", err)
	if err != nil {
		return 0, 0, fmt.Errorf("coords element not found: %w", err)
	}

	for _, el := range elems {
		txt, _ := el.Text()
		fmt.Println("Coords element found:", txt)
		if strings.Contains(txt, ",") {
			parts := strings.Split(txt, ",")
			fmt.Println("Coords element parts found:", parts)
			if len(parts) == 2 {
				latStr := strings.TrimSpace(parts[0])
				lngStr := strings.TrimSpace(parts[1])

				fmt.Println("Coords found:", latStr, lngStr)

				lat, err1 := strconv.ParseFloat(latStr, 64)
				lng, err2 := strconv.ParseFloat(lngStr, 64)
				fmt.Println("Parsed coords:", lat, lng, err1, err2)
				if err1 != nil || err2 != nil {
					return 0, 0, fmt.Errorf("failed to parse coordinates: %v, %v", err1, err2)
				}

				return lat, lng, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("coordinates not found")
}
