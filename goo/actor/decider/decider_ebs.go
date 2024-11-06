package decider

import (
	"fmt"
	"time"

	"github.com/gkstretton/asol-protos/go/machinepb"
	"github.com/gkstretton/dark/services/goo/actor/executor"
	"github.com/gkstretton/dark/services/goo/ebsinterface"
	"github.com/gkstretton/dark/services/goo/types"
)

type ebsDecider struct {
	endTime time.Time
	ebsApi  ebsinterface.EbsApi
}

func NewEbsDecider(endTime time.Time, ebsApi ebsinterface.EbsApi) Decider {
	return &ebsDecider{
		endTime: endTime,
		ebsApi:  ebsApi,
	}
}

func (d *ebsDecider) decideCollection(predictedState *machinepb.StateReport) *types.CollectionDecision {
	c := d.ebsApi.SubscribeMessages()
	defer d.ebsApi.UnsubscribeMessages(c)

	timeout := time.After(
		time.Until(d.endTime),
	)
	for {
		select {
		case <-timeout:
			l.Printf("timeout in decideCollection, returning nil")
			return nil
		case msg := <-c:
			if msg.Type != types.EbsCollectionRequest {
				l.Printf("cannot make use of message type %s in collection decider", msg.Type)
				continue
			}

			if msg.CollectionRequest == nil {
				l.Println("no CollectionRequest found in message", msg.Type)
				continue
			}

			return &types.CollectionDecision{
				VialNo:  msg.CollectionRequest.Id,
				DropsNo: 3,
			}
		}
	}
}

// decideDispense decides a random location from the unit circle
func (d *ebsDecider) decideDispense(predictedState *machinepb.StateReport) *types.DispenseDecision {
	c := d.ebsApi.SubscribeMessages()
	defer d.ebsApi.UnsubscribeMessages(c)

	// store target coordinates in here
	preemptor := executor.NewDispenseExecutor(&types.DispenseDecision{})

	timeout := time.After(
		time.Until(d.endTime),
	)
	for {
		select {
		case <-timeout:
			l.Printf("timeout in decideDispense, returning nil")
			return nil
		case msg := <-c:
			if msg.Type != types.EbsDispenseRequest {
				l.Printf("cannot make use of message type %s in dispense decider", msg.Type)
				continue
			}

			if msg.DispenseRequest != nil {
				l.Println("got dispense request in dispenseDecider")
				return &types.DispenseDecision{
					X: preemptor.X,
					Y: preemptor.Y,
				}
			}

			if msg.GoToRequest != nil {
				l.Println("got goto request in dispenseDecider")
				preemptor.X = msg.GoToRequest.X
				preemptor.Y = msg.GoToRequest.Y
				preemptor.Preempt()
				continue
			}
		}
	}
}

func (d *ebsDecider) DecideNextAction(predictedState *machinepb.StateReport) (executor.Executor, error) {
	if predictedState.Status <= machinepb.Status_SHUTTING_DOWN {
		l.Println("invalid state for actor, decided nil.")
		return nil, fmt.Errorf("invalid machine status for decision: %s", predictedState.Status)
	}
	if predictedState.PipetteState.Spent {
		// only end after the dispense is done
		if time.Now().After(d.endTime) {
			l.Println("endTime reached on decider, deciding nil.")
			return nil, nil
		}

		l.Println("collection is next, launching decider...")
		decision := d.decideCollection(predictedState)
		return executor.NewCollectionExecutor(decision), nil
	}
	l.Println("dispense is next, launching decider...")
	decision := d.decideDispense(predictedState)
	return executor.NewDispenseExecutor(decision), nil
}