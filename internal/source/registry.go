package source

import "fmt"

type AdapterFactory func(name string, cfg map[string]any) (SourceAdapter, error)

var registry = map[string]AdapterFactory{}

func Register(kind string, factory AdapterFactory) {
	registry[kind] = factory
}

func Get(kind string) (AdapterFactory, bool) {
	f, ok := registry[kind]
	return f, ok
}

func Create(kind, name string, cfg map[string]any) (SourceAdapter, error) {
	f, ok := Get(kind)
	if !ok {
		return nil, fmt.Errorf("unknown source kind: %s", kind)
	}
	return f(name, cfg)
}

func Available() []string {
	kinds := make([]string, 0, len(registry))
	for k := range registry {
		kinds = append(kinds, k)
	}
	return kinds
}
