package hello

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	opt := linebot.WithHTTPClient(client)
	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"), opt)
	if err != nil {
		log.Errorf(c, "failed to init linebot Client: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	events, err := bot.ParseRequest(r)
	if err != nil {
		log.Errorf(c, "failed to parse webhook: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, ev := range events {
		if ev.Type == linebot.EventTypeMessage {
			if b, err := json.MarshalIndent(ev, "", "  "); err != nil {
				log.Errorf(c, "failed to dump event: %s, %#v", err, ev)
			} else {
				log.Infof(c, "Event: %s", b)
			}
			switch message := ev.Message.(type) {
			case *linebot.TextMessage:
				source := ev.Source
				if source.Type == linebot.EventSourceTypeUser {
					log.Infof(c, "msg: %s", message.Text)
					if _, err = bot.PushMessage(source.UserID, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Errorf(c, "failed to push message: %s", err)
					}
				}
			}
		}
	}
	fmt.Fprint(w, "OK")
}
