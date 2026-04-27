package challenge_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/thezmc/edmotion/internal/challenge"
)

func writeFixture(path string, body string) {
	Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
}

func writeChallengeBundle(dir string, id string) {
	challengeID := id
	if challengeID == "" {
		challengeID = "derived-id"
	}

	challengeDir := filepath.Join(dir, challengeID)
	Expect(os.Mkdir(challengeDir, 0o755)).To(Succeed())

	writeFixture(filepath.Join(challengeDir, "broken"), "print('broken')\n")
	writeFixture(filepath.Join(challengeDir, "fixed"), "print('fixed')\n")
	writeFixture(filepath.Join(challengeDir, "max"), "6\n")
	writeFixture(filepath.Join(challengeDir, "flag"), fmt.Sprintf("FLAG{%s}\n", challengeID))
}

var _ = Describe("Challenge integration", Label("integration"), func() {
	Describe("RegisterRoutes", func() {
		var (
			item   *challenge.Challenge
			router chi.Router
			tmpDir string
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "challenge-routes-*")
			Expect(err).NotTo(HaveOccurred())

			writeFixture(filepath.Join(tmpDir, "broken"), "print('broken')\n")
			writeFixture(filepath.Join(tmpDir, "fixed"), "print('fixed')\n")

			item = &challenge.Challenge{
				ID:            "signal-intercept",
				MaxCharacters: 20,
				Flag:          "FLAG{test}",
				ChallengeDir:  tmpDir,
			}

			router = chi.NewRouter()
			item.RegisterRoutes(router)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("serves OPTIONS with explicit allow header", func() {
			req := httptest.NewRequest(http.MethodOptions, "/signal-intercept", nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Header().Get("Allow")).To(Equal("GET, POST, OPTIONS"))
		})

		It("returns challenge metadata and broken source on GET", func() {
			req := httptest.NewRequest(http.MethodGet, "/signal-intercept", nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Header().Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
			Expect(rr.Body.String()).To(ContainSubstring("Challenge ID: signal-intercept"))
			Expect(rr.Body.String()).To(ContainSubstring("Max Characters: 20"))
			Expect(rr.Body.String()).To(ContainSubstring("-----BEGIN BROKEN APPLICATION-----"))
			Expect(rr.Body.String()).To(ContainSubstring("print('broken')"))
			Expect(rr.Body.String()).To(ContainSubstring("-----END BROKEN APPLICATION-----"))
		})

		It("includes fixed source when context enables it", func() {
			req := httptest.NewRequest(http.MethodGet, "/signal-intercept", nil)
			req = req.WithContext(challenge.ContextWithFixedFiles(req.Context(), true))
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(ContainSubstring("-----BEGIN FIXED APPLICATION-----"))
			Expect(rr.Body.String()).To(ContainSubstring("print('fixed')"))
			Expect(rr.Body.String()).To(ContainSubstring("-----END FIXED APPLICATION-----"))
		})

		It("returns 500 when broken source file is missing", func() {
			Expect(os.Remove(filepath.Join(tmpDir, "broken"))).To(Succeed())

			req := httptest.NewRequest(http.MethodGet, "/signal-intercept", nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
		})

		It("returns 500 when fixed source is missing and fixed output is enabled", func() {
			Expect(os.Remove(filepath.Join(tmpDir, "fixed"))).To(Succeed())

			req := httptest.NewRequest(http.MethodGet, "/signal-intercept", nil)
			req = req.WithContext(challenge.ContextWithFixedFiles(req.Context(), true))
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
		})

		It("accepts a correct edit script in POST", func() {
			body := strings.NewReader("fbcwfixed\x1b")
			req := httptest.NewRequest(http.MethodPost, "/signal-intercept", body)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(ContainSubstring("FLAG"))
		})

		It("accepts a correct edit script when differences are trailing whitespace only", func() {
			body := strings.NewReader("fbcwfixed\x1bA \x1b")
			req := httptest.NewRequest(http.MethodPost, "/signal-intercept", body)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(ContainSubstring("FLAG"))
		})

		It("rejects incorrect edit scripts in POST", func() {
			body := strings.NewReader("ix\x1b")
			req := httptest.NewRequest(http.MethodPost, "/signal-intercept", body)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
		})

		It("rejects disallowed command mode keys in POST", func() {
			body := bytes.NewReader([]byte{0x19})
			req := httptest.NewRequest(http.MethodPost, "/signal-intercept", body)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(ContainSubstring("contains disallowed control characters"))
		})

		It("rejects empty scripts in POST", func() {
			body := strings.NewReader("")
			req := httptest.NewRequest(http.MethodPost, "/signal-intercept", body)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(ContainSubstring("cannot be empty"))
		})

		It("rejects oversized edit scripts in POST", func() {
			body := strings.NewReader("abcdefghijklmnopqrstuvwxyz")
			req := httptest.NewRequest(http.MethodPost, "/signal-intercept", body)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(ContainSubstring("maximum character limit"))
		})
	})

	Describe("Load", func() {
		It("returns an error when the directory does not exist", func() {
			_, err := challenge.Load(filepath.Join(os.TempDir(), "missing-challenge-dir-does-not-exist"))
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the path is not a directory", func() {
			tmpFile, err := os.CreateTemp("", "challenge-path-*")
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpFile.Close()).To(Succeed())
			defer os.Remove(tmpFile.Name())

			_, err = challenge.Load(tmpFile.Name())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a directory"))
		})

		It("returns an empty slice for an empty challenge directory", func() {
			tmpDir, err := os.MkdirTemp("", "challenge-empty-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tmpDir)

			challenges, loadErr := challenge.Load(tmpDir)
			Expect(loadErr).NotTo(HaveOccurred())
			Expect(challenges).To(BeEmpty())
		})

		It("loads only challenge directories and sets ChallengeDir", func() {
			tmpDir, err := os.MkdirTemp("", "challenge-load-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tmpDir)

			writeChallengeBundle(tmpDir, "auth-bypass")
			writeChallengeBundle(tmpDir, "")
			writeFixture(filepath.Join(tmpDir, "notes.yaml"), "id: ignore-me\n")

			challenges, loadErr := challenge.Load(tmpDir)
			Expect(loadErr).NotTo(HaveOccurred())
			Expect(challenges).To(HaveLen(2))

			ids := make([]string, 0, len(challenges))
			for _, c := range challenges {
				ids = append(ids, c.ID)
				Expect(c.ChallengeDir).To(Equal(filepath.Join(tmpDir, c.ID)))
			}

			Expect(ids).To(ConsistOf("auth-bypass", "derived-id"))
		})

		It("skips invalid challenge directories instead of returning an error", func() {
			tmpDir, err := os.MkdirTemp("", "challenge-invalid-max-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tmpDir)

			writeChallengeBundle(tmpDir, "good")
			writeChallengeBundle(tmpDir, "bad")
			writeFixture(filepath.Join(tmpDir, "bad", "max"), "not-a-number")

			challenges, loadErr := challenge.Load(tmpDir)
			Expect(loadErr).NotTo(HaveOccurred())
			Expect(challenges).To(HaveLen(1))
			Expect(challenges[0].ID).To(Equal("good"))
		})
	})
})
