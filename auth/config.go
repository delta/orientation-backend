package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var is_dev = true

type authConfig struct {
	Dauth struct {
		Client_id     string   `json:"client_id"`
		Client_secret string   `json:"client_secret"`
		Redirect_uri  string   `json:"redirect_uri"`
		Grant_type    string   `json:"grant_type"`
		Nonce         string   `json:"nonce"`
		Response_type string   `json:"response_type"`
		State         string   `json:"state"`
		Scope         []string `json:"scope"`
	} `json:"dauth"`
	Cookie struct {
		User struct {
			Name    string `json:"name"`
			Expires int64  `json:"expiry_period"`
		} `json:"user_cookie"`
		Refresh struct {
			Name    string `json:"name"`
			Expires int64  `json:"expiry_period"`
		} `json:"refresh_cookie"`
		Jwt_secret string `json:"jwt_secret"`
	} `json:"cookie"`
}

// Loads Dauth Config from the dauth.config.json file and returns it
func getConfig() authConfig {
	jsonData, err := ioutil.ReadFile("dauth.config.json")
	if err != nil {
		panic(fmt.Errorf("error reading dauth.config.json file, %+v", err))
	}
	var config authConfig
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		panic(fmt.Errorf("json Config Error, %+v", err))
	}
	return config
}
