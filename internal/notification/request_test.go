package notification

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListParamsParse(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		p := ListParams{}
		p.Parse()

		assert.Equal(t, 1, p.Page)
		assert.Equal(t, 20, p.PageSize)
		assert.Nil(t, p.BatchID)
		assert.Nil(t, p.StartDate)
		assert.Nil(t, p.EndDate)
	})

	t.Run("page size clamped", func(t *testing.T) {
		p := ListParams{PageSize: 200}
		p.Parse()

		assert.Equal(t, 20, p.PageSize)
	})

	t.Run("valid values kept", func(t *testing.T) {
		p := ListParams{Page: 3, PageSize: 50}
		p.Parse()

		assert.Equal(t, 3, p.Page)
		assert.Equal(t, 50, p.PageSize)
	})

	t.Run("batch id parsed", func(t *testing.T) {
		id := uuid.Must(uuid.NewV7())
		p := ListParams{RawBatchID: id.String()}
		p.Parse()

		require.NotNil(t, p.BatchID)
		assert.Equal(t, id, *p.BatchID)
	})

	t.Run("invalid batch id ignored", func(t *testing.T) {
		p := ListParams{RawBatchID: "not-a-uuid"}
		p.Parse()

		assert.Nil(t, p.BatchID)
	})

	t.Run("start date parsed", func(t *testing.T) {
		p := ListParams{RawStartDate: "2026-01-01T00:00:00Z"}
		p.Parse()

		require.NotNil(t, p.StartDate)
		assert.Equal(t, 2026, p.StartDate.Year())
	})

	t.Run("invalid start date ignored", func(t *testing.T) {
		p := ListParams{RawStartDate: "not-a-date"}
		p.Parse()

		assert.Nil(t, p.StartDate)
	})

	t.Run("end date parsed", func(t *testing.T) {
		p := ListParams{RawEndDate: "2026-12-31T23:59:59Z"}
		p.Parse()

		require.NotNil(t, p.EndDate)
		assert.Equal(t, 12, int(p.EndDate.Month()))
	})

	t.Run("invalid end date ignored", func(t *testing.T) {
		p := ListParams{RawEndDate: "bad"}
		p.Parse()

		assert.Nil(t, p.EndDate)
	})
}
