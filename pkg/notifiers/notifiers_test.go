package notifiers

import (
	"encoding/json"
	"github.com/armory/plank/v3"
	"testing"
)

func TestToNotifications(t *testing.T) {
	input := `
			{
    "application": "decktesting",
    "slack": [
        {
            "level": "application",
            "when": [
                "pipeline.starting",
                "pipeline.complete",
                "pipeline.failed"
            ],
            "type": "slack",
            "address": "jossuecito-dinghy-pipes"
        },
{
            "level": "application",
            "when": [
                "pipeline.starting",
                "pipeline.complete",
                "pipeline.failed"
            ],
            "type": "slack",
            "address": "jossuecito-dinghy-pipes-22222"
        }
    ],
    "lastModifiedBy": "anonymous",
    "lastModified": 1591050297150,
    "email": [
        {
            "level": "application",
            "when": [
                "pipeline.starting"
            ],
            "type": "email",
            "address": "test@test.com",
            "cc": "test@test2.com"
        }
    ]
}
	`

	notifications := make(plank.NotificationsType)
	json.Unmarshal([]byte(input), &notifications)
	if toNotif , err := ToNotifications(notifications); err != nil || len(toNotif) != 2 {
		t.Error("notifications were not parsed correctly")
	}
}