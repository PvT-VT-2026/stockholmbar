package handlers

import (
	"db-client/internal/stores"
)

type UnitHandler struct {
	store *stores.UnitStore
}

func NewUnitHandler(s *stores.UnitStore) *UnitHandler {
	return &UnitHandler{store: s}
}

// UnitHandler no longer owns its own insertion endpoint,
// as all insertions go through the submission system.
// This handler will be populated by getters only