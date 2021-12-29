package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
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

//UserAuth ...
type UserAuth struct {
	conf  oauth2.Config
	code  string
	token oauth2.Token
}

func getAuthCode(w *http.ResponseWriter, r http.Request) string {
	data := []byte{}
	r.Body.Read(data)
	return string(data)
}

func (userAuth *UserAuth) init() (string, error) {
	userAuth.conf = oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"user-read-recently-played", "user-read-playback-state", "user-top-read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
		RedirectURL: "http://localhost:8080/oauth/redirect",
	}

	url := userAuth.conf.AuthCodeURL("")
	fmt.Printf("Visit the URL for the auth dialog: %v\n\n", url)

	codeValue := ""
	server := &http.Server{Addr: ":8080", Handler: nil}
	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		codeValue = r.FormValue("code")
		w.Header().Set("content-type", "text/plain")
		fmt.Fprintln(w, "You can close the browser window now")
		if err = server.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error: %s", err)
		}
	})
	err := server.ListenAndServe()
	userAuth.code = codeValue

	return codeValue, err
}

func (userAuth *UserAuth) getUserAuthAccessToken(ctx *context.Context) error {
	tok, err := userAuth.conf.Exchange(*ctx, userAuth.code)
	if err != nil {
		return err
	}

	userAuth.token = *tok
	return err
}

func getData(url string, accessToken string) (map[string]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	req.Header = map[string][]string{
		"Authorization": {"Bearer " + accessToken},
	}
	client := http.Client{}

	response, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(respBody))
	respMap := make(map[string]string)
	json.Unmarshal(respBody, &respMap)

	return respMap, nil
}

func main() {
	err := godotenv.Load("env/.env")
	if err != nil {
		panic(err.Error())
	}

	userAuth := UserAuth{}
	_, err = userAuth.init()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	ctx := context.Background()
	userAuth.getUserAuthAccessToken(&ctx)

	fmt.Println("token: " + userAuth.token.AccessToken)
	userData, err := getData("https://api.spotify.com/v1/me", userAuth.token.AccessToken)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println(userData)
}
