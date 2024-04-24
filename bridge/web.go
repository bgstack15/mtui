package bridge

import (
	"encoding/json"
	"net/http"
	"time"
)

func (b *Bridge) HandlePost(w http.ResponseWriter, r *http.Request) {
	commands := make([]*CommandResponse, 0)
	err := json.NewDecoder(r.Body).Decode(&commands)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	b.handlers_lock.RLock()
	for _, cmd := range commands {
		for _, handler := range b.handlers[cmd.Type] {
			select {
			case handler <- cmd:
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	b.handlers_lock.RUnlock()
}

// collects commands for a certain amount of time
func collectCommands(ch chan *CommandRequest, delay time.Duration) []*CommandRequest {
	cmd_list := make([]*CommandRequest, 0)
	then := time.Now().Add(delay)
	t := time.NewTimer(delay)

	for {
		select {
		case cmd := <-ch:
			// command received, add to list
			cmd_list = append(cmd_list, cmd)
		case <-t.C:
			// nothing received in the entire duration time
		}

		if time.Now().After(then) {
			// time is up, return collected commands, if any
			return cmd_list
		}
	}
}

func (b *Bridge) HandleGet(w http.ResponseWriter, r *http.Request) {
	then := time.Now().Add(20 * time.Second)
	var cmds []*CommandRequest
	for {
		// collect commands for at least 100ms
		cmds = collectCommands(b.tx_cmds, 100*time.Millisecond)

		if len(cmds) > 0 {
			// commands received, return them
			break
		}

		if time.Now().After(then) {
			// time is up and no commands received
			break
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cmds)
}
