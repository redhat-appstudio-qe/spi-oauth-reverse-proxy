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
	if allowedRedirectsEnvIsSet && allowedRedirectsEnv != "" {
		// if the env is set, override the defualt values
		allowedRedirects = strings.Split(allowedRedirectsEnv, ",")
	}

	log.Println("ALLOWED_REDIRECT_URLS ARE: ", allowedRedirects)

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

	var html = `
		<!DOCTYPE html>
  
		<html>
		<head>
				<title>Login Page</title>
				<style>
					input[type=text] {width: 500px; height: 40px; font-size: 15px;}
					input[type=submit] {width: 200px; height: 40px; font-size: 15px;}
				</style>
		</head>
		<body>
		
		<form id="loginForm" action="%%SPI_URL%%" method="POST">
				<input type="text" id="k8s_token" name="k8s_token" placeholder="your k8s_token goes here..." required>
				<br>
				<input type="submit" id="submit_token" name="submit_token" value="Submit" required>
		</form>
		
		<script type="text/javascript">
				window.onload = function() {
						//document.getElementById('loginForm').submit();
				};
		</script>
		
		</body>
		</html>
	`

	router.GET("/login", func(context *gin.Context) {
		values := context.Request.URL.Query()
		spi_url := values.Get("url")
		if spi_url == "" {
			context.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<p>required URL is empty</p>"))
			log.Println("url is empty")
			return
		}
		u, err := url.Parse(spi_url)
		if err != nil {
			context.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<p>error parsing url</p>"))
			log.Println(err)
			return
		}

		context.Data(http.StatusOK, "text/html; charset=utf-8", []byte(strings.Replace(html, "%%SPI_URL%%", u.String(), 1)))
	})
	router.Run(":8080")
}

func isDomainAllowed(domain string, allowedDomains []string) bool {
	if len(allowedDomains) == 0 {
		log.Println("no domains allowed")
		return false
	}
	for _, d := range allowedDomains {
		match := strings.HasSuffix(domain, d)
		if match {
			log.Println(domain, " is allowed")
			return true
		}
	}

	log.Println(domain, " IS NOT ALLOWED")

	return false
}
