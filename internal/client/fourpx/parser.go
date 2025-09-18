package fourpx

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shamil/proxy_track_service-1/pkg/models"
)

func ParseHTML(htmlContent string, trackCodes []string) (map[string]*models.TrackData, error) {
	results := make(map[string]*models.TrackData)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	for _, trackCode := range trackCodes {
		trackData := parseTrackData(doc, trackCode)
		if trackData != nil {
			results[trackCode] = trackData
		} else {
			results[trackCode] = &models.TrackData{
				Countries: []string{"Unknown", "Unknown"},
				Events:    []models.Event{},
			}
		}
	}

	return results, nil
}

func parseTrackData(doc *goquery.Document, trackCode string) *models.TrackData {
	listItem := doc.Find(fmt.Sprintf(".next-list-item:contains('%s')", trackCode)).First()
	if listItem.Length() == 0 {
		return nil
	}

	return &models.TrackData{
		Countries: extractCountries(listItem),
		Events:    extractEvents(doc),
	}
}

func extractCountries(listItem *goquery.Selection) []string {
	countryText := listItem.Find("small").Text()
	if countryText == "" {
		return []string{"Unknown", "Unknown"}
	}

	parts := strings.Split(countryText, " - ")
	if len(parts) == 2 {
		return []string{
			mapCountryCode(strings.TrimSpace(parts[0])),
			strings.TrimSpace(parts[1]),
		}
	}

	return []string{"Unknown", "Unknown"}
}

func extractEvents(doc *goquery.Document) []models.Event {
	var events []models.Event

	doc.Find(".next-timeline-item").Each(func(i int, s *goquery.Selection) {
		event := extractTimelineEvent(s)
		if event != nil {
			events = append(events, *event)
		}
	})

	return events
}

func extractTimelineEvent(timelineItem *goquery.Selection) *models.Event {
	dateTimeText := timelineItem.Find(".next-timeline-item-left-content").Text()
	dateTime := extractDateTime(dateTimeText)
	if dateTime == "" {
		return nil
	}

	statusText := timelineItem.Find(".next-timeline-item-body").Text()
	status := cleanStatusText(statusText)
	if status == "" {
		return nil
	}

	return &models.Event{
		Status: status,
		Date:   parseDate(dateTime),
	}
}

func extractDateTime(text string) string {
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	if match := re.FindString(text); match != "" {
		return match
	}
	return ""
}

func cleanStatusText(text string) string {
	text = strings.TrimSpace(text)

	text = strings.ReplaceAll(text, "UTC+08:00", "")
	text = strings.TrimSpace(text)

	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return text
}

func mapCountryCode(countryName string) string {
	countryMap := map[string]string{
		"China":        "CN",
		"Russian":      "RU",
		"Russia":       "RU",
		"Kazakhstan":   "KZ",
		"Kazakh":       "KZ",
		"Germany":      "DE",
		"German":       "DE",
		"USA":          "US",
		"UnitedStates": "US",
		"US":           "US",
	}

	if code, exists := countryMap[countryName]; exists {
		return code
	}

	for name, code := range countryMap {
		if strings.Contains(countryName, name) {
			return code
		}
	}

	if len(countryName) == 2 {
		return strings.ToUpper(countryName)
	}

	return countryName
}

func parseDate(dateStr string) string {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return dateStr
}
