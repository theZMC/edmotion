package challenge

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (c *Challenge) RegisterRoutes(r chi.Router) {
	r.Get("/"+c.ID, c.fetch)
	r.Post("/"+c.ID, c.solve)
	r.Options("/"+c.ID, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", "GET, POST, OPTIONS")
	})
}

func (c *Challenge) fetch(w http.ResponseWriter, r *http.Request) {
	brokenFilePath := filepath.Join(c.ChallengeDir, brokenFileName)
	brokenData, err := os.ReadFile(brokenFilePath)
	if err != nil {
		slog.Error("reading broken file",
			slog.String("error", err.Error()),
			slog.String("path", brokenFilePath),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	serveFixedFile := ServeFixedFilesFromContext(r.Context())
	var fixedData []byte

	if serveFixedFile {
		fixedFilePath := filepath.Join(c.ChallengeDir, fixedFileName)
		fixedData, err = os.ReadFile(fixedFilePath)
		if err != nil {
			slog.Error("reading fixed file",
				slog.String("error", err.Error()),
				slog.String("path", fixedFilePath),
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Challenge ID: %s\n", c.ID)
	fmt.Fprintf(w, "Max Characters: %d\n", c.MaxCharacters)
	fmt.Fprint(w, "-----BEGIN BROKEN APPLICATION-----\n")
	_, _ = w.Write(brokenData)
	fmt.Fprint(w, "\n\n-----END BROKEN APPLICATION-----\n")

	if serveFixedFile {
		fmt.Fprint(w, "-----BEGIN FIXED APPLICATION-----\n")
		_, _ = w.Write(fixedData)
		fmt.Fprint(w, "-----END FIXED APPLICATION-----\n")
	}
}

func requestLoggerFrom(r *http.Request, challengeID string) *slog.Logger {
	reqID := middleware.GetReqID(r.Context())
	return slog.With(
		slog.String("requestID", reqID),
		slog.String("challengeID", challengeID),
	)
}
