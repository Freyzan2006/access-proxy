package server

import (
	"net/http"
	"strings"
)

type domainUtils struct {
	allowedDomains []string
}

func newDomainUtils(allowedDomains []string) *domainUtils {
	return &domainUtils{allowedDomains: allowedDomains}
}

func (u *domainUtils) extractClientDomain(r *http.Request) string {
	if origin := r.Header.Get("Origin"); origin != "" {
		return u.extractDomainFromURL(origin)
	}
	if referer := r.Header.Get("Referer"); referer != "" {
		return u.extractDomainFromURL(referer)
	}
	if host := r.Header.Get("Host"); host != "" {
		return strings.Split(host, ":")[0]
	}
	return ""
}

func (u *domainUtils) extractDomainFromURL(urlStr string) string {
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
	}

	parts := strings.Split(urlStr, "://")
	if len(parts) < 2 {
		return ""
	}

	hostParts := strings.Split(parts[1], "/")
	host := hostParts[0]
	return strings.Split(host, ":")[0]
}

func (u *domainUtils) isClientAllowed(r *http.Request) bool {
	if len(u.allowedDomains) == 0 {
		return true
	}

	clientDomain := u.extractClientDomain(r)
	for _, allowed := range u.allowedDomains {
		if clientDomain == allowed {
			return true
		}
	}
	return false
}

func (u *domainUtils) extractDomain(urlStr string) string {
	parts := splitTwo(urlStr, "://")
	if len(parts) < 2 {
		return ""
	}

	hostParts := splitTwo(parts[1], "/")
	return splitTwo(hostParts[0], ":")[0]
}

func (u *domainUtils) isTargetAllowed(target string) bool {
	if len(u.allowedDomains) == 0 {
		return true
	}

	targetDomain := u.extractDomain(target)
	for _, allowed := range u.allowedDomains {
		if targetDomain == allowed {
			return true
		}
	}
	return false
}