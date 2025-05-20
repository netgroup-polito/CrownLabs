package namespaces

import (
    "context"
    "time"

    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/klog/v2"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
    "github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

type NamespaceManager struct {
    client        client.Client
    scheme        *runtime.Scheme
    keepAliveTime time.Duration
    targetLabelKey string
    targetLabelValue string
}

// NewNamespaceManager creates a new NamespaceManager instance
func NewNamespaceManager(
    client client.Client,
    scheme *runtime.Scheme,
    keepAliveTime time.Duration,
    targetLabelKey string,
    targetLabelValue string,
) *NamespaceManager {
    return &NamespaceManager{
        client:          client,
        scheme:          scheme,
        keepAliveTime:   keepAliveTime,
        targetLabelKey:  targetLabelKey,
        targetLabelValue: targetLabelValue,
    }
}

// checkNamespaceKeepAlive checks to see if the namespace should be deleted.
func (nm *NamespaceManager) CheckNamespaceKeepAlive(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) (keepNsOpen bool, err error) {
    sPassed := time.Since(tn.Spec.LastLogin.Time)
    klog.Infof("Last login of tenant %s was %s ago", tn.Name, sPassed)

    list := &crownlabsv1alpha2.InstanceList{}
    if err := nm.client.List(ctx, list, client.InNamespace(nsName)); err != nil {
        return true, err
    }

    if sPassed > nm.keepAliveTime {
        klog.Infof("Over %s elapsed since last login of tenant %s: tenant namespace shall be absent", nm.keepAliveTime, tn.Name)
        if len(list.Items) == 0 {
            klog.Infof("No instances found in %s: namespace can be deleted", nsName)
            return false, nil
        }
        klog.Infof("Instances found in namespace %s. Namespace will not be deleted", nsName)
    } else {
        klog.Infof("Under %s (limit) elapsed since last login of tenant %s: tenant namespace shall be present", nm.keepAliveTime, tn.Name)
    }
    return true, nil
}

// deleteClusterNamespace deletes the namespace for the tenant
func (nm *NamespaceManager) DeleteNamespace(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string) error {
    ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
    err := utils.EnforceObjectAbsence(ctx, nm.client, &ns, "personal namespace")
    if err != nil {
        klog.Errorf("Error when deleting namespace of tenant %s -> %s", tn.Name, err)
    }
    return err
}

// updateNamespaceLabels updates the tenant namespace labels
func (nm *NamespaceManager) updateNamespaceLabels(ns *v1.Namespace, tnName string) {
    if ns.Labels == nil {
        ns.Labels = make(map[string]string)
    }
    ns.Labels[nm.targetLabelKey] = nm.targetLabelValue
    ns.Labels["crownlabs.polito.it/type"] = "tenant"
    ns.Labels["crownlabs.polito.it/name"] = tnName
    ns.Labels["crownlabs.polito.it/instance-resources-replication"] = "true"
    ns.Labels["crownlabs.polito.it/managed-by"] = "tenant"
}

// enforceClusterResources manages namespace lifecycle
func (nm *NamespaceManager) EnforceClusterResources(ctx context.Context, tn *crownlabsv1alpha2.Tenant, nsName string, keepNsOpen bool) (bool, error) {
    if keepNsOpen {
        ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
        if _, err := ctrl.CreateOrUpdate(ctx, nm.client, &ns, func() error {  //TODO replace CreateOrUpdate with createOrUpdateClusterResources
            nm.updateNamespaceLabels(&ns, tn.Name)
            return ctrl.SetControllerReference(tn, &ns, nm.scheme)
        }); err != nil {
            klog.Errorf("Error when updating namespace of tenant %s -> %s", tn.Name, err)
            tn.Status.PersonalNamespace.Created = false
            tn.Status.PersonalNamespace.Name = ""
            return false, err
        }
        
        tn.Status.PersonalNamespace.Created = true
        tn.Status.PersonalNamespace.Name = nsName
        return true, nil
    }

    if err := nm.DeleteNamespace(ctx, tn, nsName); err != nil {
        return false, err
    }

    tn.Status.PersonalNamespace.Created = false
    tn.Status.PersonalNamespace.Name = ""
    return false, nil
}