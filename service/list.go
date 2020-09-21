package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidListOrder    = errors.New("invalid list order")
	ErrInvalidItemsPerPage = errors.New("invalid items per page")
	ErrMaxItemsPerPage     = errors.New("too many items per page request")
)

const (
	Ascending ListOrder = iota
	Descending

	DefaultOrder   = Ascending
	DefaultPerPage = 20
	MaxPerPage     = 50
)

type ListOrder int

// Ascending checks if the order is ascending.
func (o ListOrder) Ascending() bool {
	return o == Ascending
}

// Descending checks if the order is descending.
func (o ListOrder) Descending() bool {
	return o == Descending
}

// ParseListOrder parses a string a a list order.
func ParseListOrder(in string) (ListOrder, error) {
	switch strings.ToLower(in) {
	case "desc":
		return Descending, nil
	case "asc":
		return Ascending, nil
	}
	return 0, fmt.Errorf("%w: %q is not a valid list order. Valid values: asc, desc", ErrInvalidListOrder, in)
}

// ParseListOrder parses a string a a list order
// if invalid it default to the default order.
func ParseListOrderOrDefault(in string) ListOrder {
	if o, err := ParseListOrder(in); err == nil {
		return o
	}
	return Ascending
}

func ParsePerPage(in string) (int, error) {
	x, err := strconv.Atoi(in)
	if err != nil {
		return 0, fmt.Errorf("%w: must be between %d and %d", ErrInvalidItemsPerPage, 1, MaxPerPage)
	}
	if x < 0 || x > MaxPerPage {
		return 0, fmt.Errorf("%w: must be between %d and %d", ErrMaxItemsPerPage, 1, MaxPerPage)
	}

	return x, nil
}

// ParseListOrder parses a string as a "per page" list value
// if invalid it default to the default per page.
func ParsePerPageOrDefault(in string) int {
	if x, err := ParsePerPage(in); err == nil {
		return x
	}

	return DefaultPerPage
}

// ListOptions contains the options of a resource list.
type ListOptions struct {
	// FromID if provided the list of items will begin from this resource id.
	FromID *uuid.UUID
	// Order direction of the list order: either ascending or descending.
	Order ListOrder
	// PerPage maximum number of items to return.
	PerPage int
}
