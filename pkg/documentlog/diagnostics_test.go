package documentlog

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestLogSizeDiagnostic(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		diagnostic := LogSizeDiagnostic{}
		assert.Equal(t, int64(0), diagnostic.sizeInBytes)
		diagnostic.add(2)
		assert.Equal(t, int64(2), diagnostic.sizeInBytes)
		diagnostic.add(3)
		assert.Equal(t, int64(5), diagnostic.sizeInBytes)
	})
	t.Run("goroutines", func(t *testing.T) {
		diagnostic := LogSizeDiagnostic{}
		var count = 20
		wg := sync.WaitGroup{}
		wg.Add(count)
		var i = 0
		for i = 0; i < count; i++ {
			go func() {
				diagnostic.add(1)
				wg.Done()
			}()
		}
		wg.Wait()
		assert.Equal(t, int64(count), diagnostic.sizeInBytes)
	})
}
