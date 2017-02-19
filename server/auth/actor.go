// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package auth

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/gcloud/googleoauth"
	"github.com/luci/luci-go/common/gcloud/iam"
	"github.com/luci/luci-go/common/logging"
)

// MintAccessTokenParams is passed to MintAccessTokenForServiceAccount.
type MintAccessTokenParams struct {
	// ServiceAccount is an email of a service account to mint a token for.
	ServiceAccount string

	// Scopes is a list of OAuth scopes the token should have.
	Scopes []string

	// MinTTL defines an acceptable token lifetime.
	//
	// The returned token will be valid for at least MinTTL, but no longer than
	// one hour.
	//
	// Default is 2 min.
	MinTTL time.Duration
}

// actorTokenCache is used to store access tokens of a service accounts the
// current service has "iam.serviceAccountActor" role in.
//
// The underlying token type is cachedOAuth2Token.
var actorTokenCache = tokenCache{
	Kind:                "as_actor_tokens",
	Version:             1,
	ExpRandPercent:      10,
	MinAcceptedLifetime: 5 * time.Minute,
}

// cachedOAuth2Token is gob-serializable representation of the oauth2.Token.
//
// It explicitly contains only stuff we want to be in the cache. Storing
// oauth2.Token directly is dangerous because we don't control what oauth2 lib
// has in the Token struct (it may be non-serializable).
type cachedOAuth2Token struct {
	AccessToken string
	TokenType   string
	Expiry      time.Time
}

func makeCachedOAuth2Token(tok *oauth2.Token) cachedOAuth2Token {
	return cachedOAuth2Token{
		AccessToken: tok.AccessToken,
		TokenType:   tok.TokenType,
		Expiry:      tok.Expiry,
	}
}

func (c *cachedOAuth2Token) toToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: c.AccessToken,
		TokenType:   c.TokenType,
		Expiry:      c.Expiry,
	}
}

func init() {
	gob.Register(cachedOAuth2Token{})
}

// MintAccessTokenForServiceAccount produces an access token for some service
// account that the current service has "iam.serviceAccountActor" role in.
//
// Used to implement AsActor authorization kind, but can also be used directly,
// if needed. The token is cached internally. Same token may be returned by
// multiple calls, if its lifetime allows.
//
// Recognizes transient errors and marks them, but does not automatically
// retry. Has internal timeout of 10 sec.
func MintAccessTokenForServiceAccount(ctx context.Context, params MintAccessTokenParams) (*oauth2.Token, error) {
	report := durationReporter(ctx, mintAccessTokenDuration)

	cfg := GetConfig(ctx)
	if cfg == nil || cfg.AccessTokenProvider == nil {
		report(ErrNotConfigured, "ERROR_NOT_CONFIGURED")
		return nil, ErrNotConfigured
	}

	if params.ServiceAccount == "" || len(params.Scopes) == 0 {
		err := fmt.Errorf("invalid parameters")
		report(err, "ERROR_BAD_ARGUMENTS")
		return nil, err
	}

	if params.MinTTL == 0 {
		params.MinTTL = 2 * time.Minute
	}

	sortedScopes := append([]string(nil), params.Scopes...)
	sort.Strings(sortedScopes)

	// Construct the cache key. Note that it is hashed by 'actorTokenCache' and
	// thus can be as long as necessary. Double check there's no malicious input.
	parts := append([]string{params.ServiceAccount}, sortedScopes...)
	for _, p := range parts {
		if strings.ContainsRune(p, '\n') {
			err := fmt.Errorf("forbidding character in a service account or scope name: %q", p)
			report(err, "ERROR_BAD_ARGUMENTS")
			return nil, err
		}
	}
	cacheKey := strings.Join(parts, "\n")

	// Try to find an existing cached token and check that it lives long enough.
	now := clock.Now(ctx).UTC()
	switch cached, err := actorTokenCache.Fetch(ctx, cacheKey); {
	case err != nil:
		report(err, "ERROR_CACHE")
		return nil, err
	case cached != nil && cached.Expiry.After(now.Add(params.MinTTL)):
		t := cached.Token.(cachedOAuth2Token) // let it panic on type mismatch
		report(nil, "SUCCESS_CACHE_HIT")
		return t.toToken(), nil
	}

	// Both IAM API call and token endpoint should be fast. If it gets stuck
	// longer than 10 sec, it is probably busted already.
	ctx, cancel := clock.WithTimeout(ctx, cfg.adjustedTimeout(10*time.Second))
	defer cancel()
	ctx = logging.SetFields(ctx, logging.Fields{
		"method":  "AsActor",
		"account": params.ServiceAccount,
		"scopes":  strings.Join(sortedScopes, " "),
	})
	logging.Debugf(ctx, "Minting access token")

	// Need an authenticating transport to talk to IAM.
	asSelf, err := GetRPCTransport(ctx, AsSelf, WithScopes(iam.OAuthScope))
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Failed to grab a transport for IAM call")
		report(err, "ERROR_NO_TRANSPORT")
		return nil, err
	}

	// This will do two HTTP calls: one to 'signBytes' IAM API, another to the
	// token exchange endpoint.
	tok, err := googleoauth.GetAccessToken(ctx, googleoauth.JwtFlowParams{
		ServiceAccount: params.ServiceAccount,
		Signer: &iam.Signer{
			ServiceAccount: params.ServiceAccount,
			Client: &iam.Client{
				Client: &http.Client{Transport: asSelf},
			},
		},
		Scopes: sortedScopes,
		Client: &http.Client{Transport: cfg.AnonymousTransport(ctx)},
	})

	// Both iam.Signer and googleoauth.GetAccessToken return googleapi.Error on
	// HTTP-level responses. Recognize fatal HTTP errors. Everything else (stuff
	// like connection timeouts, deadlines, etc) are transient errors.
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Failed to mint an access token")
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code < 500 {
			report(err, fmt.Sprintf("ERROR_HTTP_%d", apiErr.Code))
			return nil, err
		}
		report(err, "ERROR_TRANSIENT")
		return nil, errors.WrapTransient(err)
	}

	// Cache the token. Ignore errors here, it's not big deal, we have the token.
	err = actorTokenCache.Store(ctx, cachedToken{
		Key:     cacheKey,
		Token:   makeCachedOAuth2Token(tok),
		Created: now,
		Expiry:  tok.Expiry,
	})
	if err != nil {
		logging.WithError(err).Warningf(ctx, "Failed to store the access token in the cache")
	}

	report(nil, "SUCCESS_CACHE_MISS")
	return tok, nil
}