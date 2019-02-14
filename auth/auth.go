// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package auth contains WJT auth methods.
// Based on https://cloud.yandex.ru/docs/iam/operations/iam-token/create-for-sa
package auth

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	authURL = "https://iam.api.cloud.yandex.net/iam/v1/tokens"
)

var ps256WithSaltLengthEqualsHash = &jwt.SigningMethodRSAPSS{
	SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
	Options: &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	},
}

func loadPrivateKey(keyFile string) *rsa.PrivateKey {
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		panic(err)
	}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		panic(err)
	}
	return rsaPrivateKey
}

func signedToken(keyID, serviceAccountID, keyFile string) (string, error) {
	issuedAt := time.Now()
	token := jwt.NewWithClaims(ps256WithSaltLengthEqualsHash, jwt.StandardClaims{
		Issuer:    serviceAccountID,
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: issuedAt.Add(time.Hour).Unix(),
		Audience:  authURL,
	})
	token.Header["kid"] = keyID

	privateKey := loadPrivateKey(keyFile)
	return token.SignedString(privateKey)
}

func getIAMToken(keyID, serviceAccountID, keyFile string) (string, error) {
	jot, err := signedToken(keyID, serviceAccountID, keyFile)
	if err != nil {
		return "", err
	}
	//fmt.Println(jot)
	resp, err := http.Post(
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, jot)),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(fmt.Sprintf("%s: %s", resp.Status, body))
	}
	var data struct {
		IAMToken string `json:"iamToken"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	return data.IAMToken, nil
}
