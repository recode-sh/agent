package devenv

import "github.com/recode-sh/agent/proto"

type GRPCBuildAndStartDevEnvStreamWriter struct {
	Stream proto.Agent_BuildAndStartDevEnvServer
}

func NewGRPCBuildAndStartDevEnvStreamWriter(
	stream proto.Agent_BuildAndStartDevEnvServer,
) GRPCBuildAndStartDevEnvStreamWriter {

	return GRPCBuildAndStartDevEnvStreamWriter{
		Stream: stream,
	}
}

func (g GRPCBuildAndStartDevEnvStreamWriter) Write(
	p []byte,
) (int, error) {

	streamSendErr := g.Stream.Send(&proto.BuildAndStartDevEnvReply{
		LogLine: string(p),
	})

	if streamSendErr != nil {
		return 0, streamSendErr
	}

	return len(p), nil
}
