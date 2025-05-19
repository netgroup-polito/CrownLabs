func TestNamespaceReconciliation(t *testing.T) {
    ctx := context.Background()
    
    // Create test tenant
    tenant := &crownlabsv1alpha2.Tenant{
        ObjectMeta: metav1.ObjectMeta{
            Name: "test-tenant",
            Labels: map[string]string{
                "test-key": "test-value",
            },
        },
        Spec: crownlabsv1alpha2.TenantSpec{
            LastLogin: metav1.Time{Time: time.Now()},
        },
    }

    // Create namespace and verify
    err := k8sClient.Create(ctx, tenant)
    g.Expect(err).NotTo(gomega.HaveOccurred())

    // Wait for namespace creation
    nsName := fmt.Sprintf("tenant-%s", tenant.Name)
    g.Eventually(func() error {
        ns := &v1.Namespace{}
        return k8sClient.Get(ctx, types.NamespacedName{Name: nsName}, ns)
    }, timeout).Should(gomega.Succeed())
}