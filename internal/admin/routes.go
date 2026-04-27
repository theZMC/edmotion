package admin

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, state *State) {
	r.Put("/admin/toggle-fixed-files", func(w http.ResponseWriter, r *http.Request) {
		if !state.IsAuthorized(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		current := state.ToggleFixedFiles()
		slog.Info("toggled fixed files", slog.Bool("giveFixedFiles", current))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "giveFixedFiles set to %v", current)
	})

	r.Put("/admin/set-password", func(w http.ResponseWriter, r *http.Request) {
		if !state.IsAuthorized(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		newPassword := r.FormValue("password")
		if newPassword == "" {
			http.Error(w, "Password cannot be empty", http.StatusBadRequest)
			return
		}

		state.SetPassword(newPassword)
		slog.Info("admin password updated")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Admin password updated successfully")
	})
}
