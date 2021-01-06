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
	"time"

	cisapiv1 "github.com/F5Networks/k8s-bigip-ctlr/config/apis/cis/v1"
	cisinfv1 "github.com/F5Networks/k8s-bigip-ctlr/config/client/informers/externalversions/cis/v1"
	log "github.com/F5Networks/k8s-bigip-ctlr/pkg/vlogger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// start the ipam informer
func (ipamInfr *IPAMInformer) start() {
	var cacheSyncs []cache.InformerSynced

	if ipamInfr.ipamInformer != nil {
		log.Infof("Starting IPAM Informer")
		go ipamInfr.ipamInformer.Run(ipamInfr.stopCh)
		cacheSyncs = append(cacheSyncs, ipamInfr.ipamInformer.HasSynced)
	}

	cache.WaitForNamedCacheSync(
		"F5 CIS CRD Controller",
		ipamInfr.stopCh,
		cacheSyncs...,
	)
}

func (ipamInfr *IPAMInformer) stop() {
	close(ipamInfr.stopCh)
}

func (ipamMgr *IPAM) watchingAllNamespaces() bool {
	if 0 == len(ipamMgr.ipamInformers) {
		// Not watching any namespaces.
		return false
	}
	_, watchingAll := ipamMgr.ipamInformers[""]
	return watchingAll
}

func (ipamMgr *IPAM) addNamespacedInformer(
	namespace string,
) error {
	if ipamMgr.watchingAllNamespaces() {
		return fmt.Errorf(
			"Cannot add additional namespaces when already watching all.")
	}
	if len(ipamMgr.ipamInformers) > 0 && "" == namespace {
		return fmt.Errorf(
			"Cannot watch all namespaces when already watching specific ones.")
	}
	var crInf *IPAMInformer
	var found bool
	if crInf, found = ipamMgr.ipamInformers[namespace]; found {
		return nil
	}
	crInf = ipamMgr.newNamespacedInformer(namespace)
	ipamMgr.addEventHandlers(crInf)
	ipamMgr.ipamInformers[namespace] = crInf
	return nil
}

func (ipamMgr *IPAM) newNamespacedInformer(
	namespace string,
) *IPAMInformer {
	log.Debugf("[ipam] Creating Informers for Namespace %v", namespace)
	everything := func(options *metav1.ListOptions) {
		options.LabelSelector = ""
	}

	resyncPeriod := 0 * time.Second
	// restClientv1 := ipamMgr.kubeClient.CoreV1().RESTClient()

	ipamInf := &IPAMInformer{
		namespace: namespace,
		stopCh:    make(chan struct{}),
	}

	ipamInf.ipamInformer = cisinfv1.NewFilteredF5IPAMInformer(
		ipamMgr.kubeCRClient,
		namespace,
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		everything,
	)

	return ipamInf
}

func (ipamMgr *IPAM) addEventHandlers(ipamInf *IPAMInformer) {
	if ipamInf.ipamInformer != nil {
		ipamInf.ipamInformer.AddEventHandler(
			&cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) { ipamMgr.enqueueIPAM(obj) },
				UpdateFunc: func(old, cur interface{}) { ipamMgr.enqueueUpdatedIPAM(old, cur) },
				DeleteFunc: func(obj interface{}) { ipamMgr.enqueueDeletedIPAM(obj) },
			},
		)
	}
}

func (ipamMgr *IPAM) getNamespacedInformer(
	namespace string,
) (*IPAMInformer, bool) {
	if ipamMgr.watchingAllNamespaces() {
		namespace = ""
	}
	ipamInf, found := ipamMgr.ipamInformers[namespace]
	return ipamInf, found
}

func (ipamMgr *IPAM) enqueueIPAM(obj interface{}) {
	ipam := obj.(*cisapiv1.F5IPAM)
	key := &rqKey{
		namespace: ipam.ObjectMeta.Namespace,
		kind:      F5ipam,
		rscName:   ipam.ObjectMeta.Name,
		rsc:       obj,
	}

	ipamMgr.rscQueue.Add(key)
}

func (ipamMgr *IPAM) enqueueUpdatedIPAM(oldObj, newObj interface{}) {
	oldIPAM := oldObj.(*cisapiv1.F5IPAM)
	newIPAM := newObj.(*cisapiv1.F5IPAM)

	log.Debugf("Enqueueing ipam: %v", newIPAM)
	log.Debugf("Old ipam: %v", oldIPAM)
	key := &rqKey{
		namespace: newIPAM.ObjectMeta.Namespace,
		kind:      F5ipam,
		rscName:   newIPAM.ObjectMeta.Name,
		rsc:       newObj,
	}
	ipamMgr.rscQueue.Add(key)
}

func (ipamMgr *IPAM) enqueueDeletedIPAM(obj interface{}) {
	ipam := obj.(*cisapiv1.F5IPAM)
	log.Debugf("Enqueueing ipam: %v", ipam)
	key := &rqKey{
		namespace: ipam.ObjectMeta.Namespace,
		kind:      F5ipam,
		rscName:   ipam.ObjectMeta.Name,
		rsc:       obj,
		rscDelete: true,
	}
	ipamMgr.rscQueue.Add(key)
}
