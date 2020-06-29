package web

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/micro/go-micro/v2/api/resolver"
	res "github.com/micro/go-micro/v2/api/resolver"
	"github.com/micro/go-micro/v2/router"
	// "github.com/micro/go-micro/v2/selector"
)

var re = regexp.MustCompile("^[a-zA-Z0-9]+([a-zA-Z0-9-]*[a-zA-Z0-9]*)?$")

type Resolver struct {
	// Options
	Options resolver.Options
	// selector to choose from a pool of nodes
	// Selector selector.Selector
	// router to lookup routes
	Router router.Router
}

func (r *Resolver) String() string {
	return "web/resolver"
}

// Resolve replaces the values of Host, Path, Scheme to calla backend service
// It accounts for subdomains for service names based on namespace
func (r *Resolver) Resolve(req *http.Request, opts ...res.ResolveOption) (*res.Endpoint, error) {
	// parse the options
	options := resolver.NewResolveOptions(opts...)

	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 2 {
		return nil, errors.New("unknown service")
	}

	if !re.MatchString(parts[1]) {
		return nil, res.ErrInvalidPath
	}

	// lookup the routes for the service
	query := []router.QueryOption{
		router.QueryService(r.Options.ServicePrefix + "." + parts[1]),
		router.QueryNetwork(options.Domain),
	}
	routes, err := r.Router.Lookup(query...)
	if err != nil {
		return nil, err
	}

	// select the route to use
	// todo: update to use selector once go-micro has updated the interface
	// route, err := r.Selector.Select(routes...)
	// if err != nil {
	// 	return nil, err
	// }
	route := routes[0]

	// we're done
	return &res.Endpoint{
		Name:   parts[1],
		Method: req.Method,
		Host:   route.Address,
		Path:   "/" + strings.Join(parts[2:], "/"),
		Domain: route.Network,
	}, nil
}
