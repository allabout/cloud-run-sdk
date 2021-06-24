package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	m "cloud.google.com/go/compute/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

	idToken, err := getIDToken(addr)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		addr,
		grpc.WithAuthority(addr),
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(InjectClientAuthInterceptor(idToken)),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getIDToken(addr string) (string, error) {
	serviceURL := fmt.Sprintf("https://%s", strings.Split(addr, ":")[0])
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)

	idToken, err := m.Get(tokenURL)
	if err != nil {
		return "", err
	}

	return idToken, nil
}
