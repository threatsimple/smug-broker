package smug

import (
    "testing"
)

func TestLocalVersionCommand(t *testing.T) {


BROKEN WRITE THIS FOR LOCAL COMMANDS

    lc := &SlackBroker{}
    sb.SetupInternals()
    sb.usercache.users["U6CRHMXK4"] = &SlackUser{
        Id:"U6CRHMXK4",
        Nick:"aaaa",
        Avatar:"",
    }
    sb.usercache.users["U54321"] = &SlackUser{
        Id: "U54321",
        Nick: "boy",
        Avatar:"",
    }
    new_txt := sb.ConvertUserRefs(" <@U6CRHMXK4> congradulations!!!")
    exp_txt := " aaaa congradulations!!!"
    if new_txt !=  exp_txt {
        t.Errorf(
            "bogus slack nick conversion. got:[%s] wanted: [%s]",
            new_txt,
            exp_txt,
        )
    }
    new_txt = sb.ConvertUserRefs("<@U6CRHMXK4> dude <@U54321>")
    exp_txt = "aaaa dude boy"
    if new_txt != exp_txt {
        t.Errorf(
            "bogus slack nick conversion. got:[%s] wanted: [%s]",
            new_txt,
            exp_txt,
        )
    }
}

