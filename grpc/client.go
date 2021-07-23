package grpc

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/ishii1648/cloud-run-sdk/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// addr 127.0.0.1:443
func NewConn(addr string, localhost bool) (*grpc.ClientConn, error) {
	if localhost {
		conn, err := grpc.Dial(
			addr,
			grpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	return newTLSConn(addr)
}

func newTLSConn(addr string) (*grpc.ClientConn, error) {
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

	conn, err := grpc.Dial(
		addr,
		grpc.WithAuthority(addr),
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(AuthInterceptor(idToken)),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
