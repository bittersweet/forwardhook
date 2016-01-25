package forwardhook

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var MaxRetries = 10

// MirrorRequest will POST through body and headers from an
// incoming http.Request.
// Failures are retried up to 10 times.
func MirrorRequest(r *http.Request, url string) {
	attempt := 1
	for {
		fmt.Printf("Attempting %s try=%d\n", url, attempt)

		client := &http.Client{}

		// Use body of incoming request
		req, err := http.NewRequest("POST", url, r.Body)
		if err != nil {
			log.Println("[error] http.NewRequest:", err)
		}

		// Use initial headers
		req.Header = r.Header

		resp, err := client.Do(req)
		if err != nil {
			log.Println("[error] client.Do:", err)
			time.Sleep(10 * time.Second)
		} else {
			fmt.Printf("[success] %s status=%d\n", url, resp.StatusCode)
			break
		}

		attempt++
		if attempt > MaxRetries {
			fmt.Println("[error] MaxRetries reached")
			break
		}
	}
}
