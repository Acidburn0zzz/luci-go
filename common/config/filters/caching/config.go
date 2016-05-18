// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package caching

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/luci/luci-go/common/config"
	"golang.org/x/net/context"
)

const (
	version = "v2"
)

var errCacheMiss = errors.New("cache miss")

type contentKey byte

const (
	contentHit = contentKey(iota)
	contentErrNoConfig
)

// Cache implements a generic caching layer.
//
// The layer has no consistency expectations.
type Cache interface {
	// Store adds data to the cache. If value is nil, the cache should invalidate
	// the supplied key.
	Store(c context.Context, key string, expire time.Duration, value []byte)

	// Retrieve pulls data from the cache. If no data is available, nil will be
	// returned with no error.
	Retrieve(c context.Context, key string) []byte
}

// Options is the set of configuration options for the caching layer.
type Options struct {
	// Cache is the caching layer to use.
	Cache Cache

	// Expiration is the maximum amount of time that the configuration should be
	// retained before it must be refreshed.
	//
	// Due to implementation details of the cache layer, the configuration may be
	// retained for less time if necessary.
	Expiration time.Duration
}

// NewFilter constructs a new caching config.Filter function.
func NewFilter(o Options) config.Filter {
	return func(c context.Context, cc config.Interface) config.Interface {
		return &cacheConfig{
			Context: c,
			Options: &o,
			inner:   cc,
		}
	}
}

// cacheConfig implements a config.Interface that caches results in MemCache.
type cacheConfig struct {
	context.Context
	*Options

	inner config.Interface
}

func (cc *cacheConfig) ServiceURL() url.URL {
	return cc.inner.ServiceURL()
}

func (cc *cacheConfig) GetConfig(configSet, path string, hashOnly bool) (*config.Config, error) {
	// If we're doing hash-only lookup, we're okay with either full or hash-only
	// result. However, if we have to do the lookup, we will store the result in
	// a hash-only cache bucket.
	c := config.Config{}
	key := cc.cacheKey("configs", "full", configSet, path)
	switch err := cc.retrieve(key, &c); err {
	case nil:
		// Cache hit.
		return &c, nil
	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return nil, err
	}

	if hashOnly {
		key = cc.cacheKey("configs", "hashOnly", configSet, path)
		switch err := cc.retrieve(key, &c); err {
		case nil:
			// Cache hit.
			return &c, nil
		case errCacheMiss:
			// Cache miss, load from service.
			break
		default:
			// Return cached error.
			return nil, err
		}
	}

	ic, err := cc.inner.GetConfig(configSet, path, hashOnly)
	if err != nil {
		cc.storeErr(key, err)
		return nil, err
	}

	cc.store(key, ic)
	if !hashOnly {
		cc.store(cc.configByHashCacheKey(ic.ContentHash), ic.Content)
	}
	return ic, nil
}

func (cc *cacheConfig) GetConfigByHash(contentHash string) (string, error) {
	c := ""
	key := cc.configByHashCacheKey(contentHash)
	switch err := cc.retrieve(key, &c); err {
	case nil:
		// Cache hit.
		return c, nil
	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return "", err
	}

	c, err := cc.inner.GetConfigByHash(contentHash)
	if err != nil {
		cc.storeErr(key, err)
		return "", err
	}

	cc.store(key, c)
	return c, nil
}

func (cc *cacheConfig) configByHashCacheKey(contentHash string) string {
	return cc.cacheKey("configsByHash", contentHash)
}

func (cc *cacheConfig) GetConfigSetLocation(configSet string) (*url.URL, error) {
	v := ""
	key := cc.cacheKey("configSet", "location", configSet)
	switch err := cc.retrieve(key, &v); err {
	case nil:
		// Cache hit.
		u, err := url.Parse(v)
		if err != nil {
			return u, nil
		}

	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return nil, err
	}

	u, err := cc.inner.GetConfigSetLocation(configSet)
	if err != nil {
		cc.storeErr(key, err)
		return nil, err
	}

	cc.store(key, u.String())
	return u, nil
}

func (cc *cacheConfig) GetProjectConfigs(path string, hashesOnly bool) ([]config.Config, error) {
	var c []config.Config
	key := cc.cacheKey("projectConfigs", "full", path)
	switch err := cc.retrieve(key, &c); err {
	case nil:
		// Cache hit.
		return c, nil
	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return nil, err
	}

	if hashesOnly {
		key = cc.cacheKey("projectConfigs", "hashesOnly", path)
	}

	c, err := cc.inner.GetProjectConfigs(path, hashesOnly)
	if err != nil {
		cc.storeErr(key, err)
		return nil, err
	}

	cc.store(key, c)
	return c, nil
}

func (cc *cacheConfig) GetProjects() ([]config.Project, error) {
	p := []config.Project(nil)
	key := cc.cacheKey("projects")
	switch err := cc.retrieve(key, &p); err {
	case nil:
		// Cache hit.
		return p, nil
	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return nil, err
	}

	p, err := cc.inner.GetProjects()
	if err != nil {
		cc.storeErr(key, err)
		return nil, err
	}

	cc.store(key, p)
	return p, nil
}

func (cc *cacheConfig) GetRefConfigs(path string, hashesOnly bool) ([]config.Config, error) {
	c := []config.Config(nil)
	key := cc.cacheKey("refConfigs", "full", path)
	switch err := cc.retrieve(key, &c); err {
	case nil:
		// Cache hit.
		return c, nil
	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return nil, err
	}

	if hashesOnly {
		key = cc.cacheKey("refConfigs", "hashesOnly", path)
	}

	c, err := cc.inner.GetRefConfigs(path, hashesOnly)
	if err != nil {
		cc.storeErr(key, err)
		return nil, err
	}

	cc.store(key, c)
	return c, nil
}

func (cc *cacheConfig) GetRefs(projectID string) ([]string, error) {
	var refs []string
	key := cc.cacheKey("refs", projectID)
	switch err := cc.retrieve(key, &refs); err {
	case nil:
		// Cache hit.
		return refs, nil
	case errCacheMiss:
		// Cache miss, load from service.
		break
	default:
		// Return cached error.
		return nil, err
	}

	refs, err := cc.inner.GetRefs(projectID)
	if err != nil {
		cc.storeErr(key, err)
		return nil, err
	}

	cc.store(key, refs)
	return refs, nil
}

func (cc *cacheConfig) retrieve(key string, v interface{}) error {
	if cc.Cache == nil {
		return errCacheMiss
	}

	// Load the cache value.
	d := cc.Cache.Retrieve(cc, key)
	if len(d) == 0 {
		return errCacheMiss
	}

	// Handle the content key.
	switch contentKey(d[0]) {
	case contentHit:
		d = d[1:]
		break

	case contentErrNoConfig:
		return config.ErrNoConfig

	default:
		// Unknown content key, treat as cache miss.
		return errCacheMiss
	}

	// Unzip.
	zr, err := zlib.NewReader(bytes.NewBuffer(d))
	if err != nil {
		return errCacheMiss
	}
	defer zr.Close()

	rd, err := ioutil.ReadAll(zr)
	if err != nil {
		return errCacheMiss
	}

	// Unpack.
	if err := json.Unmarshal(rd, v); err != nil {
		return errCacheMiss
	}

	return nil
}

func (cc *cacheConfig) store(key string, v interface{}) {
	if cc.Cache == nil {
		return
	}

	// Convert "v" to zlib-compressed JSON.
	d, err := json.Marshal(v)
	if err != nil {
		return
	}

	// Write a "content hit" record.
	buf := bytes.Buffer{}
	buf.WriteByte(byte(contentHit))
	w := zlib.NewWriter(&buf)
	_, err = w.Write(d)
	if err != nil {
		w.Close()
		return
	}
	if err := w.Close(); err != nil {
		return
	}

	cc.Cache.Store(cc, key, cc.Expiration, buf.Bytes())
}

func (cc *cacheConfig) storeErr(key string, err error) {
	if cc.Cache == nil {
		return
	}

	switch err {
	case config.ErrNoConfig:
		cc.Cache.Store(cc, key, cc.Expiration, []byte{byte(contentErrNoConfig)})

	default:
		// Don't know how to store this error type.
		return
	}
}

// cacheKey constructs a cache key from a set of value segments.
//
// In order to ensure that segments remain distinct in the resulting key, each
// segment is URL query escaped, and segments are joined by the non-escaped
// character, "|".
//
//   For example, ["a|b", "c"] => "a%7Cb|c"
func (cc *cacheConfig) cacheKey(values ...string) string {
	enc := url.QueryEscape
	parts := make([]string, 0, len(values)+1)
	parts = append(parts, enc(version))
	for _, v := range values {
		parts = append(parts, enc(v))
	}

	return strings.Join(parts, "|")
}
