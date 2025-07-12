package fritzbox

import (
	"crypto/tls"
	"fmt"
	"github.com/webishdev/fritze-mqtt/log"
	"net/http"
	"net/url"
	"time"
)

type FritzClient interface {
	Login(username string, password string) (Session, error)
	Logout(s Session) error
	GetDevices(s Session) ([]Device, error)
}

type fritzClient struct {
	baseURL string
	client  *http.Client
}

func NewFritzClient(baseURL string) FritzClient {
	return &fritzClient{
		baseURL: baseURL,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (fc *fritzClient) Login(username string, password string) (Session, error) {
	initialSessionInfo, err := getSessionInfo(fc)
	if err != nil {
		return nil, err
	}

	log.PrintXML(initialSessionInfo)

	if initialSessionInfo.BlockTime > 0 {
		log.Info("waiting for %d seconds", initialSessionInfo.BlockTime)
		time.Sleep(time.Duration(initialSessionInfo.BlockTime) * time.Second)
	}

	c, err := parseChallenge(initialSessionInfo.Challenge)
	if err != nil {
		return nil, err
	}

	response, err := calculateResponse(c, password)
	if err != nil {
		return nil, err
	}

	// Perform login POST request
	formData := url.Values{}
	formData.Set("username", username)
	formData.Set("response", response)

	loginURL := fmt.Sprintf("%s/login_sid.lua?version=2", fc.baseURL)

	resp, err := fc.client.PostForm(loginURL, formData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	si, err := unmarshalSessionInfo(resp.Body)
	if err != nil {
		return nil, err
	}

	log.PrintXML(si)

	s := createSession(si)

	if !s.IsValid() {
		return nil, fmt.Errorf("login failed")
	}

	log.Info("Successfully logged in as user=%s at %s with sid=%s", username, fc.baseURL, s.GetSID())

	return s, nil
}

func (fc *fritzClient) Logout(s Session) error {
	if !s.IsValid() {
		return fmt.Errorf("session is not valid")
	}
	logoutURL := fmt.Sprintf("%s/login_sid.lua?version=2&logout&sid=%s", fc.baseURL, s.GetSID())

	resp, err := fc.client.Get(logoutURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Info("logout")

	return nil
}

func (fc *fritzClient) GetDevices(s Session) ([]Device, error) {
	s.Used()
	return getDeviceListInfos(fc, s)
}
