package constants

const (
	GRPCServerAddrProtocol = "unix"
	GRPCServerAddr         = "/tmp/recode_grpc.sock"

	SSHServerListenPort      = "2200"
	SSHServerListenAddr      = ":" + SSHServerListenPort
	SSHServerHostKeyFilePath = "/home/recode/.ssh/recode_ssh_server_host_key"

	InitInstanceScriptRepoPath = "recode-sh/agent/internal/grpcserver/init_instance.sh"
)
