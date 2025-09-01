package models

import (
	"backend/internal/db"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {

	// setting up database
	db.InitDB("dev")

	// setting up the flags
	f := &Flags{}

	err := f.Get()
	assert.NoError(t, err, err)

	err = f.Set("test", true)
	assert.NoError(t, err, err)

	err = f.Unset("test")
	assert.NoError(t, err, err)
}
