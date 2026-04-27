package challenge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Challenge) solve(w http.ResponseWriter, r *http.Request) {
	log := requestLoggerFrom(r, c.ID)

	r.Body = http.MaxBytesReader(w, r.Body, maxSolutionBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "Solution exceeds maximum request size", http.StatusRequestEntityTooLarge)
			return
		}

		log.Error("reading request body", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	solution := string(body)
	for str, ch := range vimControlChars {
		solution = strings.ReplaceAll(solution, str, string(ch))
	}

	if len(solution) > c.MaxCharacters {
		http.Error(w, fmt.Sprintf("Solution exceeds maximum character limit of %d", c.MaxCharacters), http.StatusBadRequest)
		return
	}

	if err := validateSolutionScript(solution); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	solutionFile, err := os.CreateTemp("", "solution-*.txt")
	if err != nil {
		log.Error("creating temp file for solution", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(solutionFile.Name())

	if _, err := solutionFile.WriteString(solution); err != nil {
		log.Error("writing solution to temp file", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := solutionFile.WriteString("\x1bZZ"); err != nil {
		log.Error("writing save and quit keys to solution file", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := solutionFile.Close(); err != nil {
		log.Error("closing solution file", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	brokenFilePath := filepath.Join(c.ChallengeDir, brokenFileName)
	brokenData, err := os.ReadFile(brokenFilePath)
	if err != nil {
		log.Error("reading broken file",
			slog.String("error", err.Error()),
			slog.String("path", brokenFilePath),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetFile, err := os.CreateTemp("", c.ID+"-*.py")
	if err != nil {
		log.Error("creating temp file for challenge source", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(targetFile.Name())

	if _, err := targetFile.Write(brokenData); err != nil {
		log.Error("writing challenge source to temp file", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := targetFile.Close(); err != nil {
		log.Error("closing challenge source temp file", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), vimExecutionTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		vimPath,
		"-n",
		"-u", "NONE",
		"-i", "NONE",
		"-N",
		"-Z",
		"--not-a-term",
		"-s", solutionFile.Name(),
		targetFile.Name(),
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error("getting stderr pipe", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("getting stdout pipe", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Error("starting command", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	outputBytes, err := io.ReadAll(stdout)
	if err != nil {
		log.Error("reading stdout", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log = log.With(slog.String("stdout", string(outputBytes)))

	errorBytes, err := io.ReadAll(stderr)
	if err != nil {
		log.Error("reading stderr", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log = log.With(slog.String("stderr", string(errorBytes)))

	if err := cmd.Wait(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Error("command execution timeout")
			http.Error(w, "Solution execution timed out", http.StatusBadRequest)
			return
		}

		log.Error("command execution error", slog.String("error", err.Error()))
		http.Error(w, "Incorrect solution", http.StatusBadRequest)
		return
	}

	editedData, err := os.ReadFile(targetFile.Name())
	if err != nil {
		log.Error("reading edited challenge source", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fixedFilePath := filepath.Join(c.ChallengeDir, fixedFileName)
	fixedData, err := os.ReadFile(fixedFilePath)
	if err != nil {
		log.Error("reading fixed file",
			slog.String("error", err.Error()),
			slog.String("path", fixedFilePath),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !bytes.Equal(editedData, fixedData) {
		normalizedEdited := normalizePythonSource(editedData)
		normalizedFixed := normalizePythonSource(fixedData)
		if !bytes.Equal(normalizedEdited, normalizedFixed) {
			http.Error(w, "Incorrect solution", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, c.Flag)
}

func normalizePythonSource(data []byte) []byte {
	text := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	lines := bytes.Split(text, []byte("\n"))

	for i := range lines {
		lines[i] = bytes.TrimRight(lines[i], " \t")
	}

	normalized := bytes.Join(lines, []byte("\n"))
	normalized = bytes.TrimRight(normalized, "\n")

	return normalized
}

func validateSolutionScript(solution string) error {
	if len(solution) == 0 {
		return fmt.Errorf("solution cannot be empty")
	}

	for _, ch := range solution {
		if ch < 0x20 {
			if _, ok := allowedControlChars[ch]; !ok {
				return fmt.Errorf("solution contains disallowed control characters")
			}
		}

		if ch == 0x7f {
			return fmt.Errorf("solution contains disallowed control characters")
		}
	}

	return nil
}
