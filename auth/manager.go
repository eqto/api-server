package auth

//Manager ...
type Manager struct {
	authMap map[string]Interface
}

//Get ...
func (m *Manager) Get(name string) Interface {
	if auth, ok := m.authMap[name]; ok {
		return auth
	}
	return nil
}

//Set ...
func (m *Manager) Set(name string, auth Interface) {
	m.authMap[name] = auth
}

func newManager() *Manager {
	return &Manager{authMap: make(map[string]Interface)}
}
