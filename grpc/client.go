package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"github.com/allabout/cloud-run-sdk/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// addr 127.0.0.1:443
func NewTLSConn(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	idToken, err := util.GetIDToken(addr)
	if err != nil {
		return nil, err
	}

	opts = append([]grpc.DialOption{
		grpc.WithAuthority(addr),
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(AuthInterceptor(idToken))},
		opts...,
	)

	return grpc.DialContext(ctx, addr, opts...)
}
