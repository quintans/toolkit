package main

import (
	"github.com/quintans/toolkit/web/poller"

	"net/http"
	"runtime"
	"time"

	"fmt"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()*2 + 1)

	http.Handle("/", http.FileServer(http.Dir("./www/")))

	poll := poller.NewPoller(10 * time.Second)
	http.Handle("/feed", poll)

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		data := r.URL.Query().Get("data")
		poll.Broadcast("boardChange", data+" @ "+time.Now().String())
	})

	url := ":8080"
	fmt.Println("Listening @", url)
	if err := http.ListenAndServe(url, nil); err != nil {
		panic(err)
	}
}
