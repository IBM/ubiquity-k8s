package hookexecutor

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
)

/**
 * decoder is a tool to convert a yaml/json manifest to a k8s object.
 */

type Decode func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)

type decoder struct {
	decode Decode
}

func (d *decoder) FromJson(json []byte) (runtime.Object, error) {
	return d.From(json)
}

func (d *decoder) FromYaml(yaml []byte) (runtime.Object, error) {
	return d.From(yaml)
}

func (d *decoder) From(data []byte) (runtime.Object, error) {
	obj, _, err := d.decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

var KubeDecoder *decoder = &decoder{decode: kubescheme.Codecs.UniversalDeserializer().Decode}

func FromJson(json []byte) (runtime.Object, error) {
	return KubeDecoder.FromJson(json)
}

func FromYaml(yaml []byte) (runtime.Object, error) {
	return KubeDecoder.FromYaml(yaml)
}
