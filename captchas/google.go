package captchas

import (
	"encoding/json"
	"fmt"
	"github.com/melardev/discord-message-protect/core"
	"io/ioutil"
	"net/http"
	"strings"
)

type GoogleReCaptchaValidator struct {
	SiteKey   string
	SecretKey string
}

type GoogleReCaptchaResponse struct {
	Success bool `json:"success"`
}

func NewReCaptchaValidator(config *core.HttpConfig) *GoogleReCaptchaValidator {
	g := &GoogleReCaptchaValidator{}
	args := config.Args

	if key, found := args["site_key"]; found {
		g.SiteKey = key.(string)
	}

	if secret, found := args["secret_key"]; found {
		g.SecretKey = secret.(string)
	}

	return g
}

// Initialize returns a boolean indicating if the initialization was successful and an error object if it was not
// with the details
func (g *GoogleReCaptchaValidator) Initialize() (bool, error) {
	// TODO: Validate dynamically the Google Keys
	return true, nil
}

// ValidateCaptcha validates the captcha response given by the user and provided here as the first argument
func (g *GoogleReCaptchaValidator) ValidateCaptcha(value string) (bool, error) {
	fields := map[string][]string{
		"secret":   {g.SecretKey},
		"response": {value},
	}

	response, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", fields)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	var dto GoogleReCaptchaResponse
	err = json.Unmarshal(content, &dto)
	if err != nil {
		panic(err)
	}

	return dto.Success, nil
}

func (g *GoogleReCaptchaValidator) AddHtml(template string) string {

	out := strings.Replace(template, "<CAPTCHA_JS>", `<script src='https://www.google.com/recaptcha/api.js'></script>`, -1)
	out = strings.Replace(out, "<CAPTCHA_CHALLENGE>",
		fmt.Sprintf(`<div class="g-recaptcha" data-sitekey="%s"></div>`, g.SiteKey), -1)

	return out
}
