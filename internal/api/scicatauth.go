package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID           string `json:"id"`
	AuthStrategy string `json:"authStrategy"`
	ExternalID   string `json:"externalId"`
	Profile      struct {
		DisplayName    string `json:"displayName"`
		Email          string `json:"email"`
		Username       string `json:"username"`
		ThumbnailPhoto string `json:"thumbnailPhoto"`
		ID             string `json:"id"`
		Emails         []struct {
			Value string `json:"value"`
		} `json:"emails"`
		AccessGroups []string `json:"accessGroups"`
		OidcClaims   struct {
			Exp               int      `json:"exp"`
			Iat               int      `json:"iat"`
			AuthTime          int      `json:"auth_time"`
			Jti               string   `json:"jti"`
			Iss               string   `json:"iss"`
			Aud               string   `json:"aud"`
			Sub               string   `json:"sub"`
			Typ               string   `json:"typ"`
			Azp               string   `json:"azp"`
			Sid               string   `json:"sid"`
			AtHash            string   `json:"at_hash"`
			Acr               string   `json:"acr"`
			EmailVerified     bool     `json:"email_verified"`
			AccessGroups      []string `json:"accessGroups"`
			Name              string   `json:"name"`
			PreferredUsername string   `json:"preferred_username"`
			GivenName         string   `json:"given_name"`
			FamilyName        string   `json:"family_name"`
			Email             string   `json:"email"`
		} `json:"oidcClaims"`
		ID_ string `json:"_id"`
	} `json:"profile"`
	Provider    string    `json:"provider"`
	UserID      string    `json:"userId"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	V           int       `json:"__v"`
	ScicatToken string
}

type GeneralError struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

type GroupTemplateData struct {
	FacilityName string
}

func ScicatTokenAuthMiddleware(scicatUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scicatApiKey := c.Request.Header.Get("SciCat-API-Key")

		userIdentityUrl, err := url.JoinPath(scicatUrl, "users", "my", "identity")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, GeneralError{
				Message: "couldn't create request url for scicat token verification request",
				Details: err.Error(),
			})
			return
		}

		req, err := http.NewRequest("GET", userIdentityUrl, nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, GeneralError{
				Message: fmt.Sprintf("couldn't create GET request using the path '%s'", userIdentityUrl),
				Details: err.Error(),
			})
			return
		}

		req.Header.Set("Authorization", "Bearer "+scicatApiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, GeneralError{
				Message: "couldn't make the http request to verify token validity",
				Details: err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			c.AbortWithStatusJSON(http.StatusUnauthorized, GeneralError{
				Message: "the access token provided with the request is invalid",
				Details: fmt.Sprintf("status: '%d', body: '%s'", resp.StatusCode, string(body)),
			})
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, GeneralError{})
			return
		}

		var user User
		json.Unmarshal(body, &user)
		user.ScicatToken = scicatApiKey
		c.Set("scicatUser", user)
		c.Next()
	}
}
