package api

//Module ...
type Module map[string]*RoutePath

//Get ...
func (m Module) Get(method, path string) *RoutePath {
	if routePath, ok := m[method+` `+path]; ok {
		return routePath
	}
	return nil
}

//Set ...
func (m Module) Set(method, path string, routePath *RoutePath) {
	if routePath != nil {
		m[method+` `+path] = routePath
	} else {
		m.Unset(method, path)
	}
}

//Unset ...
func (m Module) Unset(method, path string) {
	delete(m, method+` `+path)
}
