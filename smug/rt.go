package smug

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "strings"

    "github.com/mvdan/xurls"
)


type ReadThisBroker struct {
    apiurl string
    prefix string
    authcode string
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
func (rtb *ReadThisBroker) Publish(ev *Event) {
    found,_ := rtb.ParseText(ev.Text)
    if len(found) > 0 {
        rtb.Submit(
            ev.Nick,
            found,
            "",
            ev.Text,
        )
    }
}


