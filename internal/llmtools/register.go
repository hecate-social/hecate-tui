package llmtools

// RegisterAllTools registers all built-in tools to the registry.
func RegisterAllTools(r *Registry) {
	RegisterFileSystemTools(r)
	RegisterCodeExploreTools(r)
	RegisterSystemTools(r)
	RegisterWebSearchTools(r)
	RegisterMeshTools(r)
}

// NewDefaultRegistry creates a registry with all built-in tools registered.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	RegisterAllTools(r)
	return r
}
