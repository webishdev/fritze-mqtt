package fritzbox

import "testing"

func Test_calculateResponse(t *testing.T) {
	challenge, err := parseChallenge("2$10000$5A1711$2000$5A1722")
	if err != nil {
		t.Error(err)
	}

	response, err := calculateResponse(challenge, "1example!")
	if err != nil {
		t.Error(err)
	}

	if response != "5A1722$1798a1672bca7c6463d6b245f82b53703b0f50813401b03e4045a5861e689adb" {
		t.Error("invalid response")
	}
}
