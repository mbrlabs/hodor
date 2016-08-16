package hodor

import (
	"net/http"
	"strings"
)

// ============================================================================
//                              struct node
// ============================================================================
type node struct {
	part  string
	route *Route
	nodes []*node
}

func (n *node) nextNode(part string) *node {
	isParam := strings.HasPrefix(part, ":")
	for _, childNode := range n.nodes {
		// panic if a different named parameter is added to a list of child nodes, which already
		// have another named parameter
		if strings.HasPrefix(childNode.part, ":") && isParam && childNode.part != part {
			panic("Ambigious mapping of named parameters found.")
		}

		if childNode.part == part {
			return childNode
		}
	}

	newNode := &node{part: part, route: nil}
	n.nodes = append(n.nodes, newNode)
	return newNode
}

// ============================================================================
//                              struct RouteTree
// ============================================================================
type RouteTree struct {
	treeRoots map[string]*node
}

func NewRouteTree() RouteTree {
	var roots map[string]*node = make(map[string]*node)
	roots[http.MethodGet] = &node{part: "", route: nil}
	roots[http.MethodHead] = &node{part: "", route: nil}
	roots[http.MethodPost] = &node{part: "", route: nil}
	roots[http.MethodPut] = &node{part: "", route: nil}
	roots[http.MethodDelete] = &node{part: "", route: nil}
	roots[http.MethodOptions] = &node{part: "", route: nil}
	return RouteTree{treeRoots: roots}
}

func (t *RouteTree) InsertRoute(route *Route) {
	// get tree root, corresponding to the http method
	root := t.treeRoots[route.Method]
	if root == nil {
		panic("Unsupported http method: " + route.Method)
	}

	parts := strings.Split(route.GetPath(), "/")
	// handle the root pattern ("")
	if len(parts) == 1 && parts[0] == "" {
		if root.route == nil {
			root.route = route
			return
		} else {
			panic("Abmigious mapping for the root pattern")
		}
	}

	// handle other patterns
	currentNode := root
	for i, part := range parts {
		currentNode = currentNode.nextNode(part)
		// end of pattern reached. we need to assign the route here
		if i == len(parts)-1 {
			if currentNode.route == nil {
				currentNode.route = route
				return
			} else {
				panic("Abmigious mapping for: " + route.path)
			}
		}
	}

	panic("Something went terribly wrong while mapping the route: " + route.path)
}

// Returns a route and sets the url parameters of the context
func (t *RouteTree) GetRoute(ctx *Context) *Route {
	// get tree root, corresponding to the http method
	root := t.treeRoots[ctx.Request.Method]
	if root == nil {
		return nil
	}

	path := strings.Trim(ctx.Request.URL.Path, "/")

	// deny everything that contains a colon
	if strings.Contains(path, ":") {
		return nil
	}

	// handle root path
	if path == "" {
		return root.route
	}

	// handle everything else
	parts := strings.Split(path, "/")
	var currentNode *node = root
	var namedParam *node = nil
	foundPart := false
	for _, part := range parts {
		foundPart = false
		namedParam = nil
		for _, childNode := range currentNode.nodes {
			// found a static match
			if childNode.part == part {
				currentNode = childNode
				foundPart = true
				break
			}
			// found a possible match in a named param
			if strings.Contains(childNode.part, ":") {
				namedParam = childNode
			}
		}

		// found match in named param, since it's set to a non-nil value.
		// set as current node, extract value and set in context.
		if namedParam != nil && !foundPart {
			currentNode = namedParam
			foundPart = true
			key := strings.TrimLeft(currentNode.part, ":")
			ctx.UrlParams[key] = part
		}

		// if the url part can't be found, the route does not exist
		if !foundPart {
			return nil
		}
	}

	return currentNode.route
}
