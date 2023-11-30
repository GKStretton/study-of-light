package ebsinterface

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gkstretton/dark/services/goo/keyvalue"
	"github.com/gkstretton/dark/services/goo/types"
)

// consists of latest state report and current vote status
type broadcastData struct {
	CurrentVoteStatus  *types.VoteStatus
	PreviousVoteResult *types.VoteStatus
	RobotStatus        *robotStatus
}

// ManualTriggerBroadcast sends early, used after a significant update to
// improve responsiveness
func (e *ExtensionSession) ManualTriggerBroadcast() {
	if e.cleanUpDone {
		return
	}
	l.Println("manually triggering broadcast...")
	e.triggerBroadcast <- struct{}{}
}

func (e *ExtensionSession) UpdatePreviousVoteResult(data *types.VoteStatus) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.broadcastDataCache.PreviousVoteResult = data
}

func (e *ExtensionSession) UpdateCurrentVoteStatus(data *types.VoteStatus) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.broadcastDataCache.CurrentVoteStatus = data
}

func (e *ExtensionSession) updateRobotStatus(data *robotStatus) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.broadcastDataCache.RobotStatus = data
}

// broadcasts the BroadcastData cache once per second
func (e *ExtensionSession) regularBroadcast() {
	// get marshaled data, protected by lock
	d := func() ([]byte, error) {
		e.lock.Lock()
		defer e.lock.Unlock()

		jsonData, err := json.Marshal(e.broadcastDataCache)
		if err != nil {
			return nil, err
		}
		return jsonData, nil
	}

	send := func() {
		data, err := d()
		if err != nil {
			l.Printf("failed to marshal broadcast data: %v\n", err)
			return
		}
		err = e.broadcastData(data)
		if err != nil {
			l.Printf("failed to send broadcast data: %v\n", err)
			return
		}
		l.Println("sent broadcast")
	}

	next := time.After(0)
	for {
		select {
		case <-e.exitCh:
			l.Println("exiting regularBroadcast loop")
			return
		case <-e.triggerBroadcast:
			next = time.After(time.Second * 2)
			send()
		case <-next:
			next = time.After(time.Second)
			send()
		}
	}
}

// broadcastData must be called with rate limiting due to pubsub api limit
// this is officially stated as 100 per minute, but there's a thread saying it's
// 1 regen per second with pool of 100.
// https://github.com/twitchdev/issues/issues/612
// So we stick to 1 per second, 60 per minute.
func (e *ExtensionSession) broadcastData(jsonData []byte) error {
	type payload struct {
		message        string
		broadcaster_id string
		target         []string
	}
	pl := &payload{
		message:        string(jsonData),
		broadcaster_id: channelId,
		target:         []string{"broadcast"},
	}

	jsonPl, err := json.Marshal(pl)
	if err != nil {
		return err
	}
	url := "https://api.twitch.tv/helix/extensions/pubsub"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPl))
	if err != nil {
		return err
	}

	clientID := string(keyvalue.Get("TWITCH_EXTENSION_CLIENT_ID"))

	req.Header.Set("Authorization", "Bearer "+e.broadcastToken)
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}