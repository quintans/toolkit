// Idea: Each event is cached with its version.
// The clients sends a list o events with the last received event version.
// if the version is different for an cached event, and therefore pending, the event is immediatly reurned,
// if not the client waits for a new event

package poller

import (
	"container/list"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	Tokens  map[string]int64
	Channel chan []Message
}

type Message struct {
	timestamp time.Time
	Version   int64       `json:"version"`
	Name      string      `json:"name"`
	Data      interface{} `json:"data"`
}

type Poller struct {
	addClient     chan Client
	pipe          chan Message
	timeout       time.Duration
	removeChannel chan chan []Message
}

func NewPoller(timeout time.Duration) *Poller {
	this := new(Poller)
	this.addClient = make(chan Client, 1)
	this.pipe = make(chan Message, 1)
	this.removeChannel = make(chan chan []Message, 1)
	this.timeout = timeout

	channels := list.New()
	tick := time.Tick(this.timeout)
	var version int64
	var messages = make(map[string]Message)
	var clients = make(map[chan []Message]*list.Element)
	go func() {
		for {
			select {
			case c := <-this.addClient:
				// gather pending events.
				pending := make([]Message, 0)
				for k, v := range messages {
					// check if the version of the stored event has changed.
					if v.Version != 0 && v.Version != c.Tokens[k] {
						pending = append(pending, v)
					}
				}
				// if there are any pending events for this client, send them immediatly
				if len(pending) > 0 {
					c.Channel <- pending
				} else {
					// queue it
					e := channels.PushBack(c)
					// track it for removal
					clients[c.Channel] = e
				}

			case c := <-this.removeChannel:
				if e, ok := clients[c]; ok {
					channels.Remove(e)
					delete(clients, c)
				}

			case m := <-this.pipe:
				// add to message buffer
				version++
				m.Version = version
				// replaces previous event
				messages[m.Name] = m
				msgs := []Message{m}
				// brodcast
				var next *list.Element
				for e := channels.Front(); e != nil; e = next {
					next = e.Next()

					c := e.Value.(Client)
					// if it is listening to this event, send it
					if _, ok := c.Tokens[m.Name]; ok {
						// remove from list
						channels.Remove(e)
						c.Channel <- msgs
					}
				}

			case <-tick:
				// delete expired messages
				mark := time.Now().Add(-this.timeout)
				for k, v := range messages {
					if v.timestamp.Before(mark) {
						delete(messages, k)
					}
				}

			}
		}
	}()

	return this
}

func (this *Poller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// gets list of events and versions
	parameters := r.URL.Query()
	tokens := make(map[string]int64)
	for k, v := range parameters {
		if len(v) > 0 {
			version, err := strconv.ParseInt(v[0], 0, 64)
			if err == nil {
				tokens[k] = version
			}
		}
	}
	// message channel
	message := make(chan []Message, 1)
	// add this client
	this.addClient <- Client{Tokens: tokens, Channel: message}

	select {
	case <-time.After(this.timeout):
		this.removeChannel <- message
		sendMessage(w, []Message{Message{Version: 0}})

	case msg := <-message:
		sendMessage(w, msg)
	}
}

func sendMessage(w http.ResponseWriter, m []Message) {
	w.Header()["Content-Type"] = []string{"application/json"}
	raw, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(503)
	} else {
		w.Write(raw)
	}
}

func (this *Poller) Broadcast(name string, data interface{}) {
	this.pipe <- Message{
		timestamp: time.Now(),
		Name:      name,
		Data:      data,
	}
}
