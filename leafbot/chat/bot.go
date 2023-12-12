package chat

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Nomango/ark"
	"github.com/Nomango/ark/slices"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	client *openai.Client
}

func NewChatBot() (*Bot, error) {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	proxy, err := url.Parse("http://127.0.0.1:7890")
	if err != nil {
		return nil, errors.WithMessage(err, "invalid OpenAI Proxy")
	}
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	client := openai.NewClientWithConfig(config)
	return &Bot{client: client}, nil
}

func (b *Bot) CreateThread(ctx context.Context) (string, error) {
	req := openai.ThreadRequest{
		Metadata: map[string]any{},
	}
	thread, err := b.client.CreateThread(ctx, req)
	if err != nil {
		return "", err
	}
	logrus.Debugln("create thread", thread.ID)
	return thread.ID, nil
}

func (b *Bot) SubmitMessage(ctx context.Context, assistantID, threadID, message string) ([]string, error) {
	msgReq := openai.MessageRequest{
		Role:     openai.ChatMessageRoleUser,
		Content:  message,
		Metadata: map[string]any{},
	}
	msg, err := b.client.CreateMessage(ctx, threadID, msgReq)
	if err != nil {
		return nil, err
	}
	logrus.Debugln("create message", threadID, msg.ID)

	model := openai.GPT3Dot5Turbo1106
	runReq := openai.RunRequest{
		AssistantID: assistantID,
		Model:       &model,
		Metadata:    map[string]any{},
	}
	run, err := b.client.CreateRun(ctx, threadID, runReq)
	if err != nil {
		return nil, err
	}
	logrus.Debugln("start run", threadID, run.ID)

	timeout := time.After(45 * time.Second)
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, errors.New("run timeout")
		case <-tick.C:
			run, err = b.client.RetrieveRun(ctx, threadID, run.ID)
			if err != nil {
				return nil, err
			}
			switch run.Status {
			case openai.RunStatusFailed:
				return nil, errors.New("run failed")
			case openai.RunStatusCompleted:
				goto done
			case openai.RunStatusExpired:
				return nil, errors.New("run expired")
			case openai.RunStatusRequiresAction:
				// TODO: 必须处理这个状态
				return nil, errors.New("run requires action")
			default:
				continue
			}

		}
	}
done:
	msgList, err := b.client.ListMessage(ctx, threadID, ark.Ptr(10), ark.Ptr("desc"), nil, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "list message failed")
	}
	msgs := slices.Filter(msgList.Messages, func(i openai.Message) bool {
		if i.RunID != nil && *i.RunID == run.ID {
			return true
		}
		return false
	})
	if len(msgs) == 0 {
		return nil, errors.New("no message found")
	}
	return slices.Map(msgs, func(m openai.Message) string { return m.Content[0].Text.Value }), nil
}
