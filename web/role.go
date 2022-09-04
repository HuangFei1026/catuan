package web

type RoleInf interface {
	Call(c *Context)
	UseBefore(orderNum int, h HandlerFunc)
	FindBefore() []HandlerFunc
	RoleLabel() string
}

type Role struct {
	roleLabel    string
	beforeHandle []HandlerFunc
}

func NewRole(roleLabel string) *Role {
	return &Role{
		roleLabel:    roleLabel,
		beforeHandle: make([]HandlerFunc, 0),
	}
}

func (r *Role) RoleLabel() string {
	return r.roleLabel
}

func (r *Role) UseBefore(h ...HandlerFunc) {
	r.beforeHandle = append(r.beforeHandle, h...)
}

func (r *Role) FindBefore() []HandlerFunc {
	return r.beforeHandle
}

func (r *Role) Call(c *Context) {
	if r.beforeHandle != nil {
		for _, h := range r.beforeHandle {
			h(c)
		}
	}
}
