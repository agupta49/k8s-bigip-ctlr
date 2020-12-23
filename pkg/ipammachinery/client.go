package ipammachinery

import (
	"context"

	v1 "github.com/F5Networks/k8s-bigip-ctlr/config/apis/cis/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

func (c *F5IPAMConfigV1Client) F5IPAMS(namespace string) F5IPAMConfigInterface {
	return &ipamclient{
		client: c.restClient,
		ns:     namespace,
	}
}

type F5IPAMConfigV1Client struct {
	restClient rest.Interface
}

type F5IPAMConfigInterface interface {
	Create(obj *v1.F5IPAM) (*v1.F5IPAM, error)
	Update(obj *v1.F5IPAM) (*v1.F5IPAM, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.F5IPAM, error)
}

type ipamclient struct {
	client rest.Interface
	ns     string
}

func (c *ipamclient) Create(obj *v1.F5IPAM) (*v1.F5IPAM, error) {
	result := &v1.F5IPAM{}
	ctx := context.Background()
	err := c.client.Post().
		Namespace(c.ns).Resource("f5ipams").
		Body(obj).Do(ctx).Into(result)
	return result, err
}

func (c *ipamclient) Update(obj *v1.F5IPAM) (*v1.F5IPAM, error) {
	result := &v1.F5IPAM{}
	ctx := context.Background()
	err := c.client.Put().
		Namespace(c.ns).Resource("f5ipams").
		Body(obj).Do(ctx).Into(result)
	return result, err
}

func (c *ipamclient) Delete(name string, options *meta_v1.DeleteOptions) error {
	ctx := context.Background()
	return c.client.Delete().
		Namespace(c.ns).Resource("f5ipams").
		Name(name).Body(options).Do(ctx).Error()
}

func (c *ipamclient) Get(name string) (*v1.F5IPAM, error) {
	result := &v1.F5IPAM{}
	ctx := context.Background()
	err := c.client.Get().
		Namespace(c.ns).Resource("f5ipams").
		Name(name).Do(ctx).Into(result)
	return result, err
}

var SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&v1.F5IPAM{},
		&v1.F5IPAMList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func NewClient(cfg *rest.Config) (*F5IPAMConfigV1Client, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &F5IPAMConfigV1Client{restClient: client}, nil
}
