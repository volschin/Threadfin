package up2date

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	updateMetadataLimit int64 = 1024
	// 16 MiB is the smallest binary-size boundary above every supported
	// release artifact: current target builds are under 13 MB and the largest
	// published Threadfin binary is under 14 MiB.
	updateArtifactLimit int64 = 16 << 20
)

//go:embed update-signing-public-key.pem
var updateSigningPublicKeyPEM []byte

func embeddedUpdatePublicKey() (ed25519.PublicKey, error) {
	block, rest := pem.Decode(updateSigningPublicKeyPEM)
	if block == nil || len(rest) != 0 {
		return nil, fmt.Errorf("invalid embedded update public key PEM")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse embedded update public key: %w", err)
	}
	publicKey, ok := parsed.(ed25519.PublicKey)
	if !ok || len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("embedded update key is not Ed25519")
	}
	return publicKey, nil
}

func sidecarURL(artifactURL, suffix string) (string, error) {
	u, err := url.Parse(artifactURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported update URL scheme %q", u.Scheme)
	}
	u.Path += suffix
	if u.RawPath != "" {
		u.RawPath += suffix
	}
	return u.String(), nil
}

func fetchBounded(client *http.Client, requestURL string, limit int64) ([]byte, error) {
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update metadata returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("update metadata exceeds %d bytes", limit)
	}
	return body, nil
}

func fetchExpectedChecksum(client *http.Client, artifactURL string, publicKey ed25519.PublicKey) ([32]byte, error) {
	var digest [32]byte
	checksumURL, err := sidecarURL(artifactURL, ".sha256")
	if err != nil {
		return digest, err
	}
	signatureURL, err := sidecarURL(artifactURL, ".sha256.sig")
	if err != nil {
		return digest, err
	}
	checksum, err := fetchBounded(client, checksumURL, updateMetadataLimit)
	if err != nil {
		return digest, err
	}
	signature, err := fetchBounded(client, signatureURL, ed25519.SignatureSize)
	if err != nil {
		return digest, err
	}
	if len(signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, checksum, signature) {
		return digest, fmt.Errorf("invalid update checksum signature")
	}
	if len(checksum) != sha256.Size*2+1 || checksum[len(checksum)-1] != '\n' {
		return digest, fmt.Errorf("invalid signed SHA-256 checksum format")
	}
	decoded, err := hex.DecodeString(string(checksum[:len(checksum)-1]))
	if err != nil || hex.EncodeToString(decoded) != string(checksum[:len(checksum)-1]) {
		return digest, fmt.Errorf("invalid signed SHA-256 checksum")
	}
	copy(digest[:], decoded)
	return digest, nil
}

func downloadVerified(client *http.Client, artifactURL, destination string, expected [32]byte) (err error) {
	resp, err := client.Get(artifactURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update artifact returned %s", resp.Status)
	}
	if resp.ContentLength > updateArtifactLimit {
		return fmt.Errorf("update artifact exceeds %d bytes", updateArtifactLimit)
	}

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	succeeded := false
	defer func() {
		if !succeeded {
			_ = os.Remove(destination)
		}
	}()

	hash := sha256.New()
	// net/http exposes automatically decoded content through resp.Body, so the
	// limit applies to the post-decompression artifact bytes written to disk.
	written, copyErr := io.Copy(io.MultiWriter(out, hash), io.LimitReader(resp.Body, updateArtifactLimit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written > updateArtifactLimit {
		return fmt.Errorf("update artifact exceeds %d bytes", updateArtifactLimit)
	}
	if subtle.ConstantTimeCompare(hash.Sum(nil), expected[:]) != 1 {
		return fmt.Errorf("update artifact SHA-256 mismatch")
	}
	succeeded = true
	return nil
}

func prepareVerifiedUpdate(client *http.Client, artifactURL, fileType, filename, directory string, publicKey ed25519.PublicKey) (string, func() error, error) {
	emptyCleanup := func() error { return nil }
	expected, err := fetchExpectedChecksum(client, artifactURL, publicKey)
	if err != nil {
		return "", emptyCleanup, err
	}
	temporary, err := os.CreateTemp(directory, ".threadfin-update-*")
	if err != nil {
		return "", emptyCleanup, err
	}
	downloadPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(downloadPath)
		return "", emptyCleanup, err
	}
	if err := os.Remove(downloadPath); err != nil {
		return "", emptyCleanup, err
	}
	cleanupPaths := []string{downloadPath}
	cleanup := func() error {
		var cleanupErr error
		for _, cleanupPath := range cleanupPaths {
			if err := os.RemoveAll(cleanupPath); err != nil {
				cleanupErr = errors.Join(cleanupErr, fmt.Errorf("remove temporary update material: %w", err))
			}
		}
		return cleanupErr
	}
	if err := downloadVerified(client, artifactURL, downloadPath, expected); err != nil {
		return "", emptyCleanup, errors.Join(err, cleanup())
	}
	if fileType != "zip" {
		return downloadPath, cleanup, nil
	}
	extractDirectory, err := os.MkdirTemp(directory, ".threadfin-update-extract-*")
	if err != nil {
		return "", emptyCleanup, errors.Join(err, cleanup())
	}
	cleanupPaths = append(cleanupPaths, extractDirectory)
	if err := extractZIP(downloadPath, extractDirectory); err != nil {
		return "", emptyCleanup, errors.Join(err, cleanup())
	}
	candidate := filepath.Join(extractDirectory, filepath.Base(filename))
	info, err := os.Stat(candidate)
	if err != nil || !info.Mode().IsRegular() {
		candidateErr := fmt.Errorf("verified update archive does not contain %q", filepath.Base(filename))
		return "", emptyCleanup, errors.Join(candidateErr, cleanup())
	}
	return candidate, cleanup, nil
}
