package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/thezmc/edmotion/internal/admin"
	"github.com/thezmc/edmotion/internal/challenge"
	"github.com/thezmc/edmotion/internal/config"
	"github.com/thezmc/edmotion/internal/httpapi"
	"github.com/thezmc/edmotion/internal/logging"
)

type Runtime struct {
	settings config.Settings
}

func New(settings config.Settings) *Runtime {
	return &Runtime{settings: settings}
}

func (r *Runtime) Run() error {
	logger := logging.New(r.settings.LogLevel)
	slog.SetDefault(logger)

	adminState, generatedPassword, err := admin.NewState(r.settings.AdminPassword, r.settings.GiveFixedFiles)
	if err != nil {
		return err
	}

	if generatedPassword != "" {
		slog.Warn("no admin password set, generated random password", slog.String("adminPassword", generatedPassword))
	}

	catalog, err := challenge.NewCatalog(r.settings.ChallengeDir)
	if err != nil {
		return fmt.Errorf("loading challenges: %w", err)
	}

	slog.Info("loaded challenges", slog.Int("numChallenges", catalog.Len()))

	reloadCtx, cancelReload := context.WithCancel(context.Background())
	defer cancelReload()
	go catalog.StartAutoReload(reloadCtx, r.settings.CatalogReloadInterval)

	router, err := httpapi.NewRouter(httpapi.Params{
		Catalog:               catalog,
		Admin:                 adminState,
		RequestLimitPerMinute: r.settings.RequestLimitPerMinute,
	})
	if err != nil {
		return fmt.Errorf("creating router: %w", err)
	}

	slog.Info("starting server", slog.String("addr", r.settings.HTTPAddr))
	if err := http.ListenAndServe(r.settings.HTTPAddr, router); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
