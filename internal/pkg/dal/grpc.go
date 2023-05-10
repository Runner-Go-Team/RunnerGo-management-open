package dal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/go-omnibus/proof"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	services "github.com/Runner-Go-Team/RunnerGo-management-open/api"
)

var (
	//grpcClient services.KpControllerClient
	conn *grpc.ClientConn
)

func MustInitGRPC() {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		panic(fmt.Errorf("cannot load root CA certs, %w", err))
	}
	creds := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	conn, err = grpc.Dial("kpcontroller.apipost.cn:443", grpc.WithTransportCredentials(creds))

	if err != nil {
		proof.Errorf("grpc dial err", err)
	}
}

func ClientGRPC() services.KpControllerClient {
	return services.NewKpControllerClient(conn)
}
