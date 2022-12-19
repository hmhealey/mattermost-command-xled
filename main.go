package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// I copied this value from the xled docs. I don't know if this specific value is needed or if it's just an example of
// a valid random string
const loginChallenge = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"

const treeIp = "http://192.168.50.117"

func loginAndVerify(c *Client) error {
	resp, err := c.Post("xled/v1/login", map[string]interface{}{
		"challenge": loginChallenge,
	})
	if err != nil {
		return err
	}

	token, ok := resp["authentication_token"].(string)
	if !ok {
		return err
	}
	c.SetAuthToken(token)

	resp, err = c.Post("xled/v1/verify", nil)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	c := MakeClient(treeIp)

	if err := loginAndVerify(c); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 && os.Args[1] == "cycle" {
		testCycle(c)
	} else {
		listenAndServe(c)
	}
}

func listenAndServe(c *Client) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/color":
			fields := strings.Split(req.FormValue("text"), " ")
			if len(fields) != 3 {
				http.Error(w, "invalid arguments", http.StatusBadRequest)
				return
			}

			red, redErr := parseColor(fields[0])
			green, greenErr := parseColor(fields[1])
			blue, blueErr := parseColor(fields[2])
			if redErr != nil || blueErr != nil || greenErr != nil {
				http.Error(w, "invalid arguments", http.StatusBadRequest)
				return
			}

			if _, err := c.Post("xled/v1/led/color", map[string]interface{}{
				"red":   red,
				"green": green,
				"blue":  blue,
			}); err != nil {
				http.Error(w, "failed to set color", http.StatusInternalServerError)
			}

			if _, err := c.Post("xled/v1/led/mode", map[string]interface{}{
				"mode": "color",
			}); err != nil {
				http.Error(w, "failed to set mode to color", http.StatusInternalServerError)
				return
			}

			writeJson(w, map[string]interface{}{
				"text": fmt.Sprintf("Color set to #%02x%02x%02x", red, green, blue),
			})
		case "/cycle":
			testCycle(c)

			writeJson(w, map[string]interface{}{
				"text": "Tree cycled",
			})
		case "/effect":
			effectID, err := strconv.Atoi(req.FormValue("text"))
			if err != nil {
				http.Error(w, "could not parse effect ID", http.StatusBadRequest)
				return
			}

			if _, err := c.Post("xled/v1/led/mode", map[string]interface{}{
				"mode":      "effect",
				"effect_id": effectID,
			}); err != nil {
				// TODO this also triggers when an invalid effect ID is passed
				http.Error(w, "failed to change effect", http.StatusInternalServerError)
			}

			writeJson(w, map[string]interface{}{
				"text": fmt.Sprintf("Tree effect set to %d", effectID),
			})
		case "/off":
			if _, err := c.Post("xled/v1/led/mode", map[string]interface{}{
				"mode": "off",
			}); err != nil {
				// TODO this also triggers when an invalid effect ID is passed
				http.Error(w, "failed to change effect", http.StatusInternalServerError)
			}

			writeJson(w, map[string]interface{}{
				"text": "Tree turned off",
			})
		default:
			fmt.Printf("Request received to %s\n", req.URL.Path)
			http.Error(w, "", http.StatusNotFound)
		}
	})

	s := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func testCycle(c *Client) {
	if _, err := c.Post("xled/v1/led/color", map[string]interface{}{
		"red":   255,
		"green": 0,
		"blue":  0,
	}); err != nil {
		log.Fatal(err)
	}

	if _, err := c.Post("xled/v1/led/mode", map[string]interface{}{
		"mode": "color",
	}); err != nil {
		log.Fatal(err)
	}

	wait()

	if _, err := c.Post("xled/v1/led/color", map[string]interface{}{
		"red":   0,
		"green": 255,
		"blue":  0,
	}); err != nil {
		log.Fatal(err)
	}

	wait()

	if _, err := c.Post("xled/v1/led/color", map[string]interface{}{
		"red":   0,
		"green": 0,
		"blue":  255,
	}); err != nil {
		log.Fatal(err)
	}

	wait()

	if _, err := c.Post("xled/v1/led/mode", map[string]interface{}{
		"mode":      "effect",
		"effect_id": 0,
	}); err != nil {
		log.Fatal(err)
	}
}

func wait() {
	time.Sleep(time.Second)
}
