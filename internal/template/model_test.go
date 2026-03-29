package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderBody(t *testing.T) {
	t.Run("all variables resolved", func(t *testing.T) {
		tmpl := &Template{Body: "Hello {{name}}, welcome to {{company}}"}

		body, err := tmpl.RenderBody(map[string]string{"name": "Semih", "company": "Insider"})

		require.NoError(t, err)
		assert.Equal(t, "Hello Semih, welcome to Insider", body)
	})

	t.Run("no variables in body", func(t *testing.T) {
		tmpl := &Template{Body: "Hello world"}

		body, err := tmpl.RenderBody(nil)

		require.NoError(t, err)
		assert.Equal(t, "Hello world", body)
	})

	t.Run("unresolved placeholder returns error", func(t *testing.T) {
		tmpl := &Template{Body: "Hello {{name}}, code is {{code}}"}

		_, err := tmpl.RenderBody(map[string]string{"name": "Semih"})

		assert.ErrorIs(t, err, ErrUnresolvedPlaceholder)
		assert.Contains(t, err.Error(), "{{code}}")
	})

	t.Run("multiple unresolved placeholders", func(t *testing.T) {
		tmpl := &Template{Body: "{{greeting}} {{name}}, {{message}}"}

		_, err := tmpl.RenderBody(nil)

		assert.ErrorIs(t, err, ErrUnresolvedPlaceholder)
		assert.Contains(t, err.Error(), "{{greeting}}")
		assert.Contains(t, err.Error(), "{{name}}")
		assert.Contains(t, err.Error(), "{{message}}")
	})

	t.Run("empty vars with no placeholders", func(t *testing.T) {
		tmpl := &Template{Body: "Plain text"}

		body, err := tmpl.RenderBody(map[string]string{})

		require.NoError(t, err)
		assert.Equal(t, "Plain text", body)
	})

	t.Run("extra vars ignored", func(t *testing.T) {
		tmpl := &Template{Body: "Hello {{name}}"}

		body, err := tmpl.RenderBody(map[string]string{"name": "Semih", "unused": "value"})

		require.NoError(t, err)
		assert.Equal(t, "Hello Semih", body)
	})
}
