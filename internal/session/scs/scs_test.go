package scs_test

import (
	"testing"

	scssession "github.com/alexedwards/scs"
	"github.com/alexedwards/scs/stores/memstore"

	"github.com/qorpress/qorpress/internal/session/scs"
	"github.com/qorpress/qorpress/internal/session/test"
)

func TestAll(t *testing.T) {
	engine := scssession.NewManager(memstore.New(0))
	manager := scs.New(engine)
	test.TestAll(manager, t)
}
