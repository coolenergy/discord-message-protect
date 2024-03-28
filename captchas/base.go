package captchas

type ICaptchaValidator interface {
	Initialize() (bool, error)
	ValidateCaptcha(value string) (bool, error)
	AddHtml(htmlTemplate string) string
}
