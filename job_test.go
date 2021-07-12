package simplyddns

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	notify "repo.wooramel.cn/mingcheng/srk-notification"
)

func TestNewNSQSender(t *testing.T) {
	nsqAddr, ok := os.LookupEnv("NSQ_ADDR")
	if !ok {
		fmt.Println("NSQ_ADDR is not set, so ignore")
		return
	}

	notification, err := notify.NewNSQSender(notify.NSQConfig{
		Host:  nsqAddr,
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
