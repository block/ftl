// This is the echo module.
package echo

import (
	"context"
	"fmt"
	"strings"
)

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req string, db EchodbHandle) (string, error) {
	_, err := db.Get(ctx).Exec(`INSERT INTO messages (message) VALUES ($1);`, req)
	if err != nil {
		return "", err
	}

	rows, err := db.Get(ctx).Query(`SELECT DISTINCT message FROM messages;`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var message string
		err = rows.Scan(&message)
		if err != nil {
			return "", err
		}
		messages = append(messages, message)
	}
	return fmt.Sprintf("Hello, %s!!!", strings.Join(messages, ",")), nil
}
