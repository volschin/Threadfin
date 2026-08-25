package up2date

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func signedUpdateServer(t *testing.T, artifact []byte, checksum []byte, signature []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Threadfin_linux_amd64":
			_, _ = w.Write(artifact)
		case "/Threadfin_linux_amd64.sha256":
			_, _ = w.Write(checksum)
		case "/Threadfin_linux_amd64.sha256.sig":
			_, _ = w.Write(signature)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSidecarURLPreservesQueryAndAddsSuffixToPath(t *testing.T) {
	got, err := sidecarURL("https://updates.example/Threadfin_linux_amd64?token=abc", ".sha256")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://updates.example/Threadfin_linux_amd64.sha256?token=abc"
	if got != want {
		t.Fatalf("sidecar URL = %q, want %q", got, want)
	}
}

func TestSidecarURLPreservesPercentEscapedPath(t *testing.T) {
	got, err := sidecarURL("https://updates.example/releases/a%2Fb/Threadfin", ".sha256")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://updates.example/releases/a%2Fb/Threadfin.sha256"
	if got != want {
		t.Fatalf("sidecar URL = %q, want %q", got, want)
	}
}

func TestFetchExpectedChecksumVerifiesSignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("verified update"))
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, []byte("verified update"), checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()

	got, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey)
	if err != nil {
		t.Fatal(err)
	}
	if got != digest {
		t.Fatalf("digest = %x, want %x", got, digest)
	}
}

func TestFetchExpectedChecksumRejectsInvalidSignature(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("update"))
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, []byte("update"), checksum, make([]byte, ed25519.SignatureSize))
	defer server.Close()

	if _, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey); err == nil {
		t.Fatal("invalid signature was accepted")
	}
}

func TestFetchExpectedChecksumRejectsMissingSignature(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("update"))
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Threadfin_linux_amd64.sha256" {
			_, _ = w.Write(checksum)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	if _, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey); err == nil {
		t.Fatal("missing signature was accepted")
	}
}

func TestFetchExpectedChecksumRejectsSignedMalformedChecksum(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	checksum := []byte("not-a-sha256\n")
	server := signedUpdateServer(t, []byte("update"), checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()

	if _, err := fetchExpectedChecksum(server.Client(), server.URL+"/Threadfin_linux_amd64", publicKey); err == nil {
		t.Fatal("signed malformed checksum was accepted")
	}
}

func TestDownloadVerifiedWritesMatchingArtifact(t *testing.T) {
	artifact := []byte("verified update")
	digest := sha256.Sum256(artifact)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(artifact)
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "candidate")
	if err := downloadVerified(server.Client(), server.URL, destination, digest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(artifact) {
		t.Fatalf("candidate = %q, want %q", got, artifact)
	}
}

func TestDownloadVerifiedRemovesChecksumMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "tampered")
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "candidate")
	want := sha256.Sum256([]byte("trusted"))
	if err := downloadVerified(server.Client(), server.URL, destination, want); err == nil {
		t.Fatal("checksum mismatch was accepted")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("unverified candidate remains: %v", err)
	}
}

func TestPrepareVerifiedUpdateReturnsCandidateWithoutTouchingCurrentBinary(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	artifact := []byte("new executable")
	digest := sha256.Sum256(artifact)
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, artifact, checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()
	directory := t.TempDir()
	current := filepath.Join(directory, "threadfin")
	if err := os.WriteFile(current, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}

	candidate, cleanup, err := prepareVerifiedUpdate(server.Client(), server.URL+"/Threadfin_linux_amd64", "bin", "", directory, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	gotCurrent, _ := os.ReadFile(current)
	if string(gotCurrent) != "current executable" {
		t.Fatalf("current executable changed during verification: %q", gotCurrent)
	}
	gotCandidate, _ := os.ReadFile(candidate)
	if string(gotCandidate) != string(artifact) {
		t.Fatalf("candidate = %q, want %q", gotCandidate, artifact)
	}
}

func TestPrepareVerifiedUpdateMismatchLeavesNoCandidate(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	trusted := sha256.Sum256([]byte("trusted"))
	checksum := []byte(hex.EncodeToString(trusted[:]) + "\n")
	server := signedUpdateServer(t, []byte("tampered"), checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()
	directory := t.TempDir()

	if _, _, err := prepareVerifiedUpdate(server.Client(), server.URL+"/Threadfin_linux_amd64", "bin", "", directory, publicKey); err == nil {
		t.Fatal("tampered update was prepared")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary update material remains: %v", entries)
	}
}

func TestPrepareVerifiedZipReturnsExpectedBinary(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var archiveBody bytes.Buffer
	archiveWriter := zip.NewWriter(&archiveBody)
	entry, err := archiveWriter.Create("Threadfin_linux_amd64")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("verified zipped executable")); err != nil {
		t.Fatal(err)
	}
	if err := archiveWriter.Close(); err != nil {
		t.Fatal(err)
	}
	artifact := archiveBody.Bytes()
	digest := sha256.Sum256(artifact)
	checksum := []byte(hex.EncodeToString(digest[:]) + "\n")
	server := signedUpdateServer(t, artifact, checksum, ed25519.Sign(privateKey, checksum))
	defer server.Close()

	candidate, cleanup, err := prepareVerifiedUpdate(server.Client(), server.URL+"/Threadfin_linux_amd64", "zip", "Threadfin_linux_amd64", t.TempDir(), publicKey)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	got, err := os.ReadFile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "verified zipped executable" {
		t.Fatalf("zipped candidate = %q", got)
	}
}
