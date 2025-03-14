package actor

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gkstretton/asol-protos/go/machinepb"
	"github.com/gkstretton/dark/services/goo/actor/decider"
	"github.com/gkstretton/dark/services/goo/actor/executor"
	"github.com/gkstretton/dark/services/goo/ebsinterface"
	"github.com/gkstretton/dark/services/goo/events"
	"github.com/gkstretton/dark/services/goo/session"
	"github.com/gkstretton/dark/services/goo/twitchapi"
)

var l = log.New(os.Stdout, "[actor] ", log.Flags())
var waitForUser = flag.Bool("waitForUser", false, "if true, do blocking waits at certain debug points")

type decision struct {
	e   executor.Executor
	err error
}

var lock *ActorLock = &ActorLock{}

var exitCh chan struct{} = make(chan struct{}, 1)

func shouldExit() bool {
	select {
	case <-exitCh:
		return true
	default:
		return false
	}
}

func Setup(sm *session.SessionManager, ebsApi ebsinterface.EbsApi) {
	subscribeToBrokerTopics(sm, ebsApi)
}

// LaunchActor is launched to control a session after the canvas is prepared.
// It should effect art.
func LaunchActor(twitchApi *twitchapi.TwitchApi, ebsApi ebsinterface.EbsApi, actorTimeout time.Duration, seed int64, testing bool) error {
	if lock.Get() {
		return fmt.Errorf("actor already running")
	}
	lock.Set(ebsApi, true)
	defer lock.Set(ebsApi, false)

	// clear exit flag
	_ = shouldExit()

	l.Printf("Launching actor with seed: %d; testing: %t\n", seed, testing)

	endTime := time.Now().Add(actorTimeout)
	var d decider.Decider
	if ebsApi == nil {
		l.Println("using auto decider")
		d = decider.NewAutoDecider(endTime, seed, testing)
	} else {
		l.Println("using ebs decider")
		var fallback = decider.NewAutoDecider(endTime, seed, testing)
		d = decider.NewEbsDecider(endTime, ebsApi, fallback)
	}

	awaitDecision := decide(d, events.GetLatestStateReportCopy())
	decision := <-awaitDecision

	for {
		if shouldExit() {
			l.Printf("exit triggered")
			break
		}
		if decision.err != nil {
			l.Printf("error in decider, exiting actor: %s\n", decision.err)
			return decision.err
		}
		if decision.e == nil {
			l.Println("saw nil decision, exiting actor")
			break
		}
		awaitCompletion, predictedCompletionState := executor.RunExecutorNonBlocking(decision.e)

		// get next action while the action is being performed
		awaitDecision = decide(d, predictedCompletionState)

		<-awaitCompletion // ensure last action finished
		decision = <-awaitDecision
	}

	return nil
}

func decide(decider decider.Decider, predictedState *machinepb.StateReport) chan decision {
	c := make(chan decision)

	go func() {
		if predictedState == nil {
			c <- decision{
				e:   nil,
				err: fmt.Errorf("predictedState nil"),
			}
			close(c)
			return
		}

		l.Printf("making next decision...\n")
		e, err := decider.DecideNextAction(predictedState)
		if *waitForUser {
			fmt.Scanln()
		}
		if err != nil {
			l.Printf("error making next decision: %v\n", err)
		} else {
			l.Printf("made next decision: %v\n", e)
		}

		c <- decision{
			e:   e,
			err: err,
		}
		close(c)
	}()
	return c
}
