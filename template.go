package main

var extpointsTemplate = `// generated by go-extpoints -- DO NOT EDIT
package {{.Package}}

import (
	"reflect"
	"sync"
)

var registry = struct {
	sync.Mutex
	extpoints map[string]*extensionPoint
}{
	extpoints: make(map[string]*extensionPoint),
}

type extensionPoint struct {
	sync.Mutex
	iface      reflect.Type
	components map[string]interface{}
}

func newExtensionPoint(iface interface{}) *extensionPoint {
	ep := &extensionPoint{
		iface:      reflect.TypeOf(iface).Elem(),
		components: make(map[string]interface{}),
	}
	registry.Lock()
	defer registry.Unlock()
	registry.extpoints[ep.iface.Name()] = ep
	return ep
}

func (ep *extensionPoint) lookup(name string) (ext interface{}, ok bool) {
	ep.Lock()
	defer ep.Unlock()
	ext, ok = ep.components[name]
	return
}

func (ep *extensionPoint) all() map[string]interface{} {
	ep.Lock()
	defer ep.Unlock()
	all := make(map[string]interface{})
	for k, v := range ep.components {
		all[k] = v
	}
	return all
}

func (ep *extensionPoint) register(component interface{}, name string) bool {
	ep.Lock()
	defer ep.Unlock()
	if name == "" {
		name = reflect.TypeOf(component).Elem().Name()
	}
	_, exists := ep.components[name]
	if exists {
		return false
	}
	ep.components[name] = component
	return true
}

func (ep *extensionPoint) unregister(name string) bool {
	ep.Lock()
	defer ep.Unlock()
	_, exists := ep.components[name]
	if !exists {
		return false
	}
	delete(ep.components, name)
	return true
}

func implements(component interface{}) []string {
	var ifaces []string
	for name, ep := range registry.extpoints {
		if reflect.TypeOf(component).Implements(ep.iface) {
			ifaces = append(ifaces, name)
		}
	}
	return ifaces
}

func Register(component interface{}, name string) []string {
	registry.Lock()
	defer registry.Unlock()
	var ifaces []string
	for _, iface := range implements(component) {
		if ok := registry.extpoints[iface].register(component, name); ok {
			ifaces = append(ifaces, iface)
		}
	}
	return ifaces
}

func Unregister(name string) []string {
	registry.Lock()
	defer registry.Unlock()
	var ifaces []string
	for iface, extpoint := range registry.extpoints {
		if ok := extpoint.unregister(name); ok {
			ifaces = append(ifaces, iface)
		}
	}
	return ifaces
}

{{range .ExtensionPoints}}// {{.Name}}

var {{.Var}} = &{{.Type}}{
	newExtensionPoint(new({{.Name}})),
}

type {{.Type}} struct {
	*extensionPoint
}

func (ep *{{.Type}}) Unregister(name string) bool {
	return ep.unregister(name)
}

func (ep *{{.Type}}) Register(component {{.Name}}, name string) bool {
	return ep.register(component, name)
}

func (ep *{{.Type}}) Lookup(name string) ({{.Name}}, bool) {
	ext, ok := ep.lookup(name)
	return ext.({{.Name}}), ok
}

func (ep *{{.Type}}) All() map[string]{{.Name}} {
	all := make(map[string]{{.Name}})
	for k, v := range ep.all() {
		all[k] = v.({{.Name}})
	}
	return all
}

{{end}}`
