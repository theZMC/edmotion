package challenge

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
)

type Catalog struct {
	mu         sync.RWMutex
	dir        string
	challenges map[string]*Challenge
	snapshot   string

	watcherAttempts atomic.Uint64
	pollAttempts    atomic.Uint64
	changedReloads  atomic.Uint64
}

type ReloadStats struct {
	WatcherAttempts uint64
	PollAttempts    uint64
	ChangedReloads  uint64
}

func NewCatalog(dir string) (*Catalog, error) {
	challenges, err := Load(dir)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]*Challenge, len(challenges))
	for _, item := range challenges {
		byID[item.ID] = item
	}

	return &Catalog{
		dir:        dir,
		challenges: byID,
		snapshot:   snapshotFor(challenges),
	}, nil
}

func (c *Catalog) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.challenges)
}

func (c *Catalog) RegisterRoutes(r chi.Router) {
	r.Get("/{challengeID}", c.fetchByID)
	r.Post("/{challengeID}", c.solveByID)
	r.Options("/{challengeID}", c.optionsByID)
}

func (c *Catalog) Stats() ReloadStats {
	return ReloadStats{
		WatcherAttempts: c.watcherAttempts.Load(),
		PollAttempts:    c.pollAttempts.Load(),
		ChangedReloads:  c.changedReloads.Load(),
	}
}

func (c *Catalog) StartAutoReload(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	watcher, err := c.newWatcher()
	if err != nil {
		slog.Warn("challenge watcher unavailable, using periodic reload only", slog.String("error", err.Error()))
	}

	var events <-chan fsnotify.Event
	var errors <-chan error
	if watcher != nil {
		defer watcher.Close()
		events = watcher.Events
		errors = watcher.Errors
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				events = nil
				continue
			}

			c.handleWatchEvent(watcher, event)
			c.reloadAndLog("watcher")
		case watchErr, ok := <-errors:
			if !ok {
				errors = nil
				continue
			}

			slog.Warn("challenge watcher error", slog.String("error", watchErr.Error()))
		case <-ticker.C:
			c.reloadAndLog("poll")
		}
	}
}

func (c *Catalog) reloadAndLog(source string) {
	attemptCount := c.bumpAttempt(source)
	stats := c.Stats()

	changed, err := c.ReloadIfChanged()
	if err != nil {
		slog.Error("reloading challenges",
			slog.String("source", source),
			slog.Uint64("attemptsForSource", attemptCount),
			slog.Uint64("watcherAttempts", stats.WatcherAttempts),
			slog.Uint64("pollAttempts", stats.PollAttempts),
			slog.Uint64("changedReloads", stats.ChangedReloads),
			slog.String("error", err.Error()),
		)
		return
	}

	if changed {
		changedTotal := c.changedReloads.Add(1)
		stats = c.Stats()
		slog.Info("reloaded challenges",
			slog.String("source", source),
			slog.Int("numChallenges", c.Len()),
			slog.Uint64("attemptsForSource", attemptCount),
			slog.Uint64("watcherAttempts", stats.WatcherAttempts),
			slog.Uint64("pollAttempts", stats.PollAttempts),
			slog.Uint64("changedReloads", changedTotal),
		)
	}
}

func (c *Catalog) bumpAttempt(source string) uint64 {
	switch source {
	case "watcher":
		return c.watcherAttempts.Add(1)
	case "poll":
		return c.pollAttempts.Add(1)
	default:
		return 0
	}
}

func (c *Catalog) newWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(c.dir); err != nil {
		watcher.Close()
		return nil, err
	}

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if err := watcher.Add(filepath.Join(c.dir, entry.Name())); err != nil {
			slog.Warn("adding watch for challenge directory",
				slog.String("path", filepath.Join(c.dir, entry.Name())),
				slog.String("error", err.Error()),
			)
		}
	}

	return watcher, nil
}

func (c *Catalog) handleWatchEvent(watcher *fsnotify.Watcher, event fsnotify.Event) {
	if watcher == nil {
		return
	}

	if event.Op&fsnotify.Create != 0 {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if err := watcher.Add(event.Name); err != nil {
				slog.Warn("adding watch for new challenge directory",
					slog.String("path", event.Name),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	if event.Op&fsnotify.Remove != 0 || event.Op&fsnotify.Rename != 0 {
		_ = watcher.Remove(event.Name)
	}
}

func (c *Catalog) ReloadIfChanged() (bool, error) {
	challenges, err := Load(c.dir)
	if err != nil {
		return false, err
	}

	nextSnapshot := snapshotFor(challenges)

	c.mu.Lock()
	defer c.mu.Unlock()

	if nextSnapshot == c.snapshot {
		return false, nil
	}

	next := make(map[string]*Challenge, len(challenges))
	for _, item := range challenges {
		next[item.ID] = item
	}

	c.challenges = next
	c.snapshot = nextSnapshot

	return true, nil
}

func (c *Catalog) lookup(id string) (*Challenge, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.challenges[id]
	return item, ok
}

func (c *Catalog) fetchByID(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "challengeID")
	item, ok := c.lookup(challengeID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	item.fetch(w, r)
}

func (c *Catalog) solveByID(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "challengeID")
	item, ok := c.lookup(challengeID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	item.solve(w, r)
}

func (c *Catalog) optionsByID(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "challengeID")
	if _, ok := c.lookup(challengeID); !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Allow", "GET, POST, OPTIONS")
}

func snapshotFor(challenges []*Challenge) string {
	builder := strings.Builder{}

	for _, item := range challenges {
		builder.WriteString(item.ID)
		builder.WriteString("|")
		builder.WriteString(item.ChallengeDir)
		builder.WriteString("|")
		builder.WriteString(strconv.Itoa(item.MaxCharacters))
		builder.WriteString("|")
		builder.WriteString(item.Flag)
		builder.WriteString("\n")
	}

	return builder.String()
}
