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

	wd, err := selenium.NewRemote(caps, seleniumURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to connect to selenium: %w", err)
	}
	defer wd.Quit()

	if err := wd.Get(url); err != nil {
		return 0, 0, fmt.Errorf("failed to open url: %w", err)
	}

	time.Sleep(5 * time.Second)

	shareBtn, err := wd.FindElement(selenium.ByCSSSelector, "button[aria-label='Baham ko‘rish']")
	if err == nil {
		_ = shareBtn.Click()
		time.Sleep(2 * time.Second)
	}

	elems, err := wd.FindElements(selenium.ByCSSSelector, "div.card-share-view__text")
	if err != nil {
		return 0, 0, fmt.Errorf("coords element not found: %w", err)
	}

	for _, el := range elems {
		txt, _ := el.Text()
		if strings.Contains(txt, ",") {
			parts := strings.Split(txt, ",")
			if len(parts) == 2 {
				latStr := strings.TrimSpace(parts[0])
				lngStr := strings.TrimSpace(parts[1])

				lat, err1 := strconv.ParseFloat(latStr, 64)
				lng, err2 := strconv.ParseFloat(lngStr, 64)
				if err1 != nil || err2 != nil {
					return 0, 0, fmt.Errorf("failed to parse coordinates: %v, %v", err1, err2)
				}

				return lat, lng, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("coordinates not found")
}
