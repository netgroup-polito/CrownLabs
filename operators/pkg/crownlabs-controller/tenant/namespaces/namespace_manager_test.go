package namespaces

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"

    crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

func TestNamespaceManager(t *testing.T) {
    // Create scheme
    scheme := runtime.NewScheme()
    _ = v1.AddToScheme(scheme)
    _ = crownlabsv1alpha2.AddToScheme(scheme)
    t.Log("Scheme created successfully")

    // Create fake client
    client := fake.NewClientBuilder().WithScheme(scheme).Build()
    t.Log("Fake client created successfully")

    // Create namespace manager
    manager := &NamespaceManager{
        client:           client,
        scheme:          scheme,
        keepAliveTime:   24 * time.Hour,
        targetLabelKey:  "test-key",
        targetLabelValue: "test-value",
    }
    t.Log("Namespace manager created successfully")

    // Create test tenant
    tenant := &crownlabsv1alpha2.Tenant{
        ObjectMeta: metav1.ObjectMeta{
            Name: "test-tenant",
        },
        Spec: crownlabsv1alpha2.TenantSpec{
            LastLogin: metav1.Time{Time: time.Now()},
        },
    }
    t.Logf("Test tenant created with name: %s", tenant.Name)

    ctx := context.Background()
    nsName := "tenant-test-tenant"

    // Test 1: Check keep-alive for recent login
    t.Log("Starting Test 1: Check keep-alive for recent login")
    keepAlive, err := manager.CheckNamespaceKeepAlive(ctx, tenant, nsName)
    assert.NoError(t, err)
    assert.True(t, keepAlive)
    t.Logf("Keep-alive check result: %v", keepAlive)

    // Test 2: Create namespace
    t.Log("Starting Test 2: Create namespace")
    ok, err := manager.EnforceClusterResources(ctx, tenant, nsName, true)
    assert.NoError(t, err)
    assert.True(t, ok)
    t.Logf("Namespace %s creation result: %v", nsName, ok)

    // Test 3: Verify namespace exists with correct labels
    t.Log("Starting Test 3: Verify namespace exists with correct labels")
    ns := &v1.Namespace{}
    err = client.Get(ctx, types.NamespacedName{Name: nsName}, ns)
    assert.NoError(t, err)
    assert.Equal(t, "tenant", ns.Labels["crownlabs.polito.it/type"])
    assert.Equal(t, tenant.Name, ns.Labels["crownlabs.polito.it/name"])
    assert.Equal(t, "true", ns.Labels["crownlabs.polito.it/instance-resources-replication"])
    assert.Equal(t, "tenant", ns.Labels["crownlabs.polito.it/managed-by"])
    assert.Equal(t, manager.targetLabelValue, ns.Labels[manager.targetLabelKey])
    t.Logf("Namespace %s exists with correct labels", nsName)

    // Test 4: Check keep-alive for old login
    t.Log("Starting Test 4: Check keep-alive for old login")
    tenant.Spec.LastLogin = metav1.Time{Time: time.Now().Add(-48 * time.Hour)}
    keepAlive, err = manager.CheckNamespaceKeepAlive(ctx, tenant, nsName)
    assert.NoError(t, err)
    assert.False(t, keepAlive)
    t.Logf("Keep-alive check for old login result: %v", keepAlive)

    // Test 5: Delete namespace
    t.Log("Starting Test 5: Delete namespace")
    err = manager.DeleteNamespace(ctx, tenant, nsName)
    assert.NoError(t, err)
    t.Logf("Namespace %s deletion initiated", nsName)

    // Test 6: Verify namespace is deleted
    t.Log("Starting Test 6: Verify namespace is deleted")
    err = client.Get(ctx, types.NamespacedName{Name: nsName}, ns)
    assert.Error(t, err)
    t.Logf("Namespace %s deletion verified", nsName)

    // Test 7: Enforce resources with keep-alive false
    t.Log("Starting Test 7: Enforce resources with keep-alive false")
    ok, err = manager.EnforceClusterResources(ctx, tenant, nsName, false)
    assert.NoError(t, err)
    assert.False(t, ok)
    assert.False(t, tenant.Status.PersonalNamespace.Created)
    assert.Equal(t, "", tenant.Status.PersonalNamespace.Name)
    t.Log("Resources enforcement with keep-alive false completed")
}