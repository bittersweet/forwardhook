package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// maxRetries indicates the maximum amount of retries we will perform before
// giving up
var maxRetries = 10

// mirrorRequest will POST through body and headers from an
// incoming http.Request.
// Failures are retried up to 10 times.
func mirrorRequest(h http.Header, body []byte, url string) {
	attempt := 1
	for {
		fmt.Printf("Attempting %s try=%d\n", url, attempt)

		client := &http.Client{}

		rB := bytes.NewReader(body)
		req, err := http.NewRequest("POST", url, rB)
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
			resp.Body.Close()
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
	sites := os.Getenv("FORWARDHOOK_SITES")

	if sites == "" {
		log.Fatal("No sites set up, provide FORWARDHOOK_SITES")
	}

	s := strings.Split(sites, ",")
	return s
}

func handleHook(sites []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Fail on ReadAll")
		}
		r.Body.Close()

		for _, url := range sites {
			go mirrorRequest(r.Header, rB, url)
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
