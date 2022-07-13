package simplyddns

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJob_RunWebhook(t *testing.T) {

	j := Job{
		Config: &JobConfig{
			WebHook: WebHook{
				Url: "https://httpbin.org/post",
			},
		},
	}

	err := j.RunWebhook(context.TODO(), "127.0.0.1", []string{"a.com", "b.com"})
	assert.NoError(t, err)
}
