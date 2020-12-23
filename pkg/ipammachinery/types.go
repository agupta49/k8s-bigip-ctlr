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
	"github.com/F5Networks/k8s-bigip-ctlr/config/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type (
	// CRManager defines the structure of Custom Resource Manager
	IPAM struct {
		kubeCRClient  versioned.Interface
		kubeClient    kubernetes.Interface
		ipamInformers map[string]*IPAMInformer
		nsInformer    *NSInformer
		namespaces    map[string]bool
		Partition     string
		rscQueue      workqueue.RateLimitingInterface
	}
	// Params defines parameters
	Params struct {
		Config         *rest.Config
		Namespaces     []string
		NamespaceLabel string
		Partition      string
	}
	// CRInformer defines the structure of Custom Resource Informer
	IPAMInformer struct {
		namespace    string
		stopCh       chan struct{}
		ipamInformer cache.SharedIndexInformer
	}

	NSInformer struct {
		stopCh     chan struct{}
		nsInformer cache.SharedIndexInformer
	}
	rqKey struct {
		namespace string
		kind      string
		rscName   string
		rsc       interface{}
		rscDelete bool
	}
)
