package smug

import (
	"strings"
	"testing"
)

func TestHelpPattern(t *testing.T) {
	pb := &PatternRoutingBroker{}
	pb.AddPattern(&Pattern{help: "help me"})
	hp := &HelperPattern{pbroker: pb}
	val := hp.Handle(&Event{Text: "nope"}, make(chan *Event, 5))
	if val {
		t.Errorf("should not trigger helper pattern")
	}
	feedback := make(chan *Event, 5)
	val = hp.Handle(&Event{Text: "..list"}, feedback)
	if !val {
		t.Errorf("did not trigger on keyword")
	}
	listev := <-feedback
	if !strings.HasPrefix(listev.Text, "help me") {
		t.Errorf("invalid help returned %s", listev.Text)
	}
}

func TestBasicPatternParse(t *testing.T) {
	p, err := NewPattern(`xyz.+\b`, "http://feh.com")
	if err != nil {
		t.Errorf("test pattern %s", err)
		return
	}
	if m, _ := p.ExtractMatches("asdf xyzhello bye"); len(m) == 0 {
		t.Errorf("err: did not match input")
	}
	if m, _ := p.ExtractMatches("asdf bye"); len(m) != 0 {
		t.Errorf("err: found where not expected not match input")
	}

}

func TestExtendedPatterns(t *testing.T) {
	p, err := NewPattern(`(?P<bob>xyz.+)\b`, "http://feh.com")
	if err != nil {
		t.Errorf("test pattern %s", err)
		return
	}
	_, named := p.ExtractMatches("this is xyz yo")
	if _, ok := named["bob"]; !ok {
		t.Errorf("err: failed to extract named match")
	}
	_, named = p.ExtractMatches("no match here")
	if _, dontwant := named["bob"]; dontwant {
		t.Errorf("err: should not have matched")
	}
}

/*
   testwants := map[string]string {
       "feh":"meh",
   }
   for want,have := range testwants {
       if want != have {
           t.Errorf("err: have [%s] wanted [%s]", have, want)
       }
   }
*/
