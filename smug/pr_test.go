package smug

import (
    "testing"
)


func TestPatternParse(t *testing.T) {
    p,_ := NewPattern(`(?P<xyz>xyz.+\b`, "http://feh.com")
    p.parse(&Event{Text:"asdf xyzhello bye"})
    testwants := map[string]string {
        "feh":"meh",
    }
    for want,have := range testwants {
        if want != have {
            t.Errorf("err: have [%s] wanted [%s]", have, want)
        }
    }


}

