package web

import (
	"catuan/comm"
	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
	isNext bool

	respChan    chan *comm.RespResult
	actionLabel string
	roleLabel   string
	groupLabel  string
}

func (c *Context) InitRoleInfo(roleLabel, groupLabel, actionLabel string) {
	c.groupLabel = groupLabel
	c.roleLabel = roleLabel
	c.actionLabel = actionLabel
}

func (c *Context) Result(errCode int, errMsg string, data ...interface{}) {
	if len(data) == 0 {
		c.respChan <- &comm.RespResult{
			ErrCode: errCode,
			ErrMsg:  errMsg,
		}
		return
	}
	if len(data) == 1 {
		c.respChan <- &comm.RespResult{
			ErrCode: errCode,
			ErrMsg:  errMsg,
			Data:    data[0],
		}
		return
	}
	c.respChan <- &comm.RespResult{
		ErrCode: errCode,
		ErrMsg:  errMsg,
		Data:    data,
	}
}

func (c *Context) JsonResponse(resp *comm.RespResult) {
	c.JSON(200, resp)
}

func (c *Context) IsNext() bool {
	return c.isNext
}

func (c *Context) AbortHandler() {
	c.isNext = false
}

func (c *Context) ActionLabel() string {
	return c.actionLabel
}

func (c *Context) RoleLabel() string {
	return c.roleLabel
}

func (c *Context) GroupLabel() string {
	return c.groupLabel
}

func (c *Context) RespChannel() <-chan *comm.RespResult {
	return c.respChan
}
