
// broker: pattern routing
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
    "log"
    "net/http"
    "regexp"
    "strings"
)


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
    if re,err := regexp.MustCompile(reg); err != nil {
        return nil, fmt.Errorf("error with regex: %+v", err)
    }
    meth := strings.ToUpper(method)
    if meth != "GET" || meth != "POST" {
        return nil, fmt.Errorf("method must be either GET or POST")
    }
    return &Pattern{reg: re, url:url, headers: headers, method: method}, nil
}


func NewPattern(reg string, url string) (*Pattern, error) {
    return NewExtendedPattern(
        reg,
        url,
        map[string]string{},
        "POST",
    )
}


type PatternRoutingBroker struct {

    patterns []*Pattern
}


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


func (rtb *ReadThisBroker) Name() string {
    return "readthis"
}

// args [apiurl, prefix]
func (rtb *ReadThisBroker) Setup(args ...string) {
    rtb.apiurl = args[0]
    rtb.prefix = args[1]
    rtb.authcode = args[2]
}


// returns (url, tags)
func (rtb *ReadThisBroker) ParseText(line string) (string, string) {
    if strings.HasPrefix(line, rtb.prefix) {
        found := xurls.Strict().FindString(line)
        if len(found) > 0 {
            return found, "" // tags empty for now
        }
    }
    return "",""
}


// since all messages go through the Publish from the Dispatcher we can just
// hook here to look for our read this messages
func (rtb *ReadThisBroker) Publish(ev *Event, dis Dispatcher) {
    found,_ := rtb.ParseText(ev.Text)
    if len(found) > 0 {
        go rtb.Submit(
            ev.Nick,
            found,
            "",
            ev.Text,
        )
    }
}


