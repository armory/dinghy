package notifiers

import (
	"encoding/json"
	"fmt"
	"github.com/armory/plank/v4"
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

type SlackNotificationValues struct {
	When    []string `json:"when" yaml:"when" mapstructure:"when"`
	Address string   `json:"address" yaml:"address" mapstructure:"address"`
	Level   string   `json:"level" yaml:"level" mapstructure:"level"`
}

type SlackNotifications struct {
	SlackValues []SlackNotificationValues `json:"slack" yaml:"slack" mapstructure:"slack"`
}

type SlackNotification struct {
	Type    string   `json:"notificationType" yaml:"notificationType"`
	To      []string `json:"to" yaml:"to"`
	Context context  `json:"additionalContext" yaml:"additionalContext"`
}

type context struct {
	Body      string `json:"body" yaml:"body"`
	Formatter string `json:"formatter" yaml:"formatter"`
	Subject	  string `json:"subject" yaml:"subject"`
}

type SlackNotifier struct {
	Echo    string
	Channel string
}

func (s *SlackNotificationValues) GetWhen() []string {
	return s.When
}

func (s *SlackNotificationValues) GetLevel() string {
	return s.Level
}

func (s *SlackNotificationValues) GetAddress() string {
	return s.Address
}

func newSlackNotification(to, body string) *SlackNotification {
	// NOTE / TODO SLACK is an enum in Java.  Appears to be a "9" at the moment
	// Not sure how best to make sure this stays in sync.
	sn := &SlackNotification{Type: "SLACK"}
	sn.To = []string{to}
	sn.Context.Body = body
	sn.Context.Formatter = "TEXT"
	sn.Context.Subject = "Spinnaker Notification"
	return sn
}

// TODO: Configure channel
func NewSlackNotifier(s *settings.ExtSettings) *SlackNotifier {
	return &SlackNotifier{
		Echo:    s.Settings.Echo.BaseURL + "/notifications",
		Channel: s.Notifiers.Slack.Channel,
	}
}

func (gn *SlackNotifier) SendOnValidation() bool {
	return false
}

func (sn *SlackNotifier) SendSuccess(org, repo, path string, notificationsType plank.NotificationsType, content map[string]interface{}) {
	// If convertion to ApplicationNotification is fine then filter and send the notifications to the respective channels
	if appNotifications, errorConverting := ToNotifications(notificationsType); errorConverting == nil {
		slackNotifToSend := filterApplicationNotificationsChannels(appNotifications, []string{"pipeline.complete"})
		for _, sendNotif := range slackNotifToSend {
			sn.sendToEcho(newSlackNotification(sendNotif, fmt.Sprintf("Dinghy Successful Update: %s/%s/%s", org, repo, path)))
		}
	}
	sn.sendToEcho(newSlackNotification(sn.Channel, fmt.Sprintf("Dinghy Successful Update: %s/%s/%s", org, repo, path)))
}

func filterApplicationNotificationsChannels(notifications []ApplicationNotification, whenValues []string) []string {
	// Simulates a set, adding channels depending on whenValues
	var resultMap = make(map[string]bool)
	for _, val := range notifications {
		if notif, ok := val.(*SlackNotificationValues); ok {
			for _, when := range notif.GetWhen() {
				if contains(whenValues, when) {
					resultMap[notif.GetAddress()] = true
					break
				}
			}
		}
	}

	// Return from set just the slice with channels
	keys := make([]string, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys
}

func contains(genericSlice []string, value string) bool {
	if genericSlice == nil {
		return false
	}
	for _, val := range genericSlice {
		if val == value {
			return true
		}
	}
	return false
}

func (sn *SlackNotifier) SendFailure(org, repo, path string, errorDinghy error, notificationsType plank.NotificationsType, content map[string]interface{}) {
	// If convertion to ApplicationNotification is fine then filter and send the notifications to the respective channels
	if appNotifications, errorConverting := ToNotifications(notificationsType); errorConverting == nil {
		slackNotifToSend := filterApplicationNotificationsChannels(appNotifications, []string{"pipeline.failed"})
		for _, sendNotif := range slackNotifToSend {
			sn.sendToEcho(newSlackNotification(sendNotif, fmt.Sprintf("Dinghy Failed Update: %s/%s/%s (%s)", org, repo, path, errorDinghy.Error())))
		}
	}
	sn.sendToEcho(newSlackNotification(sn.Channel, fmt.Sprintf("Dinghy Failed Update: %s/%s/%s (%s)", org, repo, path, errorDinghy.Error())))
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
