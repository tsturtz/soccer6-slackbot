package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jasonlvhit/gocron"
	"github.com/nlopes/slack"
)

// run daily at 20:00 UTC and then check if it's saturday
func runSaturdayReminderCron(post *slack.Client, channel string) {
	// gocron.Every(3).Seconds().Do(saturdayReminderCron, post, channel) // (dev) TODO: update this to check env.
	gocron.Every(1).Day().At("20:00").Do(saturdayReminderCron, post, channel) // (prod)
	<-gocron.Start()
}

func saturdayReminderCron(post *slack.Client, channel string) {
	rn := time.Now().UTC()
	endOfSeason := time.Date(2018, 3, 12, 0, 0, 0, 0, time.UTC) // march 11, 2018 is the last game
	// only send if season still running
	if endOfSeason.After(rn) {
		// check if it is saturday (library's day of week support is broken)
		switch time.Now().Weekday() {
		case time.Saturday:
			post.PostMessage(channel, ">>>Hey @channel! It's Saturday - that means there's probably a game tomorrow! Type `@soccerbot next` to find out.", slack.PostMessageParameters{
				Username:    "soccerbot",
				User:        "soccerbot",
				AsUser:      false,
				Parse:       "",
				LinkNames:   1,
				Attachments: nil,
				UnfurlLinks: false,
				UnfurlMedia: true,
				IconURL:     "",
				IconEmoji:   ":soccer:",
				Markdown:    true,
				EscapeText:  true,
			})
		default:
			return
		}
	} else {
		post.PostMessage(channel, ">>>Hey it's Saturday but the regular season is over. I can't help anymore because I haven't been able to checkout the playoff schedule in time. :disappointed:\n@taylorsturtz turn off this cron!", slack.PostMessageParameters{
			Username:    "soccerbot",
			User:        "soccerbot",
			AsUser:      false,
			Parse:       "",
			LinkNames:   1,
			Attachments: nil,
			UnfurlLinks: false,
			UnfurlMedia: true,
			IconURL:     "",
			IconEmoji:   ":soccer:",
			Markdown:    true,
			EscapeText:  true,
		})
	}
}

func main() {

	token := os.Getenv("SLACK_TOKEN")
	channel := "C8TCVL3LN" // #testingbots
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	go runSaturdayReminderCron(api, channel)

	rn := time.Now().UTC()
	endOfSeason := time.Date(2018, 3, 12, 0, 0, 0, 0, time.UTC) // march 11, 2018 is the last game
	if endOfSeason.After(rn) {
		api.PostMessage(channel, "Hey guys/girls, I was just updated. :sunglasses:\nSorry about the ping this morning - I will explicitly check now if it's Saturday and auto-send game reminders at noon. :wink:\nTag me in this channel by typing `@soccerbot help` to see what else I can do!", slack.PostMessageParameters{
			Username:    "soccerbot",
			User:        "soccerbot",
			AsUser:      false,
			Parse:       "",
			LinkNames:   0,
			Attachments: nil,
			UnfurlLinks: false,
			UnfurlMedia: true,
			IconURL:     "",
			IconEmoji:   ":soccer:",
			Markdown:    true,
			EscapeText:  true,
		})
	} else {
		api.PostMessage(channel, "Hey guys/girls I just came online! :tada: Unfortunately the regular season is over and I can't help with much 'til next season.\n@taylorsturtz Update me!", slack.PostMessageParameters{
			Username:    "soccerbot",
			User:        "soccerbot",
			AsUser:      false,
			Parse:       "",
			LinkNames:   1,
			Attachments: nil,
			UnfurlLinks: false,
			UnfurlMedia: true,
			IconURL:     "",
			IconEmoji:   ":soccer:",
			Markdown:    true,
			EscapeText:  true,
		})
	}

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:

				fmt.Printf("Message: %v\n", ev)
				info := rtm.GetInfo()
				prefix := fmt.Sprintf("<@%s> ", info.User.ID)

				if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
					respond(rtm, ev, prefix)
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				// Take no action
			}
		}
	}
}

func getMonth(d string) int {
	// month // contains func
	var m int
	c := strings.Contains
	if c(d, "January") {
		m = 1
	} else if c(d, "February") {
		m = 2
	} else if c(d, "March") {
		m = 3
	} else if c(d, "April") {
		m = 4
	} else if c(d, "May") {
		m = 5
	} else if c(d, "June") {
		m = 6
	} else if c(d, "July") {
		m = 7
	} else if c(d, "August") {
		m = 8
	} else if c(d, "September") {
		m = 9
	} else if c(d, "October") {
		m = 10
	} else if c(d, "November") {
		m = 11
	} else if c(d, "December") {
		m = 12
	}
	return m
}

func respond(rtm *slack.RTM, msg *slack.MessageEvent, prefix string) {

	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	fmt.Printf("%s\n", text)
	var response string

	fullSchedule := map[string]bool{
		"schedule":         true,
		"show schedule":    true,
		"full schedule":    true,
		"full":             true,
		"games":            true,
		"all games":        true,
		"all matches":      true,
		"all":              true,
		"upcoming":         true,
		"upcoming games":   true,
		"upcoming matches": true,
	}

	nextGame := map[string]bool{
		"next":                   true,
		"next game":              true,
		"game":                   true,
		"next match":             true,
		"when is the next game?": true,
		"whens the next game?":   true,
	}

	hello := map[string]bool{
		"hi":            true,
		"hello":         true,
		"hey":           true,
		"hey soccerbot": true,
		"hi soccerbot":  true,
	}

	if fullSchedule[text] {
		rtm.SendMessage(rtm.NewOutgoingMessage("_Please be patient while I go check!_ :wink:", msg.Channel))
		// scrape soccer6 schedule
		doc, err := goquery.NewDocument("https://soccer6.net/schedule/")
		if err != nil {
			log.Fatal(err)
		}
		// find calvary chapel's games and return all of them
		doc.Find(".schedule-date .team-133").Each(func(index int, item *goquery.Selection) {
			date_ := item.Parent().Parent().Parent().Parent().Parent().Find("h5").Text()
			time_ := item.Parent().Parent().Parent().Parent().Find(".match-info .datetime-dropdown").Text()
			field := item.Parent().Parent().Parent().Parent().Find(".match-info .venue-dropdown a").Text()
			fieldSlice := field[5:12]
			score := item.Parent().Parent().Find(".match-vs .visible-print-inline").Text()
			scoreHome := ""
			scoreAway := ""
			away := item.Parent().HasClass("away-team")
			otherTeam := ""
			if away {
				otherTeam = item.Parent().Parent().Find(".home-team .match-team").Text()
			} else {
				otherTeam = item.Parent().Parent().Find(".away-team .match-team").Text()
			}
			if score == " : " {
				score = ""
			} else {
				scoreSplit := strings.Split(score, " : ")
				scoreHome = strings.TrimSpace(scoreSplit[0])
				scoreAway = strings.TrimSpace(scoreSplit[1])
				score = "- *(" + score + ")*"
			}
			winOrLose := ""
			winOrLoseEmoji := ""
			upcoming := false
			scoreHome_, err := strconv.Atoi(scoreHome)
			if err != nil {
				upcoming = true
			}
			scoreAway_, err := strconv.Atoi(scoreAway)
			if err != nil {
				upcoming = true
			}
			if !upcoming {
				if away {
					if scoreAway_ > scoreHome_ {
						winOrLose = "_Win_"
						winOrLoseEmoji = ":grinning:"
					} else if scoreAway_ < scoreHome_ {
						winOrLose = "_Loss_"
						winOrLoseEmoji = ":unamused:"
					} else if scoreAway_ == scoreHome_ {
						winOrLose = "_Draw_"
						winOrLoseEmoji = ":neutral_face:"
					}
				} else {
					if scoreAway_ < scoreHome_ {
						winOrLose = "_Win_"
						winOrLoseEmoji = ":grinning:"
					} else if scoreAway_ > scoreHome_ {
						winOrLose = "_Loss_"
						winOrLoseEmoji = ":unamused:"
					} else if scoreAway_ == scoreHome_ {
						winOrLose = "_Draw_"
						winOrLoseEmoji = ":neutral_face:"
					}
				}
			}
			response := fmt.Sprintf(">>>%s%s%s\nAt %s, on %s, vs. _%s_ %s %s %s\n", "*", date_, "*", strings.TrimSpace(time_), fieldSlice, strings.TrimSpace(otherTeam), score, winOrLoseEmoji, winOrLose)
			rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
		})
	} else if nextGame[text] {
		rtm.SendMessage(rtm.NewOutgoingMessage("_Please be patient while I go check!_ :wink:", msg.Channel))
		// scrape soccer6 schedule
		doc, err := goquery.NewDocument("https://soccer6.net/schedule/")
		if err != nil {
			log.Fatal(err)
		}

		// get current date to compare against game weeks
		now := time.Now().UTC()
		// set flag to only grab next game
		nextGameOnly := true

		// find calvary chapel's games and return the first one
		doc.Find(".schedule-date .team-133").Each(func(index int, item *goquery.Selection) {
			date_ := item.Parent().Parent().Parent().Parent().Parent().Find("h5").Text()
			time_ := item.Parent().Parent().Parent().Parent().Find(".match-info .datetime-dropdown").Text()
			time_ = strings.TrimSpace(time_)
			// scrape and parse date into this format from soccer6
			monthNum := getMonth(date_)
			month := time.Month(monthNum)
			daySplit := strings.Split(date_, " ")
			dayString := daySplit[2]
			day, err := strconv.Atoi(dayString)
			fmt.Println(err)
			hourSplit := strings.Split(time_, ":")
			hourString := hourSplit[0]
			hour, err := strconv.Atoi(hourString)
			fmt.Println(err)
			clockEmoji := ":clock11:"
			if hour == 12 {
				clockEmoji = ":clock12:"
			} else if hour == 1 {
				clockEmoji = ":clock1:"
			}
			// update year (TODO: there is a data-date available on the page - grab that and refactor month and day above)
			gameDate := time.Date(2018, month, day, hour, 0, 0, 0, time.UTC)
			// only show next game
			if (gameDate.After(now) || gameDate.Equal(now)) && nextGameOnly == true {
				field := item.Parent().Parent().Parent().Parent().Find(".match-info .venue-dropdown a").Text()
				fieldSlice := field[5:12]
				fieldEmoji := ":stadium:"
				away := item.Parent().HasClass("away-team")
				otherTeam := ""
				otherTeamEmoji := ":busts_in_silhouette:"
				if away {
					otherTeam = item.Parent().Parent().Find(".home-team .match-team").Text()
				} else {
					otherTeam = item.Parent().Parent().Find(".away-team .match-team").Text()
				}
				response = fmt.Sprintf("The next game is on %s%s%s\n>>>%s %s\n%s %s\n%s _%s_\n", "*", date_, "*", clockEmoji, time_, fieldEmoji, fieldSlice, otherTeamEmoji, strings.TrimSpace(otherTeam))
				nextGameOnly = false
			}
		})
		// handle no upcoming matches
		if nextGameOnly == true {
			response = fmt.Sprintf("%sBummer.%s No upcoming regular season matches 'til next season. :disappointed:", "*", "*")
		}
		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	} else if hello[text] {
		response := fmt.Sprintf("Hey buddy, let's go play some soccer!! :soccer::runner::dash:")
		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	} else {
		response := fmt.Sprintf(">>>Hey :smiley:\nType `@soccerbot next` for next game\nType `@soccerbot all` for all games")
		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	}
}
