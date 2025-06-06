// This is the echo module.
package naughty

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

// BeNaughty attempts to ping echo directly and returns true if successful
//
//ftl:verb export
func BeNaughty(ctx context.Context, endpoint map[string]string) (string, error) {
	url := "http://" + endpoint["name"] + ":8893/_readiness" // Replace with your actual URL

	resp, err := http.Get(url)
	if err != nil {
		return "false", nil
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return "false", nil
	}
	return strconv.FormatBool(resp.StatusCode == 200), nil
}
