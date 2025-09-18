package coords

import (
	"fmt"
	"os"
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

	fmt.Println("Connecting to Selenium WebDriver at", seleniumURL)

	wd, err := selenium.NewRemote(caps, seleniumURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to connect to selenium: %w", err)
	}
	defer wd.Quit()

	fmt.Println("Navigating to:", url)

	if err := wd.Get(url); err != nil {
		return 0, 0, fmt.Errorf("failed to open url: %w", err)
	}

	fmt.Println("Waiting for page to load...")
	wd.SetImplicitWaitTimeout(10 * time.Second)

	// Cookie popup tugmasini yopish
	cookieBtn, err := wd.FindElement(selenium.ByXPATH,
		"//div[contains(@class,'gdpr-popup-v3-button')]")
	if err == nil {
		fmt.Println("Cookie popup found, clicking...")
		_ = cookieBtn.Click()
		time.Sleep(2 * time.Second)
	} else {
		fmt.Println("No cookie popup found:", err)
	}

	// Share tugmasini bosish
	shareBtn, err := wd.FindElement(selenium.ByXPATH,
		"//button[contains(., 'Baham ko‘rish')] | //button[contains(., 'Поделиться')] | //button[contains(., 'Share')]")
	if err == nil {
		_ = shareBtn.Click()
		time.Sleep(2 * time.Second)
	} else {
		fmt.Println("Share button not found:", err)
	}

	// Debug uchun screenshot
	buf, _ := wd.Screenshot()
	_ = os.WriteFile("debug.png", buf, 0644)

	// 1️⃣ Variant: div.card-share-view__text dan olish
	elems, _ := wd.FindElements(selenium.ByCSSSelector, "div.card-share-view__text")
	for _, el := range elems {
		txt, _ := el.Text()
		fmt.Println("Coords element found:", txt)
		if strings.Contains(txt, ",") {
			parts := strings.Split(txt, ",")
			if len(parts) == 2 {
				latStr := strings.TrimSpace(parts[0])
				lngStr := strings.TrimSpace(parts[1])

				lat, err1 := strconv.ParseFloat(latStr, 64)
				lng, err2 := strconv.ParseFloat(lngStr, 64)
				if err1 == nil && err2 == nil {
					return lat, lng, nil
				}
			}
		}
	}

	// 2️⃣ Variant: input.card-share-view__link-input dan olish
	inputElem, err := wd.FindElement(selenium.ByCSSSelector, "input.card-share-view__link-input")
	if err == nil {
		val, _ := inputElem.GetAttribute("value")
		fmt.Println("Share link found:", val)
		if strings.Contains(val, "ll=") {
			parts := strings.Split(val, "ll=")
			if len(parts) > 1 {
				coordsPart := strings.Split(parts[1], "&")[0]
				xy := strings.Split(coordsPart, "%2C")
				if len(xy) == 2 {
					lng, err1 := strconv.ParseFloat(xy[0], 64)
					lat, err2 := strconv.ParseFloat(xy[1], 64)
					if err1 == nil && err2 == nil {
						return lat, lng, nil
					}
				}
			}
		}
	}

	return 0, 0, fmt.Errorf("coordinates not found")
}
