package scheduler

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gkstretton/asol-protos/go/machinepb"
	"github.com/gkstretton/asol-protos/go/topics_backend"
	"github.com/gkstretton/asol-protos/go/topics_firmware"
	"github.com/gkstretton/dark/services/goo/actor"
	"github.com/gkstretton/dark/services/goo/actor/executor"
	"github.com/gkstretton/dark/services/goo/events"
	"github.com/gkstretton/dark/services/goo/mqtt"
	"github.com/gkstretton/dark/services/goo/session"
	"github.com/gkstretton/dark/services/goo/twitchapi"
)

var sl = log.New(os.Stdout, "[session scheduler] ", log.Flags())

type SessionDescriptor struct {
	// how many minutes before session should stream start?
	streamPreStartMinutes  int
	actorDurationMinutes   int
	sessionDurationMinutes int
	runActor               bool
}

var lock *AutomationLock = &AutomationLock{}

func registerHandlers(sm *session.SessionManager, twitchApi *twitchapi.TwitchApi) {
	mqtt.Subscribe("asol/debug/runStartSequence", func(topic string, payload []byte) {
		go func() {
			fmt.Println(runStartSequence(0, false))
		}()
	})
	mqtt.Subscribe("asol/debug/runEndSequence", func(topic string, payload []byte) {
		go func() {
			fmt.Println(runEndSequence())
		}()
	})
	mqtt.Subscribe(topics_backend.TOPIC_RUN_START_SEQUENCE, func(topic string, payload []byte) {
		go func() {
			fmt.Println(runStartSequence(0, true))
		}()
	})
	mqtt.Subscribe(topics_backend.TOPIC_RUN_END_SEQUENCE, func(topic string, payload []byte) {
		go func() {
			fmt.Println(runEndSequence())
		}()
	})

	mqtt.Subscribe(topics_backend.TOPIC_RUN_FULL_SESSION, func(topic string, payload []byte) {
		go func() {
			err := RunSession(
				&SessionDescriptor{
					streamPreStartMinutes:  0,
					actorDurationMinutes:   10,
					sessionDurationMinutes: 50,
					runActor:               true,
				},
				sm, twitchApi,
			)
			if err != nil {
				fmt.Println(err)
			}
		}()
	})

	mqtt.Subscribe(topics_backend.TOPIC_RUN_MANUAL_SESSION, func(topic string, payload []byte) {
		go func() {
			err := RunSession(
				&SessionDescriptor{
					streamPreStartMinutes:  0,
					actorDurationMinutes:   10,
					sessionDurationMinutes: 50,
					runActor:               false,
				},
				sm, twitchApi,
			)
			if err != nil {
				fmt.Println(err)
			}
		}()
	})
}

func RunSession(
	d *SessionDescriptor,
	sm *session.SessionManager,
	twitchApi *twitchapi.TwitchApi,
) error {
	if lock.Get() {
		return fmt.Errorf("automation already running")
	}
	lock.Set(true)
	defer lock.Set(false)

	sl.Printf("running session with descriptor: %+v\n", d)

	beginTime := time.Now()

	err := runStartSequence(d.streamPreStartMinutes, true)
	if err != nil {
		return err
	}

	if d.runActor {
		seed := rand.Int63()
		err = sm.SetCurrentSessionSeed(seed)
		if err != nil {
			sl.Printf("failed to set seed: %v\n", err)
		}

		sl.Println("launching actor")
		err = actor.LaunchActor(twitchApi, time.Duration(d.actorDurationMinutes)*time.Minute, seed)
		if err != nil {
			sl.Println("actor error, erroring")
			mqtt.Publish(topics_backend.TOPIC_SESSION_PAUSE, "")
			// email for help
			errWrap := fmt.Errorf("actor returned error, unknown situation: %s", err)
			requestSessionIntervention(errWrap)
			// add timeout to end session after a few hours if no human response
			go sessionTimeout(time.Hour*3, true)
			return err
		}
		sl.Println("actor success")
		mqtt.Publish(topics_firmware.TOPIC_GOTO_RING_IDLE_POS, "")
	} else {
		sl.Println("ready for manual control...")
	}

	endTime := beginTime.Add(time.Duration(d.sessionDurationMinutes+d.streamPreStartMinutes) * time.Minute)

	waitForTOffset(endTime, -3, 0)

	err = runEndSequence()
	if err != nil {
		return err
	}

	return nil
}

func runStartSequence(streamPreStartMinutes int, realSession bool) error {
	sl.Println("running start sequence...")

	ch := events.Subscribe()
	defer events.Unsubscribe(ch)

	sl.Println("starting stream, and waiting...")
	if realSession {
		mqtt.Publish(topics_backend.TOPIC_STREAM_START, "")
	}
	time.Sleep(time.Duration(streamPreStartMinutes)*time.Minute + time.Second*10)

	// start time
	sl.Println("starting session")
	if realSession {
		mqtt.Publish(topics_backend.TOPIC_SESSION_BEGIN, "PRODUCTION")
	}

	sl.Println("requesting wake")
	mqtt.Publish(topics_firmware.TOPIC_WAKE, "")

	time.Sleep(time.Second * 15)
	// if not awake, abort and error
	sl.Println("checking for awake status")
	sr := events.GetLatestStateReportCopy()
	if sr.Status != machinepb.Status_IDLE_MOVING && sr.Status != machinepb.Status_IDLE_STATIONARY {
		sl.Println("invalid status, erroring")
		mqtt.Publish(topics_backend.TOPIC_SESSION_PAUSE, "")
		err := fmt.Errorf("status was %s after waking but expected idle. Pausing and aborting automation", sr.Status)
		// email for help
		requestSessionIntervention(err)
		// add timeout to end session after a few hours if no human response
		go sessionTimeout(time.Hour*3, false)
		return err
	}
	sl.Println("valid status")

	sl.Println("dispensing milk")
	mqtt.Publish(
		topics_firmware.TOPIC_FLUID,
		fmt.Sprintf(
			"%d,%d,%t",
			machinepb.FluidType_FLUID_MILK,
			milkVolume,
			false, // open drain
		),
	)

	time.Sleep(time.Second * 5)
	sl.Println("start sequence done")

	return nil
}

func runEndSequence() error {
	ch := events.Subscribe()
	defer events.Unsubscribe(ch)

	mqtt.Publish(topics_firmware.TOPIC_GOTO_RING_IDLE_POS, "")

	sl.Println("running end sequence")
	sl.Println("draining milk")

	// drain
	mqtt.Publish(
		topics_firmware.TOPIC_FLUID,
		fmt.Sprintf(
			"%d,%d,%t",
			machinepb.FluidType_FLUID_DRAIN,
			milkVolume,
			false, // additional open drain false
		),
	)

	waitForFluidComplete(ch)
	sl.Println("dispensing water")

	mqtt.Publish(
		topics_firmware.TOPIC_FLUID,
		fmt.Sprintf(
			"%d,%d,%t",
			machinepb.FluidType_FLUID_WATER,
			waterVolume,
			false,
		),
	)

	waitForFluidComplete(ch)

	sl.Println("draining water")
	mqtt.Publish(
		topics_firmware.TOPIC_FLUID,
		fmt.Sprintf(
			"%d,%d,%t",
			machinepb.FluidType_FLUID_DRAIN,
			waterVolume,
			false,
		),
	)

	time.Sleep(42 * time.Second)

	sl.Println("shutting down")
	mqtt.Publish(topics_firmware.TOPIC_SHUTDOWN, "")

	waitForSleeping(ch)

	time.Sleep(2 * time.Second)

	sl.Println("ending session")
	mqtt.Publish(topics_backend.TOPIC_SESSION_END, "")

	time.Sleep(2 * time.Minute)

	sl.Println("ending stream")
	mqtt.Publish(topics_backend.TOPIC_STREAM_END, "")

	sl.Println("end sequence done")

	return nil
}

func waitForTOffset(t time.Time, minutesOffset, secondsOffset int) {
	<-time.After(
		time.Until(
			t.
				Add(time.Minute * time.Duration(minutesOffset)).
				Add(time.Second * time.Duration(secondsOffset)),
		),
	)
}

func waitForSleeping(ch chan *machinepb.StateReport) {
	w := executor.ConditionWaiter(ch, func(sr *machinepb.StateReport) bool {
		return sr.Status == machinepb.Status_SLEEPING
	})
	<-w
}

func waitForFluidComplete(ch chan *machinepb.StateReport) {
	time.Sleep(2 * time.Second)
	w := executor.ConditionWaiter(ch, func(sr *machinepb.StateReport) bool {
		return sr.FluidRequest.Complete == true
	})
	<-w
}
