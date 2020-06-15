package notifiers

import (
	"encoding/json"
	"github.com/armory/plank/v3"
	"testing"
)

func Test_filterApplicationNotificationsChannels(t *testing.T) {
	tests := []struct {
		name string
		input string
		expected []string
		pipesFilter []string
	}{
		{
			name: "filter_one_pipeline",
			input: `
			{
    "application": "decktesting",
    "slack": [
        {
            "level": "application",
            "when": [
                "pipeline.starting",
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
	`,
			expected:  []string{"jossuecito-dinghy-pipes-22222"},
			pipesFilter: []string{"pipeline.complete"},
		},
		{
			name: "filter_two_pipelines_same_pipe",
			input: `
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
	`,
			expected:  []string{"jossuecito-dinghy-pipes-22222", "jossuecito-dinghy-pipes"},
			pipesFilter: []string{"pipeline.complete"},
		},
		{
			name: "filter_two_pipelines_different_args",
			input: `
			{
    "application": "decktesting",
    "slack": [
        {
            "level": "application",
            "when": [
                "pipeline.starting",
                "pipeline.failed"
            ],
            "type": "slack",
            "address": "jossuecito-dinghy-pipes"
        },
{
            "level": "application",
            "when": [
                "pipeline.starting",
                "pipeline.complete"
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
	`,
			expected:  []string{"jossuecito-dinghy-pipes", "jossuecito-dinghy-pipes-22222"},
			pipesFilter: []string{"pipeline.failed", "pipeline.complete"},
		},
		{
			name: "filter_one_pipeline_two_args",
			input: `
			{
    "application": "decktesting",
    "slack": [
        {
            "level": "application",
            "when": [
                "pipeline.starting",
                "pipeline.failed"
            ],
            "type": "slack",
            "address": "jossuecito-dinghy-pipes"
        },
{
            "level": "application",
            "when": [
                "pipeline.complete"
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
	`,
			expected:  []string{"jossuecito-dinghy-pipes"},
			pipesFilter: []string{"pipeline.failed", "pipeline.starting"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			notifications := make(plank.NotificationsType)
			json.Unmarshal([]byte(tt.input), &notifications)
			notifInter, _ := ToNotifications(notifications)

			got := filterApplicationNotificationsChannels(notifInter, tt.pipesFilter)

			var found int

			// Sometimes position is different so we will search them
			for _, val := range got {
				for _, exp := range tt.expected {
					if val == exp {
						found += 1
					}
				}
			}

			if found != len(tt.expected) || len(got) != len(tt.expected) {
				t.Errorf("filterApplicationNotificationsChannels() = %v, want %v", got, tt.expected)

			}

		})
	}
}