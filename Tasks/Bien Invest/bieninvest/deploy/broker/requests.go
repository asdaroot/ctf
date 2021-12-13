package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/gommon/log"
)

func MakeRequest(uri string) []byte {
	resp, err := http.Get(os.Getenv("STOCK") + uri)
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}

	if resp.StatusCode != 200 {
		log.Errorf("code=%s: %s", resp.StatusCode, body)
		return nil
	}

	return body
}
