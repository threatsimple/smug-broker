// broker: pattern routing
// IF a message.matches(some_pattern) { send(message, some_url) }
// parses messages against a regex pattern and if a match occurs,
// forwards the entire message to a given url in a json encoded POST
// if a properly formatted json body is returned, then a message will be
// dispatched appropriately


package smug


import (
    "fmt"
    "regexp"
    "strings"
    "sync"
)


// --------------------------------------------------
// Pattern
// --------------------------------------------------


type Pattern struct {
    re *regexp.Regexp
    url string
    headers map[string]string
    method string
}


func NewExtendedPattern(
        reg string,
        url string,
        headers map[string]string,
        method string,
        ) (*Pattern, error) {
    // validate incoming values
    if len(url) < 10 && ! strings.HasPrefix("http", strings.ToLower(url)) {
        return nil, fmt.Errorf("url must begin with http")
    }
    re := regexp.MustCompile(reg)
    meth := strings.ToUpper(method)
    if meth != "GET" || meth != "POST" {
        return nil, fmt.Errorf("method must be either GET or POST")
    }
    return &Pattern{re:re, url:url, headers:headers, method:method}, nil
}


func NewPattern(reg string, url string) (*Pattern, error) {
    return NewExtendedPattern(
        reg,
        url,
        map[string]string{},
        "POST",
    )
}


func (p *Pattern) parse(ev *Event) {
    fmt.Printf("matches: %+v", p.re.FindAllStringSubmatch(ev.Text,-1))
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


// args [apiurl, prefix]
func (prb *PatternRoutingBroker) Setup(args ...string) { }


// returns (url, tags)
func (prb *PatternRoutingBroker) ParseText(line string) (string, string) {
    /*
    if strings.HasPrefix(line, rtb.prefix) {
        found := xurls.Strict().FindString(line)
        if len(found) > 0 {
            return found, "" // tags empty for now
        }
    }
    */
    return "",""
}


func (prb *PatternRoutingBroker) HandleEvent(ev *Event, dis Dispatcher) {
    prb.pmux.RLock()
    defer prb.pmux.RUnlock()
    for _,ptn := range prb.patterns {
        ptn.parse(ev)
    }

}


func (prb *PatternRoutingBroker) Activate(dis Dispatcher) { }
func (prb *PatternRoutingBroker) Deactivate() { }


