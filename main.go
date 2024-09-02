package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var client = &http.Client{}

// Regex patterns for each protocol
var regexPatterns = map[string]string{
	"ss":     `(?i)(.{3})ss://`,    // Case-insensitive match for "ss"
	"vmess":  `(?i)vmess://`,        // Case-insensitive match for "vmess"
	"trojan": `(?i)trojan://`,       // Case-insensitive match for "trojan"
	"vless":  `(?i)vless://`,        // Case-insensitive match for "vless"
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func main() {
	// Load channel configuration
	configFile := "channels.txt"
	channels, err := readLines(configFile)
	if err != nil {
		log.Fatalf("Error reading channels file: %s", err)
	}

	configs := initializeConfigs()

	// Process each channel to extract configurations
	for _, channel := range channels {
		processChannel(channel, configs)
	}

	// Write unique configuration contents to respective files
	for proto, configContent := range configs {
		if configContent != "" {
			WriteToFile(RemoveDuplicate(configContent), proto+"_iran.txt")
		}
	}
}

func initializeConfigs() map[string]string {
	return map[string]string{
		"ss":     "",
		"vmess":  "",
		"trojan": "",
		"vless":  "",
		"mixed":  "",
	}
}

func processChannel(channel string, configs map[string]string) {
	allMessages := strings.Contains(channel, "{all_messages}")
	if allMessages {
		channel = strings.Split(channel, "{all_messages}")[0]
	}

	req, err := http.NewRequest("GET", channel, nil)
	if err != nil {
		log.Fatalf("Error creating request to: %s, Error: %s", channel, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error during HTTP request: %s", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing response: %s", err)
	}

	messages := doc.Find(".tgme_widget_message_wrap").Length()
	link, exist := doc.Find(".tgme_widget_message_wrap .js-widget_message").Last().Attr("data-post")

	// Load more messages if fewer than 100
	if messages < 100 && exist {
		number := strings.Split(link, "/")[1]
		doc = GetMessages(100, doc, number, channel)
	}

	// Extract configurations based on the message type
	if allMessages {
		extractAllMessages(doc, configs)
	} else {
		extractCodePreMessages(doc, configs)
	}
}

func extractAllMessages(doc *goquery.Document, configs map[string]string) {
	fmt.Println("Total messages:", doc.Find(".js-widget_message_wrap").Length())
	doc.Find(".tgme_widget_message_text").Each(func(i int, s *goquery.Selection) {
		messageText := s.Text()
		lines := strings.Split(messageText, "\n")
		extractConfigs(lines, configs)
	})
}

func extractCodePreMessages(doc *goquery.Document, configs map[string]string) {
	doc.Find("code,pre").Each(func(i int, s *goquery.Selection) {
		messageText := s.Text()
		lines := strings.Split(messageText, "\n")
		extractConfigs(lines, configs)
	})
}

func extractConfigs(lines []string, configs map[string]string) {
	for _, line := range lines {
		for protoRegex, regexValue := range regexPatterns {
			re := regexp.MustCompile(regexValue)
			if re.MatchString(line) {
				configs[protoRegex] += line + "\n"
			}
		}
	}
}

func WriteToFile(fileContent, filePath string) {
	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		log.Fatalf("Error writing to file %s: %s", filePath, err)
	}
	fmt.Printf("File %s written successfully\n", filePath)
}

func loadMore(link string) *goquery.Document {
	resp, err := client.Get(link)
	if err != nil {
		log.Println("Error making request:", err)
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Error parsing response body:", err)
		return nil
	}
	return doc
}

func GetMessages(length int, doc *goquery.Document, number string, channel string) *goquery.Document {
	x := loadMore(channel + "?before=" + number)
	if x == nil {
		return doc
	}

	doc.Find("body").AppendSelection(x.Find("body").Children())
	messages := doc.Find(".js-widget_message_wrap").Length()
	fmt.Println("Total messages after loading more:", messages)

	if messages < length {
		num, _ := strconv.Atoi(number)
		n := num - 21
		if n > 0 {
			ns := strconv.Itoa(n)
			return GetMessages(length, doc, ns, channel)
		}
	}

	return doc
}

func RemoveDuplicate(config string) string {
	lines := strings.Split(config, "\n")
	uniqueLines := make(map[string]bool)

	for _, line := range lines {
		if len(line) > 0 {
			uniqueLines[line] = true
		}
	}

	return strings.Join(getKeys(uniqueLines), "\n")
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
