package notifiers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/armory-io/dinghy/pkg/settings"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

/* For reference, the JSON payload for the Echo endpoint for sending a Slack
 * message should look like this:
 *
 * {
 *   "notificationType":9,
 *   "to":["bot-testing"],    // leading # is optional
 *	 "additionalContext":{
 *		 "body":"This is a test",
 *		 "formatter":"TEXT"
 *	 }
 * }
 */

type SlackNotification struct {
	Type    int      `json:"notificationType" yaml:"notificationType"`
	To      []string `json:"to" yaml:"to"`
	Context context  `json:"additionalContext" yaml:"additionalContext"`
}

type context struct {
	Body      string `json:"body" yaml:"body"`
	Formatter string `json:"formatter" yaml:"formatter"`
}

type SlackNotifier struct {
	Echo    string
	Channel string
}

func newSlackNotification(to, body string) *SlackNotification {
	// NOTE / TODO SLACK is an enum in Java.  Appears to be a "9" at the moment
	// Not sure how best to make sure this stays in sync.
	sn := &SlackNotification{Type: 9}
	sn.To = []string{to}
	sn.Context.Body = body
	sn.Context.Formatter = "TEXT"
	return sn
}

// TODO: Configure channel
func NewSlackNotifier(s *settings.ExtSettings) *SlackNotifier {
	return &SlackNotifier{
		Echo:    s.Settings.Echo.BaseURL + "/notifications",
		Channel: s.Notifiers.Slack.Channel,
	}
}

func (sn *SlackNotifier) SendSuccess(org, repo, path string) {
	sn.sendToEcho(newSlackNotification(sn.Channel, fmt.Sprintf("Dinghy Successful Update: %s/%s/%s", org, repo, path)))
}

func (sn *SlackNotifier) SendFailure(org, repo, path string, err error) {
	sn.sendToEcho(newSlackNotification(sn.Channel, fmt.Sprintf("Dinghy Failed Update: %s/%s/%s", org, repo, path)))
}

// Actually ships the notification off to the echo endpoint.
func (sn *SlackNotifier) sendToEcho(n *SlackNotification) error {
	postData, err := json.Marshal(n)
	if err != nil {
		return err
	}
	req, err := retryablehttp.NewRequest(http.MethodPost, sn.Echo, postData)
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = retryablehttp.NewClient().Do(req)
	if err != nil {
		return err
	}
	return nil
}
