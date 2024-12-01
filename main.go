package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gtuk/discordwebhook"
)

var wg sync.WaitGroup

var (
	token               string   = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	tracking_list       []string = []string{"XXXXXXXXXXXX", "XXXXXXXXXXXXXX", "XXXXXXXXXXXXXXX"}
	discord_webhook_url string   = "https://discord.com/api/webhooks/XXXXXXXXXXXXXXXX/XXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
)

func main() {
	for _, packageID := range tracking_list {
		wg.Add(1)
		go ChangedStateChecker(packageID)
	}
	wg.Wait()
}

func ChangedStateChecker(packageid string) {
	defer wg.Done()
	var Statecache, Statecache2 string

	for {
		Statecache = GetInfo(packageid)
		time.Sleep(30 * time.Second)
		Statecache2 = GetInfo(packageid)
		if Statecache != Statecache2 {
			SendMessage(fmt.Sprintf("Update of the package : %v, New description : %v \n", packageid, Statecache2))
			fmt.Printf("Update of the package : %v, New description : %v \n", packageid, Statecache2)
		}
		fmt.Println("New update cycle")
	}
}

func SendMessage(content string) {
	username := "PackageInfo"
	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err := discordwebhook.SendMessage(discord_webhook_url, message)
	if err != nil {
		log.Fatal(err)
	}
}

func GetInfo(packageid string) string {
	url := "https://api.17track.net/track/v2.2/gettrackinfo"
	var resturndescr string

	body := []byte(fmt.Sprintf(`[{"number": "%s"}]`, packageid))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	req.Header.Add("17token", token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		panic(err)
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if accepted, ok := data["accepted"].([]interface{}); ok && len(accepted) > 0 {
			if trackInfo, ok := accepted[0].(map[string]interface{})["track_info"].(map[string]interface{}); ok {
				if latestEvent, ok := trackInfo["latest_event"].(map[string]interface{}); ok {
					if description, ok := latestEvent["description"].(string); ok {
						resturndescr = description
					}
				}
			}
		}
	}
	return resturndescr
}
