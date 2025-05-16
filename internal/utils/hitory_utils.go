package utils

import (
	"ai-cloud/internal/model"
	"github.com/cloudwego/eino/schema"
)

func MessageList2ChatHistory(mess []*model.Message) (history []*schema.Message) {
	for _, m := range mess {
		history = append(history, message2MessagesTemplate(m))
	}
	return
}

func message2MessagesTemplate(mess *model.Message) *schema.Message {
	return &schema.Message{
		Role:    schema.RoleType(mess.Role),
		Content: mess.Content,
	}
}
