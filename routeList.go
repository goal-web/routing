package routing

import (
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/commands"
)

func NewRouteList() (contracts.Command, contracts.CommandHandlerProvider) {
	return commands.Base("route:list", "打印路由列表"),
		func(app contracts.Application) contracts.CommandHandler {
			return &RouteList{
				router: app.Get("HttpRouter").(contracts.HttpRouter),
			}
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
