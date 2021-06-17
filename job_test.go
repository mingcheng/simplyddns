package simplyddns

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	notify "repo.wooramel.cn/mingcheng/srk-notification"
)

func TestNewNSQSender(t *testing.T) {
	notification, err := notify.NewNSQSender(notify.NSQConfig{
		Host:  "172.16.1.70:4150",
		Topic: "srk-notification",
	})
	assert.NoError(t, err)
	assert.NotNil(t, notification)

	err = notification.Send(notify.Message{
		Type:      "sleep",
		Timestamp: time.Now(),
		Subject:   time.Now().String(),
		Content:   time.Now().String(),
		Receiver:  time.Now().String(),
	})
	assert.NoError(t, err)
}
