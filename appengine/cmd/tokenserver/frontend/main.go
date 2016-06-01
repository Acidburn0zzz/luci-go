// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package frontend implements HTTP server that handles requests to default
// module.
//
// It stitches together all the code.
package frontend

import (
	"net/http"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/luci/gae/service/info"
	"github.com/luci/luci-go/appengine/gaeauth/server"
	"github.com/luci/luci-go/appengine/gaemiddleware"
	"github.com/luci/luci-go/appengine/tsmon"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/server/auth"
	"github.com/luci/luci-go/server/auth/machine"
	"github.com/luci/luci-go/server/discovery"
	"github.com/luci/luci-go/server/prpc"

	"github.com/luci/luci-go/common/api/tokenserver/admin/v1"
	"github.com/luci/luci-go/common/api/tokenserver/identity/v1"
	"github.com/luci/luci-go/common/api/tokenserver/minter/v1"

	"github.com/luci/luci-go/appengine/cmd/tokenserver/services/admin/certauthorities"
	"github.com/luci/luci-go/appengine/cmd/tokenserver/services/admin/serviceaccounts"
	"github.com/luci/luci-go/appengine/cmd/tokenserver/services/identity/identityfetcher"
	"github.com/luci/luci-go/appengine/cmd/tokenserver/services/minter/tokenminter"
)

var (
	// caServer implements admin.CertificateAuthorities RPC interface.
	caServerWithoutAuth = &certauthorities.Server{}

	// caServerWithAuth adds admin check to caServer.
	caServerWithAuth = &admin.DecoratedCertificateAuthorities{
		Service: caServerWithoutAuth,
		Prelude: adminPrelude("admin.CertificateAuthorities"),
	}

	// serviceAccountsServer implements admin.ServiceAccounts RPC interface.
	serviceAccountsServerWithoutAuth = &serviceaccounts.Server{}

	// serviceAccountsServerWithAuth adds admin check to serviceAccountsServer.
	serviceAccountsServerWithAuth = &admin.DecoratedServiceAccounts{
		Service: serviceAccountsServerWithoutAuth,
		Prelude: adminPrelude("admin.ServiceAccounts"),
	}

	// identityFetcher implements identity.IdentityFetcher RPC interface.
	identityFetcher = &identityfetcher.Server{}

	// tokenMinterServer implements minter.TokenMinter RPC interface.
	//
	// It is main public API of the token server. It doesn't require any external
	// authentication (it happens inside), and so it's installed as is.
	tokenMinterServerWithoutAuth = tokenminter.NewServer(serviceAccountsServerWithoutAuth)
)

// adminPrelude returns a prelude that authorizes only administrators.
func adminPrelude(serviceName string) func(context.Context, string, proto.Message) (context.Context, error) {
	return func(c context.Context, method string, _ proto.Message) (context.Context, error) {
		logging.Infof(c, "%s: %q is calling %q", serviceName, auth.CurrentIdentity(c), method)
		switch admin, err := auth.IsMember(c, "administrators"); {
		case err != nil:
			return nil, grpc.Errorf(codes.Internal, "can't check ACL - %s", err)
		case !admin:
			return nil, grpc.Errorf(codes.PermissionDenied, "not an admin")
		}
		return c, nil
	}
}

func init() {
	router := httprouter.New()
	base := gaemiddleware.BaseProd

	// Install auth, config and tsmon handlers.
	gaemiddleware.InstallHandlers(router, base)

	// The service has no UI, so just redirect to stock RPC explorer.
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/rpcexplorer/", http.StatusFound)
	})

	// Optional warmup routes.
	router.GET("/_ah/warmup", base(warmupHandler))
	router.GET("/_ah/start", base(warmupHandler))

	// Backend routes used for cron and task queues.
	router.GET("/internal/cron/read-config", base(gaemiddleware.RequireCron(readConfigCron)))
	router.GET("/internal/cron/fetch-crl", base(gaemiddleware.RequireCron(fetchCRLCron)))

	// Install all RPC servers.
	api := prpc.Server{
		Authenticator: auth.Authenticator{
			&server.OAuth2Method{Scopes: []string{server.EmailScope}},
			&machine.MachineTokenAuthMethod{},
		},
		UnaryServerInterceptor: tsmon.NewGrpcUnaryInterceptor(nil),
	}
	admin.RegisterCertificateAuthoritiesServer(&api, caServerWithAuth)
	admin.RegisterServiceAccountsServer(&api, serviceAccountsServerWithAuth)
	identity.RegisterIdentityFetcherServer(&api, identityFetcher)
	minter.RegisterTokenMinterServer(&api, tokenMinterServerWithoutAuth) // auth inside
	discovery.Enable(&api)
	api.InstallHandlers(router, base)

	// Expose all this stuff.
	http.DefaultServeMux.Handle("/", router)
}

/// Routes.

// warmupHandler warms in-memory caches.
func warmupHandler(c context.Context, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := server.Warmup(c); err != nil {
		panic(err) // let panic catcher deal with it
	}
	w.WriteHeader(http.StatusOK)
}

// readConfigCron is handler for /internal/cron/read-config GAE cron task.
func readConfigCron(c context.Context, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Don't override manually imported configs with 'nil' on devserver.
	if info.Get(c).IsDevAppServer() {
		w.WriteHeader(http.StatusOK)
		return
	}
	if _, err := caServerWithoutAuth.ImportConfig(c, nil); err != nil {
		panic(err) // let panic catcher deal with it
	}
	w.WriteHeader(http.StatusOK)
}

// fetchCRLCron is handler for /internal/cron/fetch-crl GAE cron task.
func fetchCRLCron(c context.Context, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	list, err := caServerWithoutAuth.ListCAs(c, nil)
	if err != nil {
		panic(err) // let panic catcher deal with it
	}

	// Fetch CRL of each active CA in parallel. In practice there are very few
	// CAs there (~= 1), so the risk of OOM is small.
	wg := sync.WaitGroup{}
	errs := make([]error, len(list.Cn))
	for i, cn := range list.Cn {
		wg.Add(1)
		go func(i int, cn string) {
			defer wg.Done()
			_, err := caServerWithoutAuth.FetchCRL(c, &admin.FetchCRLRequest{Cn: cn})
			if err != nil {
				logging.Errorf(c, "FetchCRL(%q) failed - %s", cn, err)
				errs[i] = err
			}
		}(i, cn)
	}
	wg.Wait()

	// Retry cron job only on transient errors. On fatal errors let it rerun one
	// minute later, as usual, to avoid spamming logs with errors.
	status := http.StatusOK
	for _, err = range errs {
		if grpc.Code(err) == codes.Internal {
			status = http.StatusInternalServerError
			break
		}
	}
	w.WriteHeader(status)
}
