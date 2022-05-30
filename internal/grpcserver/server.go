package grpcserver

import (
	_ "embed"
	"fmt"
	"net"

	"github.com/recode-sh/agent/proto"
	"google.golang.org/grpc"
)

type agentServer struct {
	proto.UnimplementedAgentServer
}

func ListenAndServe(serverAddrProtocol, serverAddr string) error {
	tcpServer, err := net.Listen(serverAddrProtocol, serverAddr)

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterAgentServer(grpcServer, &agentServer{})

	return grpcServer.Serve(tcpServer)
}
