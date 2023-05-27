package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	router.GET("/oauth/callback", func(context *gin.Context) {
		values := context.Request.URL.Query()

		callbackRaw := values.Get("state")
		if callbackRaw == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("state from github is empty"))
		}

		code := values.Get("code")
		if code == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("code from github is empty"))
			return
		}

		callbackParsed, err := url.ParseQuery(callbackRaw)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		callback := callbackParsed.Get("calback")
		if callback == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("callback in state is empty"))
			return
		}

		state := callbackParsed.Get("state")
		if state == "" {
			context.AbortWithError(http.StatusBadRequest, fmt.Errorf("state in state is empty"))
			return
		}

		u, err := url.Parse(callback)
		if err != nil {
			context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		u.Scheme = "https"
		q := u.Query()
		q.Set("code", code)
		q.Set("state", state)
		u.RawQuery = q.Encode()

		context.Redirect(http.StatusFound, u.String())
	})

	router.Run(":8080")
}
