package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"golang.org/x/net/context"

	"github.com/luci/gae/service/memcache"
	"github.com/luci/luci-go/common/logging"
)

type crRev simpleClient

func (cr *crRev) GetJSON(c context.Context, pos string) (map[string]string, error) {
	logging.Infof(c, "GetJSON")

	itm := memcache.NewItem(c, fmt.Sprintf("crrev:%s", pos))
	err := memcache.Get(c, itm)

	if err == memcache.ErrCacheMiss {
		hc, err := getAsSelfOAuthClient(c)
		if err != nil {
			return nil, err
		}

		resp, err := hc.Get(fmt.Sprintf(cr.Host+"/_ah/api/crrev/v1/redirect/%s", pos))
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		itm.SetValue(body)
		if err = memcache.Set(c, itm); err != nil {
			return nil, fmt.Errorf("while setting memcache: %s", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("while getting from memcache: %s", err)
	}

	m := map[string]string{}
	err = json.Unmarshal(itm.Value(), &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func WithCrRev(c context.Context, baseURL string) context.Context {
	logging.Infof(c, "WithCrRev, %s", baseURL)
	cr := &crRev{Host: baseURL, Client: nil}
	c = context.WithValue(c, crRevKey, cr)
	return c
}

func GetCrRev(c context.Context) *crRev {
	ret, ok := c.Value(crRevKey).(*crRev)
	logging.Infof(c, "GetCrRev, %v, %v", ret, ok)
	if !ok {
		panic("No crrev client set in context")
	}
	return ret
}
