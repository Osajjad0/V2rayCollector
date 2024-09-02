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
	configFile := "channels.txt"
	channels, err := readLines(configFile)
	if err != nil {
		log.Fatal(err)
	}

	configs := map[string]string{
		"ss":     "",
		"vmess":  "",
		"trojan": "",
		"vless":  "",
		"mixed":  "",
	}

	myregex := map[string]string{
		"ss":     `(?i)(.{3})ss:\/\/`, // Case-insensitive match for protocols
		"vmess":  `(?i)vmess:\/\/`,
		"trojan": `(?i)trojan:\/\/`,
		"vless":  `(?i)vless:\/\/`,
	}

	for _, channel := range channels {
		processChannel(channel, configs, myregex)
	}

	// Write the unique config content to respective files.
	for proto, configContent := range configs {
		if configContent != "" {
			WriteToFile(RemoveDuplicate(configContent), proto+"_iran.txt")
		}
	}
}

func processChannel(channel string, configs map[string]string, myregex map[string]string) {
	allMessages := false
	if strings.Contains(channel, "{all_messages}") {
		allMessages = true
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

	if messages < 100 && exist {
		number := strings.Split(link, "/")[1]
		fmt.Println("Loading more messages starting from:", number)
		doc = GetMessages(100, doc, number, channel)
	}

	if allMessages {
		fmt.Println("Total messages:", doc.Find(".js-widget_message_wrap").Length())
		doc.Find(".tgme_widget_message_text").Each(func(i int, s *goquery.Selection) {
			messageText := s.Text()
			lines := strings.Split(messageText, "\n")
			extractConfigs(lines, configs, myregex)
		})
	} else {
		doc.Find("code,pre").Each(func(i int, s *goquery.Selection) {
			messageText := s.Text()
			lines := strings.Split(messageText, "\n")
			extractConfigs(lines, configs, myregex)
		})
	}
}

func extractConfigs(lines []string, configs map[string]string, myregex map[string]string) {
	for _, line := range lines {
		for protoRegex, regexValue := range myregex {
			re := regexp.MustCompile(regexValue)

			match := re.FindStringSubmatch(line)
			if match != nil {
				// Add the matched line to the corresponding protocol config.
				configs[protoRegex] += line + "\n"
			}
		}
	}
}

func WriteToFile(fileContent string, filePath string) {
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

	doc.Find("
