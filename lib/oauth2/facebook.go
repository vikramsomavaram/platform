/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package oauth2

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	facebookOauthConfig = &oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		RedirectURL:  "YOUR_REDIRECT_URL_CALLBACK",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	oauthStateString = "thisshouldberandom"
)

//const htmlIndex = `<html><body>
//Logged in with <a href="/login">facebook</a>
//</body></html>
//`
//
//func handleMain(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "text/html; charset=utf-8")
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte(htmlIndex))
//}

//func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
//	Url, err := url.Parse(oauthConf.Endpoint.AuthURL)
//	if err != nil {
//		log.Fatal("Parse: ", err)
//	}
//	parameters := url.Values{}
//	parameters.Add("client_id", oauthConf.ClientID)
//	parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
//	parameters.Add("redirect_uri", oauthConf.RedirectURL)
//	parameters.Add("response_type", "code")
//	parameters.Add("state", oauthStateString)
//	Url.RawQuery = parameters.Encode()
//	url := Url.String()
//	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
//}

func handleFacebookAuth(w http.ResponseWriter, r *http.Request) {
	//state := r.FormValue("state")
	//if state != oauthStateString {
	//	fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
	//	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	//	return
	//}

	code := r.FormValue("code")

	token, err := facebookOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
		url.QueryEscape(token.AccessToken))
	if err != nil {
		log.Printf("Get: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ReadAll: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Printf("parseResponseBody: %s\n", string(response))

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
