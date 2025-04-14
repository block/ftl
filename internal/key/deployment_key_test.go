package key

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDeploymentKey(t *testing.T) {
	ensureDeterministicRand(t)
	for _, test := range []struct {
		keyStr      string
		expected    Deployment
		expectedErr string
	}{{
		keyStr: NewDeploymentKey("test", "time").String(),
		expected: Deployment{
			Payload: DeploymentPayload{Realm: "test", Module: "time"},
			Suffix:  "17snepfuemu5iab",
		},
	}, {
		keyStr: NewDeploymentKey("test", "time").String(),
		expected: Deployment{
			Payload: DeploymentPayload{Realm: "test", Module: "time"},
			Suffix:  "5g5cadeqxpqe574v",
		},
	}, {
		keyStr:      "-0011223344",
		expectedErr: `expected prefix "dpl" for key "-0011223344"`,
	}, {
		keyStr: NewDeploymentKey("test", "module_without_hyphens").String(),
		expected: Deployment{
			Payload: DeploymentPayload{Realm: "test", Module: "module_without_hyphens"},
			Suffix:  "59gwlv6lkyexwxf1",
		}},
	} {
		keyStr := test.keyStr
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
