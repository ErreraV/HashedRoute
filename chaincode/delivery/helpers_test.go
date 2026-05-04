package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShipmentKey(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{"normal id", "SHP-1001", "SHM:SHP-1001"},
		{"empty id", "", "SHM:"},
		{"id with spaces", "  SHP-1001  ", "SHM:  SHP-1001  "},
		{"special chars", "a/b?c=d&e", "SHM:a/b?c=d&e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shipmentKey(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trims leading spaces", "  SHP-1001", "SHP-1001"},
		{"trims trailing spaces", "SHP-1001  ", "SHP-1001"},
		{"trims both sides", "  SHP-1001  ", "SHP-1001"},
		{"trims tabs and newlines", "\t\n SHP-1001 \n\t", "SHP-1001"},
		{"empty string", "", ""},
		{"whitespace only", "   \t \n  ", ""},
		{"already clean", "SHP-1001", "SHP-1001"},
		{"single char", "x", "x"},
		{"unicode space trimmed", "\u00a0SHP-1001", "SHP-1001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeID(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateStatusValue(t *testing.T) {
	t.Run("valid statuses", func(t *testing.T) {
		for _, s := range []string{StatusCreated, StatusPickedUp, StatusInTransit, StatusDelivered} {
			t.Run(s, func(t *testing.T) {
				err := validateStatusValue(s)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("invalid statuses", func(t *testing.T) {
		invalid := []string{"", "PENDING", "delivered", "created", "IN_TRANSIT_2", "x"}
		for _, s := range invalid {
			t.Run(s, func(t *testing.T) {
				err := validateStatusValue(s)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid status")
			})
		}
	})
}

func TestAssertValidTransition(t *testing.T) {
	t.Run("valid transitions", func(t *testing.T) {
		valid := [][2]string{
			{StatusCreated, StatusPickedUp},
			{StatusPickedUp, StatusInTransit},
			{StatusInTransit, StatusDelivered},
		}
		for _, pair := range valid {
			from, to := pair[0], pair[1]
			t.Run(from+"→"+to, func(t *testing.T) {
				err := assertValidTransition(from, to)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("same status", func(t *testing.T) {
		err := assertValidTransition(StatusCreated, StatusCreated)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already CREATED")

		err = assertValidTransition(StatusPickedUp, StatusPickedUp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already PICKED_UP")
	})

	t.Run("skip ahead", func(t *testing.T) {
		invalid := [][2]string{
			{StatusCreated, StatusInTransit},
			{StatusCreated, StatusDelivered},
			{StatusPickedUp, StatusDelivered},
		}
		for _, pair := range invalid {
			from, to := pair[0], pair[1]
			t.Run(from+"→"+to, func(t *testing.T) {
				err := assertValidTransition(from, to)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot go from")
			})
		}
	})

	t.Run("backwards", func(t *testing.T) {
		invalid := [][2]string{
			{StatusPickedUp, StatusCreated},
			{StatusInTransit, StatusPickedUp},
			{StatusDelivered, StatusInTransit},
			{StatusDelivered, StatusCreated},
		}
		for _, pair := range invalid {
			from, to := pair[0], pair[1]
			t.Run(from+"→"+to, func(t *testing.T) {
				err := assertValidTransition(from, to)
				assert.Error(t, err)
			})
		}
	})

	t.Run("from delivered", func(t *testing.T) {
		err := assertValidTransition(StatusDelivered, StatusCreated)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already delivered")

		err = assertValidTransition(StatusDelivered, StatusPickedUp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already delivered")

		err = assertValidTransition(StatusDelivered, StatusInTransit)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already delivered")

		err = assertValidTransition(StatusDelivered, StatusDelivered)
		// Delivered→Delivered hits from==to check first → "status is already DELIVERED"
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already DELIVERED")
	})

	t.Run("unknown from status", func(t *testing.T) {
		err := assertValidTransition("BOGUS", StatusCreated)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown current status")
	})
}
