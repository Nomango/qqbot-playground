package chat

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

var bot = Must(NewChatBot())

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

var (
	threads = sync.Map{}
	sf      singleflight.Group
)

type ThreadContext struct {
	ThreadID string
	ch       chan struct{}
}

func (c *ThreadContext) Lock() {
	c.ch <- struct{}{}
}

func (c *ThreadContext) Unlock() {
	<-c.ch
}

func ResetThread(groupKey string) {
	threads.Delete(groupKey)
}

func Chat(ctx context.Context, groupKey, message string) ([]string, error) {
	thread, err := getThread(ctx, groupKey)
	if err != nil {
		return nil, errors.WithMessage(err, "getting thread failed")
	}
	thread.Lock()
	defer thread.Unlock()
	return bot.SubmitMessage(ctx, "asst_N2y1zz1uSpqpxA7C1HIXkbEu", thread.ThreadID, message)
}

func getThread(ctx context.Context, groupKey string) (*ThreadContext, error) {
	if v, ok := threads.Load(groupKey); ok {
		return v.(*ThreadContext), nil
	}
	v, err, _ := sf.Do("new-thread-"+groupKey, func() (any, error) {
		if v, ok := threads.Load(groupKey); ok {
			return v.(*ThreadContext), nil
		}
		// id, err := bot.CreateThread(ctx)
		// if err != nil {
		// 	return nil, err
		// }
		id := "thread_qprKW4KUNcyAcRmN1P2ivKm9"
		thread := &ThreadContext{
			ThreadID: id,
			ch:       make(chan struct{}, 1),
		}
		return thread, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*ThreadContext), nil
}
