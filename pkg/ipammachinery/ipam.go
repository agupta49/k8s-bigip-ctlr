/*-
* Copyright (c) 2016-2020, F5 Networks, Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package ipammachinery

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/F5Networks/k8s-bigip-ctlr/config/apis/cis/v1"
	log "github.com/F5Networks/k8s-bigip-ctlr/pkg/vlogger"

	// apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	// apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	// apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	// apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	// apiextension "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	// "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	// v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
)

const (

	// F5IPAM is a F5 Custom Resource Kind.
	F5ipam = "F5IPAM"

	CRDPlural   string = "f5ipams"
	CRDGroup    string = "cis.f5.com"
	CRDVersion  string = "v1"
	FullCRDName string = CRDPlural + "." + CRDGroup
)

// NewIPAM creates a new IPAM Instance.
func NewIPAM(params Params) *IPAM {

	ipamMgr := &IPAM{
		namespaces:    make(map[string]bool),
		ipamInformers: make(map[string]*IPAMInformer),
		rscQueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(), "custom-resource-controller"),
	}

	if err := ipamMgr.setupInformers(); err != nil {
		log.Error("Failed to Setup Informers")
	}

	go ipamMgr.Start()
	return ipamMgr
}

func (ipamMgr *IPAM) setupInformers() error {
	for n, _ := range ipamMgr.namespaces {
		if err := ipamMgr.addNamespacedInformer(n); err != nil {
			log.Errorf("Unable to setup informer for namespace: %v, Error:%v", n, err)
		}
	}
	return nil
}

// Start the Custom Resource Manager
func (ipamMgr *IPAM) Start() {
	log.Debug("Starting")
}

func (ipamMgr *IPAM) Init() {
	var config *rest.Config
	var err error
	if config, err = rest.InClusterConfig(); err != nil {
		log.Errorf("error creating client configuration: %v", err)
	}
	kubeClient, err := apiextension.NewForConfig(config)
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
	}
	// Create the CRD
	err = CreateCRD(kubeClient)
	if err != nil {
		log.Errorf("Failed to create crd: %v", err)
	}

	// Wait for the CRD to be created before we use it.
	time.Sleep(5 * time.Second)

	// Create a new clientset which include our CRD schema
	crdclient, err := NewClient(config)
	if err != nil {
		panic(err)
	}

	// Create a new SslConfig object

	f5ipam := &v1.F5IPAM{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "f5ipam",
		},
		Spec:   v1.F5IPAMSpec{},
		Status: v1.F5IPAMStatus{},
	}
	// Create the SslConfig object we create above in the k8s cluster
	resp, err := crdclient.F5IPAMS("default").Create(f5ipam)
	if err != nil {
		fmt.Printf("error while creating object: %v\n", err)
	} else {
		fmt.Printf("object created: %v\n", resp)
	}

	obj, err := crdclient.F5IPAMS("default").Get(f5ipam.ObjectMeta.Name)
	if err != nil {
		log.Infof("error while getting the object %v\n", err)
	}
	fmt.Printf("SslConfig Objects Found: \n%v\n", obj)

}

func CreateCRD(clientset apiextension.Interface) error {
	crd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: meta_v1.ObjectMeta{Name: FullCRDName},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group:   CRDGroup,
			Version: CRDVersion,
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Plural: CRDPlural,
				Kind:   F5ipam,
			},
		},
	}
	ctx := context.Background()
	var opts meta_v1.CreateOptions
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(ctx, crd, opts)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
