package fritzbox

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

type Session interface {
	GetSID() string
	IsValid() bool
	Used()
}

type session struct {
	SID      string
	Created  time.Time
	LastUsed time.Time
}

type sessionInfo struct {
	XMLName   xml.Name `xml:"SessionInfo"`
	SID       string   `xml:"SID"`
	Challenge string   `xml:"Challenge"`
	BlockTime int      `xml:"BlockTime"`
	Rights    rights   `xml:"Rights"`
	Users     users    `xml:"Users"`
}

type rights struct {
	Name   []string `xml:"Name"`
	Access []int    `xml:"Access"`
}

type users struct {
	User []user `xml:"User"`
}

type user struct {
	Last     bool   `xml:"last,attr,omitempty"`
	Username string `xml:",chardata"`
}

type challenge struct {
	Version string
	Iter1   int
	Salt1   string
	Iter2   int
	Salt2   string
}

func createSession(si *sessionInfo) Session {
	return &session{
		SID:      si.SID,
		Created:  time.Now(),
		LastUsed: time.Now(),
	}
}

func (s *session) GetSID() string {
	return s.SID
}

func (s *session) IsValid() bool {
	return s.SID != "0000000000000000" && time.Since(s.LastUsed) < 20*time.Minute
}

func (s *session) Used() {
	s.LastUsed = time.Now()
}

func getSessionInfo(fc *fritzClient) (*sessionInfo, error) {
	resp, err := fc.client.Get(fmt.Sprintf("%s/login_sid.lua?version=2", fc.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return unmarshalSessionInfo(resp.Body)
}

func unmarshalSessionInfo(body io.Reader) (*sessionInfo, error) {
	var si sessionInfo
	if unmarshalErr := xml.NewDecoder(body).Decode(&si); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &si, nil
}

func parseChallenge(c string) (*challenge, error) {
	parts := strings.Split(c, "$")
	if len(parts) != 5 || parts[0] != "2" {
		return nil, fmt.Errorf("invalid challenge format or unsupported version: %s", c)
	}

	iter1, err := parseInt(parts[1])
	if err != nil {
		return nil, err
	}

	iter2, err := parseInt(parts[3])
	if err != nil {
		return nil, err
	}

	return &challenge{
		Version: parts[0],
		Iter1:   iter1,
		Salt1:   parts[2],
		Iter2:   iter2,
		Salt2:   parts[4],
	}, nil
}

func calculateResponse(challenge *challenge, password string) (string, error) {
	salt1, err := hex.DecodeString(challenge.Salt1)
	if err != nil {
		return "", err
	}

	salt2, err := hex.DecodeString(challenge.Salt2)
	if err != nil {
		return "", err
	}

	hash1, _ := pbkdf2.Key(sha256.New, password, salt1, challenge.Iter1, sha256.Size)

	// Second PBKDF2 hash with dynamic salt
	hash2, _ := pbkdf2.Key(sha256.New, string(hash1), salt2, challenge.Iter2, sha256.Size)
	hash2Hex := hex.EncodeToString(hash2)

	// Format response as specified
	response := fmt.Sprintf("%s$%s", challenge.Salt2, hash2Hex)
	return response, nil
}

func parseInt(s string) (int, error) {
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}
