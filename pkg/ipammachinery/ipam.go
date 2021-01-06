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
	"fmt"

	v1 "github.com/F5Networks/k8s-bigip-ctlr/config/apis/cis/v1"
	log "github.com/F5Networks/k8s-bigip-ctlr/pkg/vlogger"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
)

const (

	// F5IPAM is a F5 Custom Resource Kind.
	F5ipam = "F5IPAM"

	//CRDPlural   string = "f5ipams"
	CRDGroup    string = "cis.f5.com"
	CRDVersion  string = "v1"
	//FullCRDName string = CRDPlural + "." + CRDGroup
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
	log.Debugf("Amit: IPAM New")
	go ipamMgr.Start()
	return ipamMgr
}

func (ipamMgr *IPAM) setupInformers() error {
	if err := ipamMgr.addNamespacedInformer("default"); err != nil {
		log.Errorf("Unable to setup informer for namespace: %v, Error:%v", "default", err)
	}
	return nil
}

// Start the Custom Resource Manager
func (ipamMgr *IPAM) Start() {
	log.Debug("[ipam] Starting")
}

func (ipamMgr *IPAM) Init() {
	log.Debugf("[ipam] Init")
	var config *rest.Config
	var err error
	if config, err = rest.InClusterConfig(); err != nil {
		log.Errorf("[ipam] error creating client configuration: %v", err)
	}
	// kubeClient, err := apiextension.NewForConfig(config)
	// if err != nil {
	// 	log.Errorf("Failed to create client: %v", err)
	// }
	// Create the CRD
	// err = CreateCRD(kubeClient)
	// if err != nil {
	// 	log.Errorf("Failed to create crd: %v", err)
	// }

	// // Wait for the CRD to be created before we use it.
	// time.Sleep(5 * time.Second)

	// Create a new clientset which include our CRD schema
	crdclient, err := NewClient(config)
	if err != nil {
		panic(err)
	}

	// Create a new SslConfig object

	// f5ipam := &v1.F5IPAM{
	// 	ObjectMeta: meta_v1.ObjectMeta{
	// 		Name: "f5ipam",
	// 	},
	// 	Spec:   v1.F5IPAMSpec{},
	// 	Status: v1.F5IPAMStatus{},
	// }
	// Create the SslConfig object we create above in the k8s cluster
	// resp, err := crdclient.F5IPAMS("default").Create(f5ipam)
	// if err != nil {
	// 	fmt.Printf("error while creating object: %v\n", err)
	// } else {
	// 	fmt.Printf("object created: %v\n", resp)
	// }

	obj, err := crdclient.F5IPAMS("default").Get("f5ipam")
	if err != nil {
		log.Infof("[ipam] Error while getting the object %v\n", err)
	}

	f5ipam := &v1.F5IPAM{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "f5ipam",
		},
		Spec: obj.Spec,
		Status: v1.F5IPAMStatus{
			HostNames: []v1.HostNameIP{
				{
					Host: "foo.com",
					IP:   "1.1.1.1",
				},
			},
		},
	}
	f5ipam.SetResourceVersion(obj.ResourceVersion)
	obj1, err1 := crdclient.F5IPAMS("default").Update(f5ipam)
	if err1 != nil {
		log.Infof("[ipam] error while updating the object %v\n", err1)
	}
	fmt.Printf("[ipam] Updating Objects Found: \n%v\n", obj1)
}
