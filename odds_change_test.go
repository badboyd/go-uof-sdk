package uof

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOddsChange(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/odds_change-0.xml")
	assert.Nil(t, err)

	oc := &OddsChange{}
	err = xml.Unmarshal(buf, oc)
	assert.Nil(t, err)

	tests := []struct {
		name string
		f    func(t *testing.T, oc *OddsChange)
	}{
		{"unmarshal", testOddsChangeUnmarshal},
		{"status", testOddsChangeStatus},
		{"urn", testOddsChangeURN},
		{"specifier", testOddsChangeSpecifiers},
		{"marketStatus", testOddsChangeMarketStatus},
	}
	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) { s.f(t, oc) })
	}

	//testu.PP(oc)
}

func testOddsChangeUnmarshal(t *testing.T, oc *OddsChange) {
	assert.Len(t, oc.Markets, 7)
	assert.Equal(t, 123, oc.EventID)
	assert.Equal(t, 2, int(oc.Producer))
	assert.Equal(t, 1234, int(oc.Timestamp))
	assert.Equal(t, 1, *oc.BettingStatus)
	assert.Equal(t, 2, *oc.BetstopReason)

	assert.Equal(t, int64(12345), *oc.Markets[0].NextBetstop)

	// market line calcualtion in unmarshal
	assert.Equal(t, 0, oc.Markets[4].LineID)
	assert.Equal(t, 2701050930, oc.Markets[0].LineID)

	// outcome with 'normal' id
	assert.Equal(t, 1, oc.Markets[3].Outcomes[0].ID)
	assert.Equal(t, 0, oc.Markets[3].Outcomes[0].PlayerID)
	assert.Equal(t, 2, oc.Markets[3].Outcomes[1].ID)
	assert.Equal(t, 0, oc.Markets[3].Outcomes[1].PlayerID)

	// oucome with player id
	assert.Equal(t, 1234, oc.Markets[4].Outcomes[0].ID)
	assert.Equal(t, 1234, oc.Markets[4].Outcomes[0].PlayerID)
	assert.Equal(t, 4322, oc.Markets[4].Outcomes[1].ID)
	assert.Equal(t, 4322, oc.Markets[4].Outcomes[1].PlayerID)
}

func testOddsChangeStatus(t *testing.T, oc *OddsChange) {
	assert.Equal(t, EventStatusLive, oc.EventStatus.Status)
	assert.Equal(t, 7, *oc.EventStatus.MatchStatus)
	assert.Equal(t, 2, *oc.EventStatus.HomeScore)

	mt := *oc.EventStatus.Clock.MatchTime
	assert.Equal(t, ClockTime("75:02"), mt)
	assert.Equal(t, "75:02", mt.String())
	assert.Equal(t, "75", mt.Minute())
}

func testOddsChangeMarketStatus(t *testing.T, oc *OddsChange) {
	m0 := oc.Markets[0]
	m1 := oc.Markets[1]
	m2 := oc.Markets[2]
	m3 := oc.Markets[3]
	m6 := oc.Markets[6]

	assert.Equal(t, MarketStatusActive, m0.Status)
	assert.Equal(t, MarketStatusActive, m1.Status)
	assert.Equal(t, MarketStatusInactive, m2.Status)
	assert.Equal(t, MarketStatusSuspended, m3.Status)
	assert.Equal(t, MarketStatusCancelled, m6.Status)
}

func testOddsChangeURN(t *testing.T, oc *OddsChange) {
	assert.Equal(t, 123, oc.EventURN.ID())
	assert.Equal(t, URNTypeMatch, oc.EventURN.Type())
}

func testOddsChangeSpecifiers(t *testing.T, oc *OddsChange) {
	s := oc.Markets[0].Specifiers
	assert.Equal(t, 1, len(s))
	assert.Equal(t, "41.5", s["score"])

	s = oc.Markets[3].Specifiers
	assert.Equal(t, 4, len(s))
	assert.Equal(t, "2", s["pero"])
}

func TestSpecifiersParsing(t *testing.T) {
	data := []struct {
		specifiers        string
		extendedSpecifers string
		specifiersMap     map[string]string
		variantSpecifier  string
	}{
		{
			specifiers:    "total=1.5|from=1|to=15",
			specifiersMap: map[string]string{"total": "1.5", "from": "1", "to": "15"},
		},
		{
			specifiers:        "total=1.5|from=1",
			extendedSpecifers: "to=15",
			specifiersMap:     map[string]string{"total": "1.5", "from": "1", "to": "15"},
		},
		{
			extendedSpecifers: "to=15",
			specifiersMap:     map[string]string{"to": "15"},
		},
		{
			specifiers:        "from=1",
			extendedSpecifers: "||",
			specifiersMap:     map[string]string{"from": "1"},
		},

		{
			specifiers:       "total=1.5|variant=sr:exact_goals:4+|from=1|to=15",
			specifiersMap:    map[string]string{"total": "1.5", "from": "1", "to": "15", "variant": "sr:exact_goals:4+"},
			variantSpecifier: "sr:exact_goals:4+",
		},
	}
	for i, d := range data {
		s := toSpecifiers(d.specifiers, d.extendedSpecifers)
		assert.Equal(t, len(d.specifiersMap), len(s))
		m := Market{Specifiers: d.specifiersMap}
		assert.Equal(t, d.variantSpecifier, m.VariantSpecifier())
		for k, v := range d.specifiersMap {
			assert.Equal(t, v, s[k], fmt.Sprintf("failed on %d", i))
		}
	}

}
