package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gkstretton/dark/services/goo/contentscheduler"
	"github.com/gkstretton/dark/services/goo/ebsinterface"
	"github.com/gkstretton/dark/services/goo/email"
	"github.com/gkstretton/dark/services/goo/events"
	"github.com/gkstretton/dark/services/goo/filesystem"
	"github.com/gkstretton/dark/services/goo/keyvalue"
	"github.com/gkstretton/dark/services/goo/livecapture"
	"github.com/gkstretton/dark/services/goo/mqtt"
	"github.com/gkstretton/dark/services/goo/obs"
	"github.com/gkstretton/dark/services/goo/scheduler"
	"github.com/gkstretton/dark/services/goo/session"
	"github.com/gkstretton/dark/services/goo/twitchapi"
	"github.com/gkstretton/dark/services/goo/types"
	"github.com/gkstretton/dark/services/goo/vialprofiles"
)

var (
	test = flag.Bool("test", false, "if true, just run test code")
)

func testEBS() {
	es, err := ebsinterface.NewExtensionSession(time.Hour * 2)
	if err != nil {
		panic(err)
	}
	es.UpdateCurrentVoteStatus(&types.VoteStatus{
		VoteType: types.VoteTypeCollection,
		CollectionVoteStatus: &types.CollectionVoteStatus{
			TotalVotes: 5,
			VoteCounts: map[int]int{5: 25},
		},
	})
	v := es.SubscribeVotes()
	defer es.UnsubscribeVotes(v)

	for {
		vote := <-v
		fmt.Printf("got vote:\n%+v\n", vote)
	}
}

func main() {
	flag.Parse()

	if *test {
		testEBS()
		return
	}

	filesystem.AssertBasePaths()

	mqtt.Start()
	keyvalue.Start()
	email.Start()

	sm := session.NewSessionManager(false)
	twitchApi := twitchapi.Start()

	events.Start(sm)
	livecapture.Start(sm)
	obs.Start(sm)
	vialprofiles.Start(sm)
	scheduler.Start(sm, twitchApi)
	contentscheduler.Start(sm)

	// Block to prevent early quit
	fmt.Println("finished init, main loop sleeping.")
	for {
		time.Sleep(time.Millisecond * time.Duration(100))
	}
}