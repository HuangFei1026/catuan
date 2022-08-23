package web

import "catuan/comm"

type ContextInf interface {
	Result(errCode int, errMsg string, data ...interface{})
	IsNext() bool
	AbortHandler()
	ActionLabel() string
	RoleLabel() string
	GroupLabel() string

	RespChannel() <-chan *comm.RespResult
	JsonResponse(resp *comm.RespResult)
}
