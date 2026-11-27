// Copyright 2020-2026 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package forge

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// ClusterAPIServerPort is the port exposed by the Service of the Kamaji control planes.
	ClusterAPIServerPort = 6443
	// ClusterAPIServerPortName is the name of the port exposed by the Service of the Kamaji control planes.
	ClusterAPIServerPortName = "kube-apiserver"
	// ClusterAPIServerDefaultPathPrefix is the default path prefix used to expose the workload cluster API servers.
	ClusterAPIServerDefaultPathPrefix = "/clusters"
	// ClusterAPIServerCAKey is the key hosting the CA certificate, both in the Kamaji secret and in the ConfigMap.
	ClusterAPIServerCAKey = "ca.crt"
	// ClusterKubeconfigKey is the key of the Cluster API admin kubeconfig secret hosting the kubeconfig.
	ClusterKubeconfigKey = "value"
	// ClusterAdminsClusterRoleBindingName is the name of the ClusterRoleBinding granting cluster-admin
	// permissions to the owners of the workload cluster.
	ClusterAdminsClusterRoleBindingName = "crownlabs-cluster-admins"
	// ClusterAdminClusterRoleName is the name of the ClusterRole bound to the owners of the workload cluster.
	ClusterAdminClusterRoleName = "cluster-admin"

	// ClusterOIDCUsernameClaim is the OIDC claim used as username by the workload cluster API servers.
	ClusterOIDCUsernameClaim = "preferred_username"
	// ClusterOIDCUsernamePrefix disables the username prefixing, so that usernames match the tenant names.
	ClusterOIDCUsernamePrefix = "-"
	// ClusterOIDCGroupsPrefix is the prefix of OIDC groups, matching the one adopted by the CrownLabs cluster.
	ClusterOIDCGroupsPrefix = "kubernetes:"
	// ClusterOIDCCAVolumeName is the name of the volume hosting the OIDC issuer CA in the API server pod.
	ClusterOIDCCAVolumeName = "crownlabs-oidc-ca"
	// ClusterOIDCCAMountPath is the path the OIDC issuer CA is mounted at in the API server container.
	ClusterOIDCCAMountPath = "/etc/crownlabs/oidc"
)

// ClusterAPIServerOpts contains the configuration to expose the API server of the workload clusters.
type ClusterAPIServerOpts struct {
	// OIDCIssuerURL is the URL of the OIDC issuer trusted by the workload cluster API servers.
	// The API server exposure is disabled if empty.
	OIDCIssuerURL string
	// OIDCClientID is the OIDC client ID the tokens must be issued for.
	OIDCClientID string
	// OIDCGroupsClaim is the OIDC claim containing the groups of the user.
	OIDCGroupsClaim string
	// PathPrefix is the prefix of the path the API servers are exposed at.
	PathPrefix string
	// OIDCCA is the PEM-encoded CA used to verify the OIDC issuer certificate. If empty, the system CAs are used.
	OIDCCA []byte
}

// KamajiAPIServerConfig groups the customizations applied to the kube-apiserver of the Kamaji control planes.
type KamajiAPIServerConfig struct {
	ExtraArgs         []string
	ExtraVolumes      []corev1.Volume
	ExtraVolumeMounts []corev1.VolumeMount
}

// Enabled returns whether the exposure of the workload cluster API servers is configured.
func (o *ClusterAPIServerOpts) Enabled() bool {
	return o.OIDCIssuerURL != "" && o.OIDCClientID != ""
}

// KamajiAPIServerOIDCConfig returns the kube-apiserver customizations to authenticate the CrownLabs users through
// OIDC, mounting the issuer CA (if configured) from the ConfigMap with the given name. It returns nil if disabled.
func KamajiAPIServerOIDCConfig(opts *ClusterAPIServerOpts, caConfigMapName string) *KamajiAPIServerConfig {
	if !opts.Enabled() {
		return nil
	}

	cfg := &KamajiAPIServerConfig{ExtraArgs: []string{
		"--oidc-issuer-url=" + opts.OIDCIssuerURL,
		"--oidc-client-id=" + opts.OIDCClientID,
		"--oidc-username-claim=" + ClusterOIDCUsernameClaim,
		"--oidc-username-prefix=" + ClusterOIDCUsernamePrefix,
	}}
	if opts.OIDCGroupsClaim != "" {
		cfg.ExtraArgs = append(cfg.ExtraArgs,
			"--oidc-groups-claim="+opts.OIDCGroupsClaim,
			"--oidc-groups-prefix="+ClusterOIDCGroupsPrefix)
	}
	if len(opts.OIDCCA) > 0 {
		cfg.ExtraArgs = append(cfg.ExtraArgs, fmt.Sprintf("--oidc-ca-file=%s/%s", ClusterOIDCCAMountPath, ClusterAPIServerCAKey))
		cfg.ExtraVolumes = []corev1.Volume{{
			Name: ClusterOIDCCAVolumeName,
			VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: caConfigMapName},
			}},
		}}
		cfg.ExtraVolumeMounts = []corev1.VolumeMount{{
			Name: ClusterOIDCCAVolumeName, MountPath: ClusterOIDCCAMountPath, ReadOnly: true,
		}}
	}
	return cfg
}

// ClusterOIDCCAConfigMapName returns the name of the ConfigMap hosting the OIDC issuer CA for the given workload cluster.
func ClusterOIDCCAConfigMapName(clusterName string) string {
	return clusterName + StringSeparator + "oidc-ca"
}

// ClusterAPIServerPath returns the path the API server of the given workload cluster is exposed at.
func ClusterAPIServerPath(pathPrefix, namespace, clusterName string) string {
	if pathPrefix == "" {
		pathPrefix = ClusterAPIServerDefaultPathPrefix
	}
	return fmt.Sprintf("/%s/%s/%s", strings.Trim(pathPrefix, "/"), namespace, clusterName)
}

// ClusterAPIServerURL returns the external URL of the API server exposed at the given host and path.
func ClusterAPIServerURL(host, path string) string {
	return fmt.Sprintf("https://%s%s", host, path)
}

// ClusterAPIServerServiceHostname returns the in-cluster hostname of the Service of the Kamaji control plane,
// which is included in the SANs of the API server certificate.
func ClusterAPIServerServiceHostname(serviceName, namespace string) string {
	return fmt.Sprintf("%s.%s.svc", serviceName, namespace)
}

// ClusterAPIServerInternalURL returns the in-cluster URL of the API server of the Kamaji control plane, given
// the address of its Service.
func ClusterAPIServerInternalURL(serviceAddress string) string {
	return "https://" + net.JoinHostPort(serviceAddress, strconv.Itoa(ClusterAPIServerPort))
}

// ClusterAPIServerExposureName returns the name shared by the HTTPRoute, BackendTLSPolicy and CA ConfigMap
// exposing the API server of the given workload cluster.
func ClusterAPIServerExposureName(clusterName string) string {
	return clusterName + StringSeparator + "api-server"
}

// ClusterAPIServerCASecretName returns the name of the secret generated by Kamaji, hosting the cluster CA.
func ClusterAPIServerCASecretName(controlPlaneName string) string {
	return controlPlaneName + StringSeparator + "ca"
}

// ClusterKubeconfigSecretName returns the name of the secret generated by Cluster API, hosting the admin kubeconfig.
func ClusterKubeconfigSecretName(clusterName string) string {
	return clusterName + StringSeparator + "kubeconfig"
}

// ClusterAPIServerHTTPRouteSpec forges the spec of the HTTPRoute exposing the API server of a workload
// cluster at the given path, stripping the path prefix before forwarding the requests.
func ClusterAPIServerHTTPRouteSpec(path, serviceName string, expo *ExpositionConfig) gatewayv1.HTTPRouteSpec {
	return gatewayv1.HTTPRouteSpec{
		CommonRouteSpec: gatewayv1.CommonRouteSpec{
			ParentRefs: []gatewayv1.ParentReference{ParentReference(expo.GatewayName, expo.GatewayNamespace)},
		},
		Rules: []gatewayv1.HTTPRouteRule{{
			Matches: []gatewayv1.HTTPRouteMatch{HTTPRouteMatch(path)},
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Path: &gatewayv1.HTTPPathModifier{
						Type:               gatewayv1.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: ptr.To("/"),
					},
				},
			}},
			BackendRefs: []gatewayv1.HTTPBackendRef{HTTPBackendRef(serviceName, ClusterAPIServerPort)},
		}},
	}
}

// ClusterAPIServerBackendTLSPolicySpec forges the spec of the BackendTLSPolicy instructing the Gateway to
// connect through TLS to the API server, validating its certificate against the CA in the given ConfigMap.
func ClusterAPIServerBackendTLSPolicySpec(serviceName, namespace, caConfigMapName string) gatewayv1.BackendTLSPolicySpec {
	return gatewayv1.BackendTLSPolicySpec{
		TargetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
			LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
				Group: "",
				Kind:  "Service",
				Name:  gatewayv1.ObjectName(serviceName),
			},
			SectionName: ptr.To(gatewayv1.SectionName(ClusterAPIServerPortName)),
		}},
		Validation: gatewayv1.BackendTLSPolicyValidation{
			CACertificateRefs: []gatewayv1.LocalObjectReference{{
				Group: "",
				Kind:  "ConfigMap",
				Name:  gatewayv1.ObjectName(caConfigMapName),
			}},
			Hostname: gatewayv1.PreciseHostname(ClusterAPIServerServiceHostname(serviceName, namespace)),
		},
	}
}

// ClusterAdminsSubjects returns the subjects granted cluster-admin permissions in a workload cluster,
// i.e., the tenant owning the instance and the managers of the workspace of the template.
func ClusterAdminsSubjects(tenantName, workspaceName string) []rbacv1.Subject {
	subjects := []rbacv1.Subject{{
		Kind:     rbacv1.UserKind,
		APIGroup: rbacv1.GroupName,
		Name:     tenantName,
	}}
	if workspaceName != "" {
		subjects = append(subjects, rbacv1.Subject{
			Kind:     rbacv1.GroupKind,
			APIGroup: rbacv1.GroupName,
			Name:     ClusterOIDCGroupsPrefix + WorkspaceRoleName(workspaceName, clv1alpha2.Manager),
		})
	}
	return subjects
}

// ClusterAdminsRoleRef returns the reference to the ClusterRole bound to the owners of a workload cluster.
func ClusterAdminsRoleRef() rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "ClusterRole",
		Name:     ClusterAdminClusterRoleName,
	}
}

// ClusterAdminsLabels returns the labels of the ClusterRoleBinding granting permissions in a workload cluster.
func ClusterAdminsLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[LabelManagedByKey] = labelManagedByInstanceValue
	return labels
}
