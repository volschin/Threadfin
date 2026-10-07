package imgcache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	imageResponseLimit = 10 << 20
	imageTempPrefix    = ".threadfin-image-"
)

// Cache : Cache strcut
type Cache struct {
	path     string
	cacheURL string
	caching  bool
	images   map[string]string
	client   *http.Client
	cacheMu  sync.Mutex
	Queue    []string
	Cache    []string
	Image    imageFunc
	sync.RWMutex
}

type imageFunc struct {
	GetURL  func(string, string, string, bool, int, string) string
	Caching func()
	Remove  func()
}

// New : New cahce
func (c *Cache) cacheImage(src string) (filteredSource string, cached bool) {
	resp, err := c.client.Get(src)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK || resp.ContentLength > imageResponseLimit {
		return "", false
	}

	filteredSource = strings.Split(src, "?")[0]
	u, err := url.Parse(filteredSource)
	if err != nil {
		return "", false
	}
	filename := strToMD5(filteredSource) + filepath.Ext(u.Path)
	file, err := os.CreateTemp(c.path, imageTempPrefix+"*")
	if err != nil {
		return "", false
	}
	defer os.Remove(file.Name())

	written, copyErr := io.Copy(file, io.LimitReader(resp.Body, imageResponseLimit+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || written > imageResponseLimit {
		return "", false
	}

	// Only publish a closed, complete image. Cleanup and URL lookups can run
	// while the network transfer is in progress.
	c.Lock()
	defer c.Unlock()
	if err := os.Rename(file.Name(), filepath.Join(c.path, filename)); err != nil {
		return "", false
	}
	c.images[filename] = c.cacheURL + filename

	return filteredSource, true
}

func New(path, cacheURL string, caching bool) (c *Cache, err error) {

	c = &Cache{}

	c.images = make(map[string]string)
	c.client = &http.Client{Timeout: 10 * time.Second}
	c.path = path
	c.cacheURL = cacheURL
	c.caching = caching
	c.Queue = []string{}
	c.Cache = []string{}

	c.Image.GetURL = func(src string, http_domain string, http_port string, force_https bool, https_port int, https_domain string) (cacheURL string) {

		c.Lock()
		defer c.Unlock()

		src = strings.Trim(src, "\r\n")

		if !c.caching {
			return src
		}

		u, err := url.Parse(src)

		if err != nil || len(filepath.Ext(u.Path)) == 0 {
			return src
		}

		src_filtered := strings.Split(src, "?")
		var filename = fmt.Sprintf("%s%s", strToMD5(src_filtered[0]), filepath.Ext(u.Path))

		if cacheURL, ok := c.images[filename]; ok {
			if c.caching && force_https {
				u, err := url.Parse(cacheURL)
				if err == nil {
					cacheURL = fmt.Sprintf("https://%s:%d%s", https_domain, https_port, u.Path)
				}
			} else if c.caching && http_domain != "" {
				u, err := url.Parse(cacheURL)
				if err == nil {
					var baseUrl string
					if strings.Contains(http_domain, ":") {
						baseUrl = http_domain
					} else {
						baseUrl = fmt.Sprintf("%s:%s", http_domain, http_port)
					}
					cacheURL = fmt.Sprintf("http://%s%s", baseUrl, u.Path)
				}
			}
			return cacheURL
		}

		if slices.Index(c.Cache, filename) == -1 {
			if slices.Index(c.Queue, src) == -1 {
				c.Queue = append(c.Queue, src)
			}

		} else {
			c.images[filename] = c.cacheURL + filename
			src = c.cacheURL + filename
		}

		return src
	}

	c.Image.Caching = func() {
		// Serialize download passes, without holding the map/queue lock over I/O.
		c.cacheMu.Lock()
		defer c.cacheMu.Unlock()
		c.Lock()
		queue := slices.Clone(c.Queue)
		c.Unlock()

		for _, src := range queue {
			if _, cached := c.cacheImage(src); cached {
				c.Lock()
				c.Queue = removeStringFromSlice(src, c.Queue)
				c.Unlock()
			}
		}
	}

	c.Image.Remove = func() {

		c.Lock()
		defer c.Unlock()

		files, err := os.ReadDir(c.path)
		if err != nil {
			return
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), imageTempPrefix) {
				continue
			}
			if c.caching {
				if _, ok := c.images[file.Name()]; ok {
					continue
				}
			}
			if err := os.RemoveAll(filepath.Join(c.path, file.Name())); err == nil {
				c.Cache = removeStringFromSlice(file.Name(), c.Cache)
			}
		}

	}

	files, err := os.ReadDir(c.path)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), imageTempPrefix) {
			c.Cache = append(c.Cache, file.Name())
		}
	}

	return
}

// Pending returns the queue length without racing with lookups or caching.
func (c *Cache) Pending() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.Queue)
}
