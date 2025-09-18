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

	time.Sleep(5 * time.Second)

	wd.SetImplicitWaitTimeout(10 * time.Second)

	shareBtn, err := wd.FindElement(selenium.ByXPATH,
		"//button[contains(., 'Baham ko‘rish')] | //button[contains(., 'Поделиться')] | //button[contains(., 'Share')]")
	if err == nil {
		_ = shareBtn.Click()
		time.Sleep(2 * time.Second)
	}

	buf, _ := wd.Screenshot()
	os.WriteFile("debug.png", buf, 0644)

	wd.SetImplicitWaitTimeout(15 * time.Second)

	elems, err := wd.FindElements(selenium.ByCSSSelector, "div.card-share-view__text")
	fmt.Println("Coords elements found:", err)
	if err != nil {
		return 0, 0, fmt.Errorf("coords element not found: %w", err)
	}

	src, _ := wd.PageSource()
	fmt.Println(src)

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
