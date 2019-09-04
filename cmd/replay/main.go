package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/minus5/svckit/file"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
	"github.com/minus5/uof/queue"
)

const (
	EnvBookmakerID = "UOF_BOOKMAKER_ID"
	EnvToken       = "UOF_TOKEN"
)

func env(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Errorf("env %s not found", name)
	}
	return val
}

var (
	scenarioID   int
	eventID      int
	sample       bool
	token        string
	speed        int
	maxDelay     int
	outputFolder string
)

func init() {
	var showSampleEvents bool
	flag.IntVar(&speed, "speed", 100, "replay speed, speed times faster than in reality")
	flag.IntVar(&maxDelay, "max-delay", 10, "maximum delay between messages in milliseconds (this is helpful especially in pre-match odds where delay can be even a few hours or more)")
	flag.IntVar(&scenarioID, "scenario", 0, "scenario (1,2 or 3) to replay")
	flag.IntVar(&eventID, "event", 0, "event to replay")
	flag.BoolVar(&sample, "sample", false, "replay sample events")
	flag.BoolVar(&showSampleEvents, "show", false, "show interesting sample events and exit")
	flag.StringVar(&outputFolder, "out", "./tmp", "output fodler location")
	flag.Parse()

	if showSampleEvents {
		for _, e := range sampleEvents() {
			fmt.Printf("%-27s %s\n", e.URN, e.Description)
		}
		os.Exit(0)
	}

}

func main() {
	ctx := signal.InteruptContext()
	token := env(EnvToken)
	conn, err := queue.DialReplay(ctx, env(EnvBookmakerID), token)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("connected")

	rpl := api.Replay(token)
	if eventID > 0 {
		eventURN := fmt.Sprintf("sr:match:%d", eventID)
		must(rpl.StartEvent(eventURN, speed, maxDelay))
	}
	if scenarioID > 0 {
		must(rpl.StartScenario(scenarioID, speed, maxDelay))
	}
	if sample {
		must(rpl.Reset())
		for _, s := range sampleEvents() {
			must(rpl.Add(s.URN))
		}
		must(rpl.Play(speed, maxDelay))
	}

	done(saveMsgs(conn.Listen()))

}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func done(in <-chan uof.QueueMsg) {
	for _ = range in {
		fmt.Print(".")
	}
}

func saveMsgs(in <-chan uof.QueueMsg) <-chan uof.QueueMsg {
	out := make(chan uof.QueueMsg, 128)
	go func() {
		defer close(out)
		for m := range in {
			out <- m
			saveMsg(m)
		}
	}()
	return out
}

func saveMsg(m uof.QueueMsg) {
	fn := fmt.Sprintf("%s/%011d-%s", outputFolder, m.Timestamp, m.RoutingKey)
	if err := file.Save(fn, m.Body); err != nil {
		log.Fatal(err)
	}
}

type sampleEvent struct {
	Description string
	URN         string
}

func sampleEvents() []sampleEvent {
	// stolen from the ExampleReplayEvents.cs in the C# SDK
	return []sampleEvent{
		sampleEvent{"Soccer Match - English Premier League 2017 (Watford vs Westham)", "sr:match:11830662"},
		sampleEvent{"Soccer Match w Overtime - Primavera Cup", "sr:match:12865222"},
		sampleEvent{
			"Soccer Match w Overtime & Penalty Shootout - KNVB beker 17/18 - FC Twente Enschede vs Ajax Amsterdam",
			"sr:match:12873164"},
		sampleEvent{"Soccer Match with Rollback Betsettlement from Prematch Producer", "sr:match:11958226"},
		sampleEvent{
			"Soccer Match aborted mid-game - new match played later (first match considered cancelled according to betting rules}",
			"sr:match:11971876"},
		sampleEvent{"Soccer Match w PlayerProps {prematch odds only}", "sr:match:12055466"},
		sampleEvent{"Tennis Match - ATP Paris Final 2017", "sr:match:12927908"},
		sampleEvent{"Tennis Match where one of the players retired", "sr:match:12675240"},
		sampleEvent{"Tennis Match with bet_cancel adjustments using rollback_bet_cancel", "sr:match:13616027"},
		sampleEvent{
			"Tennis Match w voided markets due to temporary loss of coverage - no ability to verify results",
			"sr:match:13600533"},
		sampleEvent{"Basketball Match - NBA Final 2017 - {Golden State Warriors vs Cleveland Cavaliers}",
			"sr:match:11733773"},
		sampleEvent{"Basketball Match w voided DrawNoBet {2nd half draw}", "sr:match:12953638"},
		sampleEvent{"Basketball Match w PlayerProps", "sr:match:12233896"},
		sampleEvent{"Icehockey Match - NHL Final 2017 {6th match - {Nashville Predators vs Pittsburg Penguins}",
			"sr:match:11784628"},
		sampleEvent{"Icehockey Match with Rollback BetCancel", "sr:match:11878140"},
		sampleEvent{"Icehockey Match with overtime + rollback_bet_cancel + match_status=\"aet\"",
			"sr:match:11878386"},
		sampleEvent{"American Football Game - NFL 2018/2018 {Chicago Bears vs Atlanta Falcons}",
			"sr:match:11538563"},
		sampleEvent{"American Football Game w PlayerProps", "sr:match:13552497"},
		sampleEvent{"Handball Match - DHB Pokal 17/18 {SG Flensburg-Handewitt vs Fuchse Berlin}",
			"sr:match:12362564"},
		sampleEvent{"Baseball Game - MLB 2017 {Final Los Angeles Dodgers vs Houston Astros}",
			"sr:match:12906380"},
		sampleEvent{"Badminton Game - Indonesia Masters 2018", "sr:match:13600687"},
		sampleEvent{"Snooker - International Championship 2017 {Final Best-of-19 frames}", "sr:match:12927314"},
		sampleEvent{"Darts - PDC World Championship 17/18 - {Final}", "sr:match:13451765"},
		sampleEvent{"CS:GO {ESL Pro League 2018}", "sr:match:13497893"},
		sampleEvent{"Dota2 {The International 2017 - Final}", "sr:match:12209528"},
		sampleEvent{"League of Legends Match {LCK Spring 2018}", "sr:match:13516251"},
		sampleEvent{"Cricket Match [Premium Cricket] - The Ashes 2017 {Australia vs England}",
			"sr:match:11836360"},
		sampleEvent{
			"Cricket Match {rain affected} [Premium Cricket] - ODI Series New Zealand vs. Pakistan 2018",
			"sr:match:13073610"},
		sampleEvent{"Volleyball Match {includes bet_cancels}", "sr:match:12716714"},
		sampleEvent{"Volleyball match where Betradar loses coverage mid-match - no ability to verify results",
			"sr:match:13582831"},
		sampleEvent{"Aussie Rules Match {AFL 2017 Final}", "sr:match:12587650"},
		sampleEvent{"Table Tennis Match {World Cup 2017 Final", "sr:match:12820410"},
		sampleEvent{"Squash Match {Qatar Classic 2017}", "sr:match:12841530"},
		sampleEvent{"Beach Volleyball", "sr:match:13682571"},
		sampleEvent{"Badminton", "sr:match:13600687"},
		sampleEvent{"Bowls", "sr:match:13530237"},
		sampleEvent{"Rugby League", "sr:match:12979908"},
		sampleEvent{"Rugby Union", "sr:match:12420636"},
		sampleEvent{"Rugby Union 7s", "sr:match:13673067"},
		sampleEvent{"Handball", "sr:match:12362564"},
		sampleEvent{"Futsal", "sr:match:12363102"},
		sampleEvent{"Golf Winner Events + Three Balls - South African Open {Winner events + Three balls}",
			"sr:simple_tournament:66820"},
		sampleEvent{"Season Outrights {Long-term Outrights} - NFL 2017/18", "sr:season:40175"},
		sampleEvent{"Race Outrights {Short-term Outrights} - Cycling Tour Down Under 2018", "sr:stage:329361"},
	}
}
