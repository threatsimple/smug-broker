package smug

import (
    "testing"
)


func TestSimplifyParse(t *testing.T) {
    sb := &SlackBroker{}
    sb.SetupInternals()

    testwants := map[string]string {
        "hey feh": sb.SimplifyParse("hey <http://feh/b|feh>"),
        "hey http://feh/a": sb.SimplifyParse("hey <http://feh/a|>"),
        "hey http://feh/b": sb.SimplifyParse("hey <http://feh/b>"),
    }
    for want, have := range testwants {
        if want != have {
            t.Errorf("err: have [%s] wanted [%s]", have, want)
        }
    }
}


func TestConvertSlackRefs(t *testing.T) {
    sb := &SlackBroker{}
    sb.SetupInternals()

    u1 := &SlackUser{
        Id:"U6CRHMXK4",
        Nick:"aaaa",
        Avatar:"",
    }
    sb.usercache.CacheUser(u1)

    u2 := &SlackUser{
        Id: "U54321",
        Nick: "boy",
        Avatar:"",
    }
    sb.usercache.users[u2.Id] = u2
    sb.usercache.nicks[u2.Nick] = u2

    testwants := map[string]string {
        sb.ConvertRefsToUsers(" <@U6CRHMXK4> congradulations!!!", true):
            " aaaa congradulations!!!",
        sb.ConvertRefsToUsers("<@U6CRHMXK4> dude <@U54321>", true):
            "aaaa dude boy",
        sb.ConvertRefsToUsers("<@U54321> dude <@U54321>", true):
            "boy dude boy",
        sb.ConvertRefsToUsers(" <@88888> congradulations!!!", true):
            " <@88888> congradulations!!!",
        sb.ConvertUsersToRefs("boy: gobble", true):
            "@U54321: gobble",
        sb.ConvertUsersToRefs("nope: hi", true):
            "nope: hi",
        sb.ConvertUsersToRefs("nope: hey @boy", true):
            "nope: hey <@U54321>",
        sb.ConvertUsersToRefs("nope: hey meh@feh.com", true):
            "nope: hey meh@feh.com",
        sb.ConvertUsersToRefs("boy: hey @aaaa", true):
            "@U54321: hey <@U6CRHMXK4>",
        sb.ConvertUsersToRefs("hey @aaaa", true):
            "hey <@U6CRHMXK4>",
        sb.ConvertUsersToRefs("hey @aaaa haha", true):
            "hey <@U6CRHMXK4> haha",
        sb.ConvertUsersToRefs("hey @aaaa and @boy", true):
            "hey <@U6CRHMXK4> and <@U54321>",
        sb.ConvertUsersToRefs("hey @aaaa and @boy happy", true):
            "hey <@U6CRHMXK4> and <@U54321> happy",
        sb.ConvertUsersToRefs("hey @AAAA haha", true):
            "hey <@U6CRHMXK4> haha",
        sb.ConvertUsersToRefs("hey @AAaa haha", true):
            "hey <@U6CRHMXK4> haha",
    }

    for want,have := range testwants {
        if want != have {
            t.Errorf("err: have [%s] wanted [%s]", have, want)
        }
    }


}

