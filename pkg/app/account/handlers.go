package account

import (
	"net/http"

	"github.com/gopress/internal/render"

	// "github.com/gopress/qorpress/pkg/models/users"
	"github.com/gopress/qorpress/pkg/utils"
)

// Controller posts controller
type Controller struct {
	View *render.Render
}

// Profile profile show page
func (ctrl Controller) Profile(w http.ResponseWriter, req *http.Request) {
	var (
		currentUser = utils.GetCurrentUser(req)
		// tx                              = utils.GetDB(req)
	)

	ctrl.View.Execute("profile", map[string]interface{}{
		"CurrentUser": currentUser,
	}, req, w)
}

// Update update profile page
func (ctrl Controller) Update(w http.ResponseWriter, req *http.Request) {
	// FIXME
}
