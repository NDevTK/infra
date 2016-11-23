package som

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"github.com/luci/luci-go/appengine/gaeauth/client"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/logging"
	//  "github.com/luci/gae/service/info"
	"github.com/luci/gae/service/urlfetch"
)

var (
	gitilesScope = "https://www.googleapis.com/auth/gerritcodereview"
)

func getGitiles(c context.Context, URL string) ([]byte, error) {
	token, err := client.GetAccessToken(c, []string{gitilesScope})
	if err != nil {
		return nil, err
	}

	trans := auth.NewModifyingTransport(urlfetch.Get(c), func(req *http.Request) error {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		logging.Infof(c, "request: %+v", req)
		return nil
	})

	client := &http.Client{Transport: trans}

	resp, err := client.Get(URL)
	fmt.Printf("read: %v, %v\n", URL, err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status code %d from %s", resp.StatusCode, resp.Request.URL)
	}

	reader := base64.NewDecoder(base64.StdEncoding, resp.Body)
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return b, nil
}
