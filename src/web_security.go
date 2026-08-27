package src

import "net/http"

const browserContentSecurityPolicy = "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' data: https://cdnjs.cloudflare.com; img-src 'self' data: http: https:; connect-src 'self' ws: wss:; form-action 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'"

func setBrowserSecurityHeaders(header http.Header) {
	header.Set("Content-Security-Policy", browserContentSecurityPolicy)
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Frame-Options", "DENY")
}
