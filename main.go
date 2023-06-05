package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	// ALLOWED_REDIRECT_URLS will list the urls that the proxy is allowed to redirect to
	// one of the domains listed here should match the one provided by the callback query parameter encoded in state
	allowedRedirects := []string{}

	allowedRedirectsEnv, allowedRedirectsEnvIsSet := os.LookupEnv("ALLOWED_REDIRECT_URLS")
	if allowedRedirectsEnvIsSet {
		// if the env is set, override the defualt values
		allowedRedirects = strings.Split(allowedRedirectsEnv, ",")
	}

	log.Println("configured ALLOWED_REDIRECT_URLS: ", allowedRedirects)

	router.GET("/oauth/callback", func(context *gin.Context) {

		values := context.Request.URL.Query()

		callbackRaw := values.Get("state")
		if callbackRaw == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("state from github is empty"))
			log.Println("state from github is empty")
			return
		}

		code := values.Get("code")
		if code == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("code from github is empty"))
			log.Println("code from github is empty")
			return
		}

		callbackParsed, err := url.ParseQuery(callbackRaw)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		callback := callbackParsed.Get("callback")
		if callback == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("callback in state is empty"))
			log.Println("callback in state is empty")
			return
		}

		state := callbackParsed.Get("state")
		if state == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("state in state is empty"))
			log.Println("state in state is empty")
			return
		}

		u, err := url.Parse(callback)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		u.Scheme = "https"
		q := u.Query()
		q.Set("code", code)
		q.Set("state", state)
		u.RawQuery = q.Encode()

		redirectAllowed := isDomainAllowed(u.Host, allowedRedirects)
		if !redirectAllowed {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("redirect url is not allowed"))
			return
		}
		context.Redirect(http.StatusMovedPermanently, u.String())
	})

	router.Run(":8080")
}

func isDomainAllowed(domain string, allowedDomains []string) bool {
	for _, d := range allowedDomains {
		match := strings.HasSuffix(domain, d)
		if match {
			log.Println(domain, " is allowed")
			return true
		}
	}
	log.Println(domain, " is not allowed")
	return false
}
