package routing

import (
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/commands"
)

func NewRouteList(app contracts.Application) contracts.Command {
	return &RouteList{
		Command: commands.Base("route:list", "打印路由列表"),
		router:  app.Get("HttpRouter").(contracts.HttpRouter),
	}
}

type RouteList struct {
	commands.Command
	router contracts.HttpRouter
}

func (cmd RouteList) Handle() any {
	cmd.router.Print()
	return nil
}
