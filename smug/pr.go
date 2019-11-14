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
    "time"
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


func (p *Pattern) Handle(ev *Event, feedback chan *Event) bool {
    matches, named := p.ExtractMatches(ev.Text)
    if len(matches) == 0 {
        return false
    }
    go p.Submit(ev, ev.Actor, ev.Text, named, feedback)
    return true
}


type JsonBlock struct {
    Text string `json:text`
    Img string  `json:img`
    Title string `json:title`
}


type JsonResponse struct {
    Text string `json:text`
    Blocks []JsonBlock `json:blocks`
}


func (p *Pattern) Submit(
        originEvt *Event,
        actor string,
        text string,
        named NamedGroups,
        feedback chan *Event,
        ) {
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
        return
    }
    req,err := http.NewRequest(p.method, p.url, bytes.NewBuffer(reqbody))
    req.Header.Set("Content-Type", "application/json")
    for h,v := range p.headers {
        req.Header.Set(h, v)
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf(
            "ERR readthis post failed to %s body=%s %+v\n",
            p.url, reqbody, err)
        return
    }
    defer resp.Body.Close()
    body,_ := ioutil.ReadAll(resp.Body)
    if err != nil || ! strings.HasPrefix(resp.Status, "200") {
        fmt.Printf("ERR resp  %s %+v %s\n", err, resp.Status, string(body))
        return
    }
    // now attempt to see if anything returned
    if len(string(body)) > 0 {
        var dat JsonResponse
        fmt.Printf("JSON BODY IS\n%s", string(body))
        if err = json.Unmarshal(body, &dat); err != nil {
            // just abadon hope here
            fmt.Printf("ERR WITH JSON UNMARSHAL got body of %s", string(body))
            return
        }
        text := dat.Text
        blocks := []*EventBlock{}
        for _,blk := range dat.Blocks {
            blocks = append(blocks,
                &EventBlock{Title: blk.Title, Text:blk.Text, ImgUrl:blk.Img},
            )
        }
        fmt.Printf("dat: %+v \n\n", dat)
        feedback <- &Event{
            IsCmdOutput: true,
            Origin: nil, // PRB will set this
            ReplyBroker: originEvt.ReplyBroker,
            ReplyTarget: originEvt.ReplyTarget,
            Actor: "",
            Text: text,
            ContentBlocks: blocks,
            ts: time.Now(),
        }
    }
}


// --------------------------------------------------
// PatternRoutingBroker
// --------------------------------------------------


type PatternRoutingBroker struct {
    log *Logger
    pmux sync.RWMutex
    feedback chan *Event
    patterns []*Pattern
}


func (prb *PatternRoutingBroker) AddPattern(newp *Pattern) {
    prb.pmux.Lock()
    prb.patterns = append(prb.patterns, newp)
    prb.pmux.Unlock()
}


func (prb *PatternRoutingBroker) Name() string {
    return "pattern-router"
}


// args [regex,apiurl,method,headers]
func (prb *PatternRoutingBroker) Setup(args ...string) {
    prb.log = NewLogger(prb.Name())
    prb.feedback = make(chan *Event, 100)
}


func (prb *PatternRoutingBroker) HandleEvent(ev *Event, dis Dispatcher) {
    prb.pmux.RLock()
    defer prb.pmux.RUnlock()
    for _,ptn := range prb.patterns {
        if ptn.Handle(ev, prb.feedback) {
            break
        }
    }
}


func (prb *PatternRoutingBroker) Activate(dis Dispatcher) {
    for {
        ev := <-(prb.feedback)
        ev.Origin = prb
        fmt.Printf("recvd from Channel: %+v", ev)
        dis.Broadcast(ev)
    }
}

func (prb *PatternRoutingBroker) Deactivate() { }


