package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
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
		"ss":     `(.{3})ss:\/\/`,
		"vmess":  `vmess:\/\/`,
		"trojan": `trojan:\/\/`,
		"vless":  `vless:\/\/`,
	}

	for _, channel := range channels {
		processChannel(channel, configs, myregex)
	}

	for proto, configContent := range configs {
		WriteToFile(RemoveDuplicate(configContent), proto+"_iran.txt")
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
		log.Fatalf("Error when requesting to: %s Error: %s", channel, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	messages := doc.Find(".tgme_widget_message_wrap").Length()
	link, exist := doc.Find(".tgme_widget_message_wrap .js-widget_message").Last().Attr("data-post")

	if messages < 100 && exist {
		number := strings.Split(link, "/")[1]
		fmt.Println(number)
		doc = GetMessages(100, doc, number, channel)
	}

	if allMessages {
		fmt.Println(doc.Find(".js-widget_message_wrap").Length())
		doc.Find(".tgme_widget_message_text").Each(func(j int, s *goquery.Selection) {
			messageText := s.Text()
			lines := strings.Split(messageText, "\n")
			extractConfigs(lines, configs, myregex)
		})
	} else {
		doc.Find("code,pre").Each(func(j int, s *goquery.Selection) {
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
			line = re.ReplaceAllStringFunc(line, func(match string) string {
				return "\n" + match
			})

			if len(strings.Split(line, "\n")) > 1 {
				myConfigs := strings.Split(line, "\n")
				for _, myConfig := range myConfigs {
					if myConfig != "" {
						re := regexp.MustCompile(regexValue)
						myConfig = strings.ReplaceAll(myConfig, " ", "")
						match := re.FindStringSubmatch(myConfig)
						if len(match) >= 1 {
							switch protoRegex {
							case "ss":
								if match[1][:3] == "vme" {
									configs["vmess"] += "\n" + myConfig + "\n"
								} else if match[1][:3] == "vle" {
									configs["vless"] += "\n" + myConfig + "\n"
								} else {
									configs["ss"] += "\n" + myConfig[3:] + "\n"
								}
							default:
								configs[protoRegex] += "\n" + myConfig + "\n"
							}
						}
					}
				}
			}
		}
	}
}

func WriteToFile(fileContent string, filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		err = os.WriteFile(filePath, []byte{}, 0644)
		if err != nil {
			fmt.Println("Error clearing file:", err)
			return
		}
	} else if os.IsNotExist(err) {
		_, err = os.Create(filePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
	} else {
		fmt.Println("Error checking file:", err)
		return
	}

	err = os.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("File written successfully")
}

func loadMore(link string) *goquery.Document {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil
	}

	resp, err := client.Do(req)
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
	fmt.Println(messages)

	if messages > length {
		return doc
	} else {
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

	uniqueString := strings.Join(getKeys(uniqueLines), "\n")
	return uniqueString
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
