package scheduler

import (
	"fmt"
	"time"

	"github.com/gkstretton/asol-protos/go/topics_backend"
	"github.com/gkstretton/dark/services/goo/mqtt"
	"github.com/gkstretton/dark/services/goo/session"
	"github.com/gkstretton/dark/services/goo/twitchapi"
)

const (
	//ml of milk
	milkVolume            = 200
	waterVolume           = 300
	streamPreStartMinutes = 5
)

// time the session will begin UTC
var mainSessionStartTime = RecurringTime{
	day:    time.Sunday,
	hour:   18,
	minute: 30,
	second: 0,
}

// defineSchedule works by launching go routines watching for the specified
// time, to trigger the stated action.
func defineSchedule(sm *session.SessionManager, twitchApi *twitchapi.TwitchApi) {
	go scheduleWatcher(&Schedule{
		name:    "FRIDGE_ON",
		enabled: true,
		function: func() {
			mqtt.Publish(topics_backend.TOPIC_FRIDGE_SWITCH, topics_backend.PAYLOAD_SMART_SWITCH_ON)
			// notify routine operator to fill with milk
			requestFridgeMilk()
		},
		recurringTime: mainSessionStartTime,
		hourOffset:    -10,
		minuteOffset:  0,
	})

	go scheduleWatcher(&Schedule{
		name:    "RUN_SESSION",
		enabled: true,
		function: func() {
			err := RunSession(
				&SessionDescriptor{
					streamPreStartMinutes:  streamPreStartMinutes,
					actorDurationMinutes:   10,
					sessionDurationMinutes: 50,
					runActor:               true,
				},
				sm, twitchApi,
			)
			if err != nil {
				fmt.Println(err)
			}
		},
		recurringTime: mainSessionStartTime,
		minuteOffset:  -streamPreStartMinutes,
	})

	go scheduleWatcher(&Schedule{
		name:    "FRIDGE_OFF",
		enabled: true,
		function: func() {
			mqtt.Publish(topics_backend.TOPIC_FRIDGE_SWITCH, topics_backend.PAYLOAD_SMART_SWITCH_OFF)
		},
		recurringTime: mainSessionStartTime,
		hourOffset:    2, // 2 hours after session start (expect cleaning done by then.)
		minuteOffset:  0,
	})
}
