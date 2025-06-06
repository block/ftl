package deploymentcontext

import (
	"os"
)

var _ RouteProvider = (*templateRouteTable)(nil)

func NewTemplateRouter(template string, realm string) RouteProvider {
	return &templateRouteTable{template: template, realm: realm}
}

type templateRouteTable struct {
	template string
	realm    string
}

// Route implements deploymentcontext.RouteProvider.
func (t *templateRouteTable) Route(module string) string {
	return os.Expand(t.template, func(s string) string {
		switch s {
		case "module":
			return module
		case "realm":
			return t.realm
		}
		return ""
	})
}

// Subscribe implements deploymentcontext.RouteProvider.
func (t *templateRouteTable) Subscribe() chan string {
	return make(chan string)
}

// Unsubscribe implements deploymentcontext.RouteProvider.
func (t *templateRouteTable) Unsubscribe(c chan string) {
	close(c)
}
