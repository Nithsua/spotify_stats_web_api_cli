package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func getAppAuthAccessToken() (map[string]string, error) {

	data := os.Getenv("CLIENT_ID") + ":" + os.Getenv("CLIENT_SECRET")
	encodedClient := base64.StdEncoding.EncodeToString([]byte(data))
	fmt.Println(encodedClient)

	reqBody := []byte("grant_type=client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedClient)

	client := http.Client{}

	response, err := client.Do(req)
	if err != nil {
		return nil, err

	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	parsedBody := make(map[string]string)
	json.Unmarshal(body, &parsedBody)

	return parsedBody, nil
}

func main() {

}
