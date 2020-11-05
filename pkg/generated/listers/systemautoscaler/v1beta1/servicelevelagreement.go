// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/lterrac/system-autoscaler/pkg/apis/systemautoscaler/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ServiceLevelAgreementLister helps list ServiceLevelAgreements.
// All objects returned here must be treated as read-only.
type ServiceLevelAgreementLister interface {
	// List lists all ServiceLevelAgreements in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.ServiceLevelAgreement, err error)
	// ServiceLevelAgreements returns an object that can list and get ServiceLevelAgreements.
	ServiceLevelAgreements(namespace string) ServiceLevelAgreementNamespaceLister
	ServiceLevelAgreementListerExpansion
}

// serviceLevelAgreementLister implements the ServiceLevelAgreementLister interface.
type serviceLevelAgreementLister struct {
	indexer cache.Indexer
}

// NewServiceLevelAgreementLister returns a new ServiceLevelAgreementLister.
func NewServiceLevelAgreementLister(indexer cache.Indexer) ServiceLevelAgreementLister {
	return &serviceLevelAgreementLister{indexer: indexer}
}

// List lists all ServiceLevelAgreements in the indexer.
func (s *serviceLevelAgreementLister) List(selector labels.Selector) (ret []*v1beta1.ServiceLevelAgreement, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ServiceLevelAgreement))
	})
	return ret, err
}

// ServiceLevelAgreements returns an object that can list and get ServiceLevelAgreements.
func (s *serviceLevelAgreementLister) ServiceLevelAgreements(namespace string) ServiceLevelAgreementNamespaceLister {
	return serviceLevelAgreementNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ServiceLevelAgreementNamespaceLister helps list and get ServiceLevelAgreements.
// All objects returned here must be treated as read-only.
type ServiceLevelAgreementNamespaceLister interface {
	// List lists all ServiceLevelAgreements in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.ServiceLevelAgreement, err error)
	// Get retrieves the ServiceLevelAgreement from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.ServiceLevelAgreement, error)
	ServiceLevelAgreementNamespaceListerExpansion
}

// serviceLevelAgreementNamespaceLister implements the ServiceLevelAgreementNamespaceLister
// interface.
type serviceLevelAgreementNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ServiceLevelAgreements in the indexer for a given namespace.
func (s serviceLevelAgreementNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.ServiceLevelAgreement, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ServiceLevelAgreement))
	})
	return ret, err
}

// Get retrieves the ServiceLevelAgreement from the indexer for a given namespace and name.
func (s serviceLevelAgreementNamespaceLister) Get(name string) (*v1beta1.ServiceLevelAgreement, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("servicelevelagreement"), name)
	}
	return obj.(*v1beta1.ServiceLevelAgreement), nil
}