// Copyright 2016 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package svcmux contains utility functions used by code generated by
// svcmux tool.
// It is not designed to be used directly.
package svcmux

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// VersionMetadataKey is key in gRPC metadata that specifies requested
// service version.
const VersionMetadataKey = "X-Luci-Service-Version"

// GetServiceVersion extracts requested service version from metadata in c.
func GetServiceVersion(c context.Context, defaultVer string) string {
	if md, ok := metadata.FromIncomingContext(c); ok {
		values := md[VersionMetadataKey]
		if len(values) != 0 {
			return values[0]
		}
	}
	return defaultVer
}

// NoImplementation creates an error for a service version that does not have an
// implementation.
func NoImplementation(version string) error {
	return status.Errorf(codes.Unimplemented, "version %s not implemented", version)
}
