package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fastwego/feishu"
	"github.com/fastwego/feishu/apis/bot/group_manage"
	"github.com/fastwego/feishu/apis/message"
	"github.com/fastwego/feishu/types/event_types"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var App *feishu.App
var rdb *redis.Client
var ctx = context.Background()

func init() {
	viper.SetConfigFile("config.json")
	_ = viper.ReadInConfig()

	App = feishu.NewApp(feishu.AppConfig{
		AppId: viper.GetString("AppId"),
		AppSecret: viper.GetString("AppSecret"),
		VerificationToken: viper.GetString("VerificationToken"),
		EncryptKey: viper.GetString("EncryptKey"),
	})

	rdb = redis.NewClient(&redis.Options{
		Addr: viper.GetString("RedisAddr"),
		Password: viper.GetString("RedisPassword"),
		DB: viper.GetInt("RedisDB"),
	})
}

func fetchFeedContent()  {
	defaultFeedUrls := viper.GetStringSlice("DefaultFeedUrls")
	fp := gofeed.NewParser()
	for _, url := range defaultFeedUrls {
		fmt.Println(url)
		feed, _ := fp.ParseURL(url)
		latestVersion := feed.Items[0].Title
		val, _ := rdb.Get(ctx, url).Result()
		if len(val) == 0 {
			rdb.Set(ctx, url, latestVersion, 0)
		} else if val != latestVersion {
			rdb.Set(ctx, url, latestVersion, 0)
			sendFeishuTextMessageToAllChatList(feed.Title + ": " +  latestVersion + " published! Let's see what's new! \n" + url[0:len(url) - 5])
		}
	}
}

func sendFeishuTextMessageToAllChatList(content string)  {
	params := url.Values{}
	chatListResp, _ := group_manage.ChatList(App, params)

	chatListRespJson := []byte(string(chatListResp))

	var chatListResponse ChatListResponse
	jsonParseErr := json.Unmarshal(chatListRespJson, &chatListResponse)

	if jsonParseErr != nil {
		fmt.Println(jsonParseErr)
		return
	}


	for _, chatItem := range chatListResponse.Data.Groups {

		releaseUpdateMessage := TextMessage{
			ChatId: chatItem.ChatId,
			MessageType: "text",
			Content: TextMessageContent{
				Text: content,
			},
		}

		releaseUpdateMessageJson, _ := json.Marshal(releaseUpdateMessage)
		fmt.Println(releaseUpdateMessageJson)

		_, messageErr := message.Send(App, releaseUpdateMessageJson)
		if messageErr != nil {
			fmt.Println(messageErr)
		}
	}
}

func crontabTask() {
	ticker := time.Tick(time.Minute * 5)
	for {
		<- ticker
		go fetchFeedContent()
	}
}

func main() {
	fetchFeedContent()
	go crontabTask()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.POST("/api/feishu/rss-robot", RSSRobot)

	svr := &http.Server {
	   Addr: viper.GetString("LISTEN"),
	   Handler: router,
	}

	go func ()  {
		err := svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	timeout := time.Duration(5) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := svr.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}

}

func RSSRobot(c *gin.Context) {
	event, err := App.Server.ParseEvent(c.Request)

	if err != nil {
		return
	}

	switch event.(type) {
	case event_types.EventChallenge:
		App.Server.Challenge(c.Writer, event.(event_types.EventChallenge))

	case event_types.EventMessageText:
		userMsg := event.(event_types.EventMessageText)
		defaultFeedUrls := viper.GetStringSlice("DefaultFeedUrls")
		content := "Source Code: Subscrible List: \n"
		for _, url := range defaultFeedUrls {
			content += url + "\n"
		}
		defaultReplyMessage := TextMessage{
			ChatId: userMsg.Event.OpenChatID,
			MessageType: "text",
			Content: TextMessageContent{
				Text: content,
			},
		}
		defaultReplyMessageJson, _ := json.Marshal(defaultReplyMessage)
		message.Send(App, defaultReplyMessageJson)
	}
}