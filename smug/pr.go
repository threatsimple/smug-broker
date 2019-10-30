// broker: pattern routing
// IF a message.matches(some_pattern) { send(message, some_url) }
// parses messages against a regex pattern and if a match occurs,
// forwards the entire message to a given url in a json encoded POST
// if a properly formatted json body is returned, then a message will be
// dispatched appropriately


package smug


import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "regexp"
    "strings"
    "sync"
)


// --------------------------------------------------
// Pattern
// --------------------------------------------------


type Pattern struct {
    name string
    re *regexp.Regexp
    url string
    headers map[string]string
    vars map[string]string
    method string
}


// for our group matches
type NamedGroups map[string]string


func NewExtendedPattern(
        name string,
        reg string,
        url string,
        headers map[string]string,
        vars map[string]string,
        method string,
        ) (*Pattern, error) {
    // validate incoming values a smidge
    if len(url) < 10 && ! strings.HasPrefix("http", strings.ToLower(url)) {
        return nil, fmt.Errorf("url must begin with http")
    }
    re,err := regexp.Compile(reg)
    if err != nil {
        return nil, fmt.Errorf("error compiling regex: %s", err)
    }
    meth := strings.ToUpper(method)
    if ! (meth == "GET" || meth == "POST") {
        return nil, fmt.Errorf("method must be either GET or POST")
    }
    return &Pattern{
        name:name,
        re:re,
        url:url,
        headers:headers,
        method:method,
    }, nil
}


func NewPattern(reg string, url string) (*Pattern, error) {
    return NewExtendedPattern(
        "n/a",
        reg,
        url,
        map[string]string{},
        map[string]string{},
        "POST",
    )
}


func (p *Pattern) ExtractMatches(text string) ([]string, NamedGroups) {
    matches := p.re.FindStringSubmatch(text)
    named := make(NamedGroups)
    if len(matches) == 0 {
        return matches,named
    }
    for i, name := range p.re.SubexpNames() {
        if i != 0 && name != "" {
            named[name] = matches[i]
        }
    }
    return matches,named
}


func (p *Pattern) Handle(ev *Event) bool {
    matches, named := p.ExtractMatches(ev.Text)
    if len(matches) == 0 {
        return false
    }
    go p.Submit(ev.Actor, ev.Text, named)
    return true
}


func (p *Pattern) Submit(actor string, text string, named NamedGroups) error {
    payload := map[string]string{
        "actor": actor,
        "text": text,
    }
    for k,v := range named {
        payload[k] = v
    }
    for k,v := range p.vars {
        payload[k] = v
    }
    reqbody, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("COULD NOT ENCODE %+v", err)
    }
    req,err := http.NewRequest(p.method, p.url, bytes.NewBuffer(reqbody))
    req.Header.Set("Content-Type", "application/json")
    for h,v := range p.headers {
        req.Header.Set(h, v)
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf(
            "ERR readthis post failed to %s body=%s %+v",
            p.url, reqbody, err)
    }
    defer resp.Body.Close()
    if err != nil || resp.Status != "200" {
        body,_ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("ERR resp %+v %s", err, string(body))
        // XXX how to get returned text to broker dispatch??
        // use a channel?
    }
    return nil
}


// --------------------------------------------------
// PatternRoutingBroker
// --------------------------------------------------


type PatternRoutingBroker struct {
    pmux sync.RWMutex
    patterns []*Pattern
}


func (prb *PatternRoutingBroker) AddPattern(newp *Pattern) {
    prb.pmux.Lock()
    prb.patterns = append(prb.patterns, newp)
    prb.pmux.Unlock()
}


/*
func (rtb *ReadThisBroker) Submit(
        nick string, url string, tags string, text string ) {
    reqbody, err := json.Marshal(map[string]string{
        "nick":nick,
        "url":url,
        "text":text,
    })
    if err != nil {
        log.Printf("COULD NOT ENCODE %+v", err)
        return
    }
    req,err := http.NewRequest("POST", rtb.apiurl, bytes.NewBuffer(reqbody))
    req.Header.Set("X-ReadThis-Auth", rtb.authcode)
    req.Header.Set("Content-Type", "application/json")
    log.Printf("readthis posting to %s body=%s", rtb.apiurl, reqbody)
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("ERR readthis post failed to %s body=%s %+v",
            rtb.apiurl, reqbody, err)
        return
    }
    defer resp.Body.Close()
    if err != nil || resp.Status != "200" {
        body,_ := ioutil.ReadAll(resp.Body)
        log.Printf("ERR resp %+v %s", err, string(body))
        return
    }
}
*/


func (prb *PatternRoutingBroker) Name() string {
    return "pattern-router"
}


// args [regex,apiurl,method,headers]
func (prb *PatternRoutingBroker) Setup(args ...string) { }


func (prb *PatternRoutingBroker) HandleEvent(ev *Event, dis Dispatcher) {
    prb.pmux.RLock()
    defer prb.pmux.RUnlock()
    for _,ptn := range prb.patterns {
        ptn.Handle(ev)
    }
}


func (prb *PatternRoutingBroker) Activate(dis Dispatcher) { }

func (prb *PatternRoutingBroker) Deactivate() { }


