package serviceuser

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/paulscherrerinstitute/scicat-cli/v3/datasetUtils"
)

type ScicatServiceUser struct {
	scicatUrl   string
	username    string
	password    string
	scicatToken string
	expiry      time.Time
	mutex       sync.Mutex
}

func (su *ScicatServiceUser) GetToken() (string, error) {
	su.mutex.Lock()
	defer su.mutex.Unlock()
	if time.Now().After(su.expiry) {
		err := su.refreshToken()
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func (su *ScicatServiceUser) refreshToken() error {
	user, _, err := datasetUtils.AuthenticateUser(http.DefaultClient, su.scicatUrl, su.username, su.password, false)
	if err != nil {
		return err
	}
	token, ok := user["accessToken"]
	if !ok {
		return fmt.Errorf("token wasn't part of the user struct")
	}
	su.scicatToken = token
	// TODO: add a way to check the next expiry date!
	return nil
}
