package challenge

import (
	"context"
	"os"
	"os/exec"
	"time"
)

type contextKey struct{}

var (
	serveFixedFilesKey = contextKey{}
	vimControlChars    = map[string]rune{
		"<TAB>": 0x9,
		"<ESC>": 0x1b,
		"<C-V>": 0x16,
	}
	allowedControlChars = map[rune]struct{}{
		0x9:  {},
		0x1b: {},
		0x16: {},
	}
	vimPath = discoverVimPath()
)

const (
	maxSolutionBodyBytes = 4 * 1024
	vimExecutionTimeout  = 3 * time.Second
	brokenFileName       = "broken"
	fixedFileName        = "fixed"
	flagFileName         = "flag"
	maxFileName          = "max"
)

type Challenge struct {
	ID            string
	MaxCharacters int
	Flag          string
	ChallengeDir  string
}

func ContextWithFixedFiles(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, serveFixedFilesKey, enabled)
}

func ServeFixedFilesFromContext(ctx context.Context) bool {
	return ctx.Value(serveFixedFilesKey) == true
}

func discoverVimPath() string {
	path := os.Getenv("VIM_PATH")
	if path != "" {
		return path
	}

	path, err := exec.LookPath("vim")
	if err != nil {
		return "vim"
	}

	return path
}
