package main

type Protocol string

const (
	ProtocolNitroRPCv02 Protocol = "NitroRPC/0.2"
)

func (p Protocol) String() string {
	return string(p)
}
func IsSupportedProtocol(protocol Protocol) bool {
	return protocol == ProtocolNitroRPCv02
}
