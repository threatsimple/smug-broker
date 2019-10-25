package smug


import (
    "testing"
)


func TestPatternParse(t *testing.T) {
    p,err := NewPattern(`(?P<xyz>xyz.+)\b`, "http://feh.com")
    if err != nil {
        t.Errorf("test pattern %s", err)
        return
    }
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


