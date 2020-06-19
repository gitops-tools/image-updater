package cmd

import (
	"crypto/tls"
	"net/http"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func createClientFromViper() (*scm.Client, error) {
	authToken := viper.GetString(authTokenFlag)
	if viper.GetBool(insecureFlag) {
		return factory.NewClient(
			viper.GetString(driverFlag),
			viper.GetString(apiEndpointFlag),
			"",
			factory.Client(makeInsecureClient(authToken)))

	}
	return factory.NewClient(
		viper.GetString(driverFlag),
		viper.GetString(apiEndpointFlag),
		authToken)
}

func makeInsecureClient(token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &http.Client{
		Transport: &oauth2.Transport{
			Source: ts,
			Base: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}
