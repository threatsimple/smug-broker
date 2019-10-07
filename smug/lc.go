package smug

import (
    "fmt"
    "strings"
    "time"
)


type Command interface {
    exec(oldE *Event, newE *Event, dis Dispatcher)
    help() string
    match(*Event) bool
}


var smugversion string


const Prefix = ".."


/*
 * ********************************************************
 * version command
 * ********************************************************
 */


type VersionCommand struct {
    Version string
}


func (vc *VersionCommand) exec(oldE *Event, newE *Event, dis Dispatcher) {
    msg := fmt.Sprintf("version: %s", vc.Version)
    newE.Text = msg
    newE.RawText = msg
    newE.ts = time.Now()
    dis.Broadcast(newE)
}


func (vc *VersionCommand) help() string {
    return fmt.Sprintf("%sversion - returns version of smug", Prefix)
}


func (vc *VersionCommand) match(ev *Event) bool {
    opstr := fmt.Sprintf("%sversion ", Prefix)
    if strings.HasPrefix(ev.Text, opstr) { return true }
    return false
}


/*
 * ********************************************************
 * ** local cmd broker handles incoming local commands   **
 * ********************************************************
 */

type LocalCmdBroker struct {
    log *Logger
    prefixCmds []Command
    botNick string
    botAvatar string
}


func (lcb *LocalCmdBroker) Name() string {
    return "localcmd"
}


// args [botnick, botavatar, version string]
func (lcb *LocalCmdBroker) Setup(args ...string) {
    lcb.log = NewLogger("locmd")
    if len(args) != 3 {
        lcb.log.Fatal("command broker thrown with too few args")
    }
    lcb.botNick = args[0]
    lcb.botAvatar = args[1]
    lcb.prefixCmds = []Command{
        &VersionCommand{Version:args[2]},
    }
}


func (lcb *LocalCmdBroker) NewEvent() *Event {
    return &Event{
        Origin: lcb,
        Nick: lcb.botNick,
        Avatar: lcb.botAvatar,
        ts: time.Now(),
    }
}


// since all messages go through the Publish from the Dispatcher we can just
// hook here to look for local commands
func (lcb *LocalCmdBroker) Publish(ev *Event, dis Dispatcher) {
    // short circuit if not prefixed by cmd prefix
    // there may come a time when we have embedded commands
    if len(ev.Text) >= len(Prefix) && ev.Text[:len(Prefix)] == Prefix {
        for _,cmd := range lcb.prefixCmds {
            if cmd.match(ev) {
                cmd.exec(ev, lcb.NewEvent(), dis)
                return
            }
        }
    }
}


