/**
 * File: misc.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Sunday, December 27th 2020, 7:01:38 pm
 * Last Modified: Sunday, December 27th 2020, 8:38:55 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"github.com/go-resty/resty/v2"
	ddns "github.com/mingcheng/simplyddns"
	"net"
)

var (
	log = ddns.NewLogger()
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"

func RawIPByURL(url string) (addr net.IP, err error) {
	var data string
	data, err = RawStrByURL(context.Background(), url, nil)
	if err != nil {
		return nil, err
	}

	return net.ParseIP(data), nil
}

func RawStrByURL(ctx context.Context, url string, headers map[string]string) (result string, err error) {
	var (
		req  *resty.Request
		resp *resty.Response
	)

	req = resty.New().R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Referer":    url,
			"User-Agent": UserAgent,
		})

	if headers != nil {
		req.SetHeaders(headers)
	}

	resp, err = req.Get(url)

	if !resp.IsSuccess() {
		return "", err
	}

	return string(resp.Body()), nil
}
