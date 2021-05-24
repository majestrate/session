package rpc

type RPC interface {
	Call(method string, params map[string]interface{})
}
