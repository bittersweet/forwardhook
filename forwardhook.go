package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// maxRetries indicates the maximum amount of retries we will perform before
// giving up
var maxRetries = 10

// Config contains the sites parsed from FORWARDHOOK_SITES
type Config struct {
	Sites string
}

// mirrorRequest will POST through body and headers from an
// incoming http.Request.
// Failures are retried up to 10 times.
func mirrorRequest(h http.Header, body []byte, url string) {
	attempt := 1
	for {
		fmt.Printf("Attempting %s try=%d\n", url, attempt)

		client := &http.Client{}

		// Use body of incoming request
		bR := bytes.NewReader(body)
		req, err := http.NewRequest("POST", url, bR)
		if err != nil {
			log.Println("[error] http.NewRequest:", err)
		}

		// Set headers from request
		req.Header = h

		resp, err := client.Do(req)
		if err != nil {
			log.Println("[error] client.Do:", err)
			time.Sleep(10 * time.Second)
		} else {
			fmt.Printf("[success] %s status=%d\n", url, resp.StatusCode)
			break
		}

		attempt++
		if attempt > maxRetries {
			fmt.Println("[error] maxRetries reached")
			break
		}
	}
}

// parseSites gets sites out of the FORWARDHOOK_SITES environment variable.
// There is no validation at the moment but you can add 1 or more sites,
// separated by commas.
func parseSites() []string {
	var c Config

	err := envconfig.Process("forwardhook", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	strings := strings.Split(c.Sites, ",")
	return strings
}

func handleHook(sites []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("fail on readall")
		}

		for _, url := range sites {
			go mirrorRequest(r.Header, body, url)
		}

		w.WriteHeader(http.StatusOK)
	})
}

func main() {
	sites := parseSites()
	fmt.Println("Will forward hooks to:", sites)

	http.Handle("/", handleHook(sites))

	fmt.Printf("Listening on port 8000\n")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
