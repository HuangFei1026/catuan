package web

type GroupInf interface {
	GroupLabel() string
	RoleLabel() string
	Call(c *Context)

	UseBefore(h HandlerFunc, destAction ...string)
	FindAction(actionName string) (HandlerFunc, bool)
	BindAction(actionName string, handler HandlerFunc)
}

type Group struct {
	groupName      string
	roleName       string
	beforeHandlers map[string][]HandlerFunc
	actionHandlers map[string]HandlerFunc
	commHandlers   []HandlerFunc
}

func NewGroup(roleLabel string, groupLabel string) *Group {
	return &Group{
		groupName:      groupLabel,
		roleName:       roleLabel,
		beforeHandlers: make(map[string][]HandlerFunc),
		actionHandlers: make(map[string]HandlerFunc),
		commHandlers:   make([]HandlerFunc, 0),
	}
}

func (g *Group) GroupLabel() string {
	return g.groupName
}

func (g *Group) RoleLabel() string {
	return g.roleName
}

func (g *Group) Call(c *Context) {
	//执行Before handler
	if g.commHandlers != nil {
		for _, h := range g.commHandlers {
			h(c)
		}
	}
	if !c.IsNext() {
		return
	}
	//dest action before handlers
	if _, ok := g.beforeHandlers[c.ActionLabel()]; ok {
		for _, h := range g.beforeHandlers[c.ActionLabel()] {
			h(c)
		}
	}
	if !c.IsNext() {
		return
	}
	//dest action handler
	if h, ok := g.actionHandlers[c.ActionLabel()]; ok {
		h(c)
	} else {
		c.Result(-1, "action not found")
		c.AbortHandler()
		return
	}
}

func (g *Group) UseBefore(h HandlerFunc, destAction ...string) {
	if len(destAction) == 0 {
		g.commHandlers = append(g.commHandlers, h)
	} else {
		for _, action := range destAction {
			if _, ok := g.beforeHandlers[action]; !ok {
				g.beforeHandlers[action] = make([]HandlerFunc, 0)
			}
			g.beforeHandlers[action] = append(g.beforeHandlers[action], h)
		}
	}
}

func (g *Group) FindBefore(destAction string) []HandlerFunc {
	if _, ok := g.beforeHandlers[destAction]; !ok {
		return nil
	}
	return g.beforeHandlers[destAction]
}

func (g *Group) BindAction(action string, h HandlerFunc) {
	g.actionHandlers[action] = h
}

func (g *Group) FindAction(action string) (HandlerFunc, bool) {
	h, ok := g.actionHandlers[action]
	return h, ok
}
