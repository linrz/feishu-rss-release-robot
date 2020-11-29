package main

type TextMessage struct {
	ChatId string `json:"chat_id"`
	MessageType string `json:"msg_type"`
	Content TextMessageContent `json:"content"`
}

type TextMessageContent struct {
	Text string `json:"text"`
}

type ChatListResponse struct {
	Code int `json:"code"`
	Data ChatListData `json:"data"`
}

type ChatListData struct {
	Groups []ChatItem `json:"groups"`
}

type ChatItem struct {
	ChatId string `json:"chat_id"`
	Name string `json:"name"` 
}