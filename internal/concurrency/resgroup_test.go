package concurrency

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
)

func TestResourceGroup(t *testing.T) {
	t.Run("successful concurrent execution", func(t *testing.T) {
		rg := &ResourceGroup[int]{}

		rg.Go(func() (int, error) {
			return 1, nil
		})

		rg.Go(func() (int, error) {
			return 2, nil
		})

		rg.Go(func() (int, error) {
			return 3, nil
		})

		results, err := rg.Wait()
		assert.NoError(t, err)

		actualSum := 0
		for _, v := range results {
			actualSum += v
		}
		assert.Equal(t, 6, actualSum)
	})

	t.Run("handles errors", func(t *testing.T) {
		rg := &ResourceGroup[string]{}
		expectedErr := errors.New("test error")

		rg.Go(func() (string, error) {
			return "", errors.WithStack(expectedErr)
		})

		rg.Go(func() (string, error) {
			return "success", nil
		})

		results, err := rg.Wait()
		assert.EqualError(t, err, expectedErr.Error())
		assert.Equal(t, results, nil)
	})
}
