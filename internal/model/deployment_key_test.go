package model

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDeploymentKey(t *testing.T) {
	ensureDeterministicRand(t)
	for _, test := range []struct {
		key         DeploymentKey
		str         string
		expected    DeploymentKey
		expectedErr string
	}{
		{key: NewDeploymentKey("time"),
			expected: DeploymentKey{
				Payload: DeploymentPayload{Module: "time"},
				Suffix:  "17snepfuemu5iab",
			},
		},
		{key: NewDeploymentKey("time"),
			expected: DeploymentKey{
				Payload: DeploymentPayload{Module: "time"},
				Suffix:  "5g5cadeqxpqe574v",
			},
		},
		{str: "-0011223344", expectedErr: `expected prefix "dpl" for key "-0011223344"`},
		{key: NewDeploymentKey("module-with-hyphens"), expected: DeploymentKey{
			Payload: DeploymentPayload{Module: "module-with-hyphens"},
			Suffix:  "59gwlv6lkyexwxf1",
		},
		},
	} {
		keyStr := test.str
		if keyStr == "" {
			keyStr = test.key.String()
		}
		t.Run(keyStr, func(t *testing.T) {
			decoded, err := ParseDeploymentKey(keyStr)
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expected, decoded)
		})
	}
}

func TestZeroDeploymentKey(t *testing.T) {
	_, err := ParseDeploymentKey("")
	assert.EqualError(t, err, "expected prefix \"dpl\" for key \"\"")
}
