package function

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

const urlData = "https://sigarra.up.pt/feup/pt/ASSD_TLP_GERAL.FUNC_VIEW"
const urlLogin = "https://sigarra.up.pt/feup/pt/vld_validacao.validacao"
const urlLogout = "https://sigarra.up.pt/feup/pt/vld_validacao.sair"
const loginApp = "162"
const userAgent = "Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.1) Gecko/2008071615 Fedora/3.0.1-1.fc9 Firefox/3.0.1 DedoFEUP/1.0"

// these two depend on urlData
const loginAmo = "1674"
const loginAddress = "ASSD_TLP_GERAL.FUNC_VIEW"

// Day holds attendance info for one single day
type Day struct {
	Type,
	Date,
	Balance,
	BalanceAccrual,
	Unjustified,
	UnjustifiedAccrual,
	MorningIn,
	MorningOut,
	AfternoonIn,
	AfternoonOut string
}

// GetToken extract token from cookiejar
func GetToken(client *http.Client) string {
	u, err := url.Parse(urlData)
	if err != nil {
		log.Fatal(err)
	}

	token := make([]string, 2)

	for _, c := range client.Jar.Cookies(u) {
		if c.Name == "SI_SESSION" {
			token[0] = c.Value
		}
		if c.Name == "SI_SECURITY" {
			token[1] = c.Value
		}
	}

	return strings.Join(token, "#")
}

// Login log in to Sigarra
func Login(username, password string) (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Jar: jar}

	form := url.Values{
		"p_app":     {loginApp},
		"p_amo":     {loginAmo},
		"p_address": {loginAddress},
		"p_user":    {username},
		"p_pass":    {password},
	}
	req, err := http.NewRequest("POST", urlLogin, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.PostForm = form
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if !strings.Contains(
		string(body),
		"<meta http-equiv=\"Refresh\" content=\"0;url="+urlData+"\">",
	) {
		return "", fmt.Errorf("invalid login")
	}

	return GetToken(client), nil
}

// Logout terminate session in Sigarra
func Logout(tokenString string) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client := &http.Client{Jar: jar}

	err = setupJar(tokenString, jar)
	if err != nil {
		return err
	}

	form := url.Values{
		"p_address": {loginAddress},
	}
	req, err := http.NewRequest("POST", urlLogout, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.PostForm = form
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// GetData retrieve the attendance data from Sigarra
func GetData(tokenString string) ([]Day, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	err = setupJar(tokenString, jar)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", urlData, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(body)

	if strings.Contains(html, "<img src=\"/feup/pt/imagens/SemPermissoes\"") {
		return nil, fmt.Errorf("not logged in")
	}

	dayRows := regexp.MustCompile(`(?s)<tr class="dia-(.+?)\b.*?">(.*?)</tr>`)
	dayCols := regexp.MustCompile(`(?s)<td .*?class="(.*?)">(.*?)</td>`)
	timeRE := regexp.MustCompile(`-?\d+:\d+`)
	var days []Day
	var data string
	for _, match := range dayRows.FindAllStringSubmatch(html, -1) {
		day := Day{}
		day.Type = match[1]
		for _, match2 := range dayCols.FindAllStringSubmatch(match[2], -1) {
			if match2[2] == "---" {
				data = "---"
			} else {
				data = timeRE.FindString(match2[2])
			}

			switch {
			case strings.Contains(match2[1], "data k"):
				day.Date = match2[2]
			case strings.Contains(match2[1], "saldo-d"):
				day.Balance = data
			case strings.Contains(match2[1], "saldo-a"):
				day.BalanceAccrual = data
			case strings.Contains(match2[1], "injust-d"):
				day.Unjustified = data
			case strings.Contains(match2[1], "injust-a"):
				day.UnjustifiedAccrual = data
			case strings.Contains(match2[1], "marca am"):
				if day.MorningIn == "" {
					day.MorningIn = data
				} else {
					day.MorningOut = data
				}
			case strings.Contains(match2[1], "marca pm"):
				if day.AfternoonIn == "" {
					day.AfternoonIn = data
				} else {
					day.AfternoonOut = data
				}
			case strings.Contains(match2[1], "horario k"):
				// ignore this column
			case strings.Contains(match2[2], ">Lupa<"):
				// ignore this column
			default:
				// log unexpected ones for debugging
				log.Printf("Unkown class %s with value %s\n", match2[1], match2[2])
			}
		}
		days = append(days, day)
	}

	return days, nil
}

func setupJar(tokenString string, jar *cookiejar.Jar) error {
	u, err := url.Parse(urlData)
	if err != nil {
		return err
	}
	token := strings.Split(tokenString, "#")
	if len(token) != 2 {
		return fmt.Errorf("invalid token")
	}

	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "SI_SESSION",
		Value:  token[0],
		Path:   "/",
		Domain: u.Hostname(),
	}
	cookies = append(cookies, cookie)
	cookie = &http.Cookie{
		Name:   "SI_SECURITY",
		Value:  token[1],
		Path:   "/",
		Domain: u.Hostname(),
	}
	cookies = append(cookies, cookie)
	jar.SetCookies(u, cookies)
	return nil
}
