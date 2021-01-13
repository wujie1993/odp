package tpls

const (
	CLIENT_CODE_TPL = `
{{ $package := .Package }}
package {{ .Package }}

import (
	"context"

	"github.com/wujie1993/waves/pkg/client/rest"
	obj{{ .Package }} "github.com/wujie1993/waves/pkg/orm/{{ .Package }}"
)

type Client struct {
	rest.RESTClient
}
{{ range .Registries }}
func (c Client) {{ .Name }}s({{ if .Namespaced }}namespace string{{ end }}) {{ ToLower .Name }}s {
	return {{ ToLower .Name }}s{
		{{- if .Namespaced }}
		namespace:  namespace,
		{{- end }}
		RESTClient: c.RESTClient,
	}
}
{{ end }}
func NewClient(cli rest.RESTClient) Client {
	return Client{
		RESTClient: cli,
	}
}
{{ range .Registries }}
type {{ ToLower .Name }}s struct {
	rest.RESTClient
	namespace string
}

func (c {{ ToLower .Name }}s) Get(ctx context.Context, name string) (*obj{{ $package }}.{{ .Name }}, error) {
	result := &obj{{ $package }}.{{ .Name }}{}
	if err := c.RESTClient.Get().
		Version("{{ $package }}").
		{{- if .Namespaced }}
		Namespace(c.namespace).
		{{- end }}
		Resource("{{ ToLower .Name }}s").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c {{ ToLower .Name }}s) Create(ctx context.Context, obj *obj{{ $package }}.{{ .Name }}) (*obj{{ $package }}.{{ .Name }}, error) {
	result := &obj{{ $package }}.{{ .Name }}{}
	if err := c.RESTClient.Post().
		Version("{{ $package }}").
		{{- if .Namespaced }}
		Namespace(c.namespace).
		{{- end }}
		Resource("{{ ToLower .Name }}s").
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c {{ ToLower .Name }}s) List(ctx context.Context) ([]obj{{ $package }}.{{ .Name }}, error) {
	result := []obj{{ $package }}.{{ .Name }}{}
	if err := c.RESTClient.Get().
		Version("{{ $package }}").
		{{- if .Namespaced }}
		Namespace(c.namespace).
		{{- end }}
		Resource("{{ ToLower .Name }}s").
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c {{ ToLower .Name }}s) Update(ctx context.Context, obj *obj{{ $package }}.{{ .Name }}) (*obj{{ $package }}.{{ .Name }}, error) {
	result := &obj{{ $package }}.{{ .Name }}{}
	if err := c.RESTClient.Put().
		Version("{{ $package }}").
		{{- if .Namespaced }}
		Namespace(c.namespace).
		{{- end }}
		Resource("{{ ToLower .Name }}s").
		Name(obj.Metadata.Name).
		Data(obj).
		Do(ctx).
		Into(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c {{ ToLower .Name }}s) Delete(ctx context.Context, name string) (*obj{{ $package }}.{{ .Name }}, error) {
	result := &obj{{ $package }}.{{ .Name }}{}
	if err := c.RESTClient.Delete().
		Version("{{ $package }}").
		{{- if .Namespaced }}
		Namespace(c.namespace).
		{{- end }}
		Resource("{{ ToLower .Name }}s").
		Name(name).
		Do(ctx).
		Into(result); err != nil {
		return nil, err
	}
	return result, nil
}
{{ end }}
`
)
