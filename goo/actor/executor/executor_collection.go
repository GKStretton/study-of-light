package executor

import (
	"fmt"
	"time"

	"github.com/gkstretton/asol-protos/go/machinepb"
	"github.com/gkstretton/dark/services/goo/events"
	"github.com/gkstretton/dark/services/goo/types"
)

type collectionExecutor struct {
	vialNo int
	volUl  int
}

func NewCollectionExecutor(d *types.CollectionDecision) *collectionExecutor {
	if d == nil {
		return nil
	}
	vol := d.DropsNo * int(getVialDropVolume(d.VialNo))
	return &collectionExecutor{
		vialNo: d.VialNo,
		volUl:  vol,
	}
}

func (e *collectionExecutor) Preempt() {}

func (e *collectionExecutor) PredictOutcome(state *machinepb.StateReport) *machinepb.StateReport {
	state.CollectionRequest.Completed = true
	state.CollectionRequest.RequestNumber++
	state.CollectionRequest.VialNumber = uint64(e.vialNo)
	state.CollectionRequest.VolumeUl = float32(e.volUl)

	state.PipetteState.VolumeTargetUl = float32(e.volUl)
	state.PipetteState.VialHeld = uint32(e.vialNo)
	state.PipetteState.Spent = false

	return state
}

func (e *collectionExecutor) Execute() {
	collect(e.vialNo, e.volUl)
	// wait for collection to start
	time.Sleep(time.Second * 1)
	<-events.ConditionWaiter(func(sr *machinepb.StateReport) bool {
		return sr.CollectionRequest.Completed
	})
}

func (e *collectionExecutor) String() string {
	return fmt.Sprintf("collectionExecutor (vialNo: %d, volUl: %d)", e.vialNo, e.volUl)
}
