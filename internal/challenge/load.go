package challenge

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func Load(dir string) ([]*Challenge, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("opening challenges directory: %w", err)
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("challenges path is not a directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading challenges directory: %w", err)
	}

	challenges := make([]*Challenge, 0, len(entries))
	if len(entries) == 0 {
		slog.Warn("no challenge entries found in directory", slog.String("path", dir))
		return challenges, nil
	}

	slog.Debug("loading challenges from directory", slog.String("path", dir), slog.Int("numEntries", len(entries)))

	for _, entry := range entries {
		if !entry.IsDir() {
			slog.Debug("skipping non-challenge file", slog.String("path", entry.Name()))
			continue
		}

		slog.Debug("loading challenge", slog.String("path", entry.Name()))

		challengeDir := filepath.Join(dir, entry.Name())
		item, err := loadOne(challengeDir, entry.Name())
		if err != nil {
			slog.Error("skipping invalid challenge",
				slog.String("id", entry.Name()),
				slog.String("path", challengeDir),
				slog.String("error", err.Error()),
			)
			continue
		}

		challenges = append(challenges, item)
	}

	slices.SortFunc(challenges, func(a *Challenge, b *Challenge) int {
		return strings.Compare(a.ID, b.ID)
	})

	return challenges, nil
}

func loadOne(challengeDir string, id string) (*Challenge, error) {
	if _, err := os.Stat(filepath.Join(challengeDir, brokenFileName)); err != nil {
		return nil, fmt.Errorf("checking broken file: %w", err)
	}

	if _, err := os.Stat(filepath.Join(challengeDir, fixedFileName)); err != nil {
		return nil, fmt.Errorf("checking fixed file: %w", err)
	}

	maxData, err := os.ReadFile(filepath.Join(challengeDir, maxFileName))
	if err != nil {
		return nil, fmt.Errorf("reading max file: %w", err)
	}

	maxChars, err := strconv.Atoi(strings.TrimSpace(string(maxData)))
	if err != nil {
		return nil, fmt.Errorf("parsing max characters: %w", err)
	}

	flagData, err := os.ReadFile(filepath.Join(challengeDir, flagFileName))
	if err != nil {
		return nil, fmt.Errorf("reading flag file: %w", err)
	}

	item := &Challenge{
		ID:            id,
		MaxCharacters: maxChars,
		Flag:          strings.TrimSpace(string(flagData)),
		ChallengeDir:  challengeDir,
	}

	slog.Debug("loaded challenge",
		slog.String("id", item.ID),
		slog.Int("maxCharacters", item.MaxCharacters),
		slog.String("flag", item.Flag),
	)

	return item, nil
}
