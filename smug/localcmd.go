// broker: localcmd
// provides interface into commands which execute local to the smug instance
// instead of being routed elsewhere

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

const opVer = "version"

type VersionCommand struct {
	log     *Logger
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
	return fmt.Sprintf("%s%s - returns version of smug", Prefix, opVer)
}

func (vc *VersionCommand) match(ev *Event) bool {
	opstr := fmt.Sprintf("%s%s", Prefix, opVer)
	vc.log.Debugf("version matching %s to %s", opstr, ev.Text)
	if strings.HasPrefix(ev.Text, opstr) {
		vc.log.Debugf("version found a match")
		return true
	}
	return false
}

/*
 * ********************************************************
 * ** local cmd broker handles incoming local commands   **
 * ********************************************************
 */

type LocalCmdBroker struct {
	log        *Logger
	prefixCmds []Command
	botNick    string
	botAvatar  string
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
		&VersionCommand{Version: args[2], log: lcb.log},
	}
}

func (lcb *LocalCmdBroker) NewEvent(oldEvent *Event) *Event {
	return &Event{
		IsCmdOutput: true,
		Origin:      lcb,
		Actor:       lcb.botNick,
		Avatar:      lcb.botAvatar,
		ts:          time.Now(),
		ReplyBroker: oldEvent.ReplyBroker,
		ReplyTarget: oldEvent.ReplyTarget,
	}
}

func (lcb *LocalCmdBroker) HandleEvent(ev *Event, dis Dispatcher) {
	// short circuit if not prefixed by cmd prefix
	// there may come a time when we have embedded commands
	if len(ev.Text) >= len(Prefix) && ev.Text[:len(Prefix)] == Prefix {
		lcb.log.Debugf("inside Handle, matched")
		for _, cmd := range lcb.prefixCmds {
			if cmd.match(ev) {
				cmd.exec(ev, lcb.NewEvent(ev), dis)
				return
			}
		}
	}
}

func (lcb *LocalCmdBroker) Activate(dis Dispatcher) {}
func (lcb *LocalCmdBroker) Deactivate()             {}
