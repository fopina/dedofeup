package function

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"gotest.tools/assert"
)

func TestFailedLogin(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://sigarra.up.pt/feup/pt/vld_validacao.validacao",
		httpmock.NewStringResponder(200, "wtv"),
	)

	_, err := Login("a", "b")

	assert.Equal(t, httpmock.GetTotalCallCount(), 1)
	info := httpmock.GetCallCountInfo()
	assert.Equal(
		t,
		info["POST https://sigarra.up.pt/feup/pt/vld_validacao.validacao"],
		1,
	)
	assert.Error(t, err, "invalid login")
}

func TestGoodLogin(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	x := httpmock.NewStringResponse(
		200,
		`<meta http-equiv="Refresh" content="0;url=https://sigarra.up.pt/feup/pt/ASSD_TLP_GERAL.FUNC_VIEW">`,
	)
	x.Header.Add("Set-Cookie", "SI_SESSION=123")
	x.Header.Add("Set-Cookie", "SI_SECURITY=456")
	x.Header.Add("Set-Cookie", "IGNORED=999")

	httpmock.RegisterResponder(
		"POST",
		"https://sigarra.up.pt/feup/pt/vld_validacao.validacao",
		httpmock.ResponderFromResponse(x),
	)

	token, err := Login("a", "b")
	assert.Equal(t, httpmock.GetTotalCallCount(), 1)
	info := httpmock.GetCallCountInfo()
	assert.Equal(
		t,
		info["POST https://sigarra.up.pt/feup/pt/vld_validacao.validacao"],
		1,
	)
	assert.NilError(t, err)
	assert.Equal(t, token, "123#456")
}

func TestGetDataBadToken(t *testing.T) {
	days, err := GetData("")
	assert.Equal(t, len(days), 0)
	assert.Error(t, err, "invalid token")
}

func TestGetDataTokenExpired(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	x := httpmock.NewStringResponse(
		200,
		"bla bla <img src=\"/feup/pt/imagens/SemPermissoes\" bla bla",
	)

	httpmock.RegisterResponder(
		"GET",
		"https://sigarra.up.pt/feup/pt/ASSD_TLP_GERAL.FUNC_VIEW",
		httpmock.ResponderFromResponse(x),
	)

	days, err := GetData("a#b")
	assert.Equal(t, len(days), 0)
	assert.Error(t, err, "not logged in")
}

func TestGetData(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	x := httpmock.NewStringResponse(
		200,
		string(helperLoadBytes(t, "getdata.html")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://sigarra.up.pt/feup/pt/ASSD_TLP_GERAL.FUNC_VIEW",
		httpmock.ResponderFromResponse(x),
	)

	days, err := GetData("a#b")
	assert.NilError(t, err)
	assert.DeepEqual(t, days, []Day{
		{
			Type:               "normal",
			Date:               "2019-07-01",
			Balance:            "0:05",
			BalanceAccrual:     "0:05",
			Unjustified:        "0:00",
			UnjustifiedAccrual: "0:00",
			MorningIn:          "09:41",
			MorningOut:         "13:02",
			AfternoonIn:        "13:55",
			AfternoonOut:       "16:46",
			FutureDay:          false,
		},
		{
			Type:               "normal",
			Date:               "2019-07-02",
			Balance:            "0:16",
			BalanceAccrual:     "0:21",
			Unjustified:        "0:00",
			UnjustifiedAccrual: "0:00",
			MorningIn:          "09:44",
			MorningOut:         "12:40",
			AfternoonIn:        "13:41",
			AfternoonOut:       "17:01",
			FutureDay:          false,
		},
		{
			Type:               "normal",
			Date:               "2019-07-03",
			Balance:            "0:20",
			BalanceAccrual:     "0:41",
			Unjustified:        "0:00",
			UnjustifiedAccrual: "0:00",
			MorningIn:          "09:20",
			MorningOut:         "13:02",
			AfternoonIn:        "13:56",
			AfternoonOut:       "16:40",
			FutureDay:          false,
		},
		{
			Type:               "normal",
			Date:               "2019-07-04",
			Balance:            "0:18",
			BalanceAccrual:     "0:59",
			Unjustified:        "2:30",
			UnjustifiedAccrual: "2:30",
			MorningIn:          "09:47",
			MorningOut:         "12:32",
			AfternoonIn:        "13:37",
			AfternoonOut:       "17:10",
			FutureDay:          false,
		},
		{
			Type:               "actual",
			Date:               "2019-07-05",
			Balance:            "-3:05",
			BalanceAccrual:     "0:59",
			Unjustified:        "0:00",
			UnjustifiedAccrual: "0:00",
			MorningIn:          "09:47",
			MorningOut:         "---",
			AfternoonIn:        "---",
			AfternoonOut:       "---",
			FutureDay:          false,
		},
		{
			Type:               "normal",
			Date:               "2019-07-06",
			Balance:            "0:00",
			BalanceAccrual:     "0:00",
			Unjustified:        "0:00",
			UnjustifiedAccrual: "0:00",
			MorningIn:          "---",
			MorningOut:         "---",
			AfternoonIn:        "---",
			AfternoonOut:       "---",
			FutureDay:          true,
		},
	})
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
