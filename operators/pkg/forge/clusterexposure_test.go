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

package forge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Workload cluster API server exposure", func() {
	DescribeTable("KamajiAPIServerOIDCConfig arguments",
		func(opts forge.ClusterAPIServerOpts, expected []string) {
			cfg := forge.KamajiAPIServerOIDCConfig(&opts, "oidc-ca")
			if expected == nil {
				Expect(cfg).To(BeNil())
				return
			}
			Expect(cfg.ExtraArgs).To(Equal(expected))
			Expect(cfg.ExtraVolumes).To(BeEmpty())
			Expect(cfg.ExtraVolumeMounts).To(BeEmpty())
		},
		Entry("disabled without issuer", forge.ClusterAPIServerOpts{OIDCClientID: "k8s"}, nil),
		Entry("disabled without client ID", forge.ClusterAPIServerOpts{OIDCIssuerURL: "https://issuer"}, nil),
		Entry("without groups", forge.ClusterAPIServerOpts{OIDCIssuerURL: "https://issuer", OIDCClientID: "k8s"}, []string{
			"--oidc-issuer-url=https://issuer", "--oidc-client-id=k8s",
			"--oidc-username-claim=preferred_username", "--oidc-username-prefix=-",
		}),
		Entry("with groups", forge.ClusterAPIServerOpts{OIDCIssuerURL: "https://issuer", OIDCClientID: "k8s", OIDCGroupsClaim: "groups"}, []string{
			"--oidc-issuer-url=https://issuer", "--oidc-client-id=k8s",
			"--oidc-username-claim=preferred_username", "--oidc-username-prefix=-",
			"--oidc-groups-claim=groups", "--oidc-groups-prefix=kubernetes:",
		}),
	)

	It("mounts the custom OIDC CA in the API server", func() {
		opts := forge.ClusterAPIServerOpts{OIDCIssuerURL: "https://issuer", OIDCClientID: "k8s", OIDCCA: []byte("pem")}
		spec := forge.KamajiControlPlaneSpec(&clv1alpha2.ClusterFlavor{}, forge.KamajiAPIServerOIDCConfig(&opts, "oidc-ca"))
		Expect(spec.ApiServer.ExtraArgs).To(ContainElement("--oidc-ca-file=/etc/crownlabs/oidc/ca.crt"))
		Expect(spec.Deployment.ExtraVolumes).To(HaveLen(1))
		Expect(spec.Deployment.ExtraVolumes[0].ConfigMap.Name).To(Equal("oidc-ca"))
		Expect(spec.ApiServer.ExtraVolumeMounts).To(HaveLen(1))
		Expect(spec.ApiServer.ExtraVolumeMounts[0].Name).To(Equal(spec.Deployment.ExtraVolumes[0].Name))
		Expect(spec.ApiServer.ExtraVolumeMounts[0].MountPath).To(Equal("/etc/crownlabs/oidc"))
	})

	DescribeTable("ClusterAPIServerPath",
		func(prefix, expected string) {
			Expect(forge.ClusterAPIServerPath(prefix, "ns", "cluster")).To(Equal(expected))
		},
		Entry("default prefix", "", "/clusters/ns/cluster"),
		Entry("custom prefix with slashes", "/k8s/", "/k8s/ns/cluster"),
	)

	It("strips the path prefix before forwarding the requests to the API server", func() {
		expo := forge.ExpositionConfig{GatewayName: "gw", GatewayNamespace: "gw-ns"}
		spec := forge.ClusterAPIServerHTTPRouteSpec("/clusters/ns/cluster", "cp", &expo)
		Expect(spec.Rules).To(HaveLen(1))
		rule := spec.Rules[0]
		Expect(*rule.Matches[0].Path.Value).To(Equal("/clusters/ns/cluster"))
		Expect(rule.Filters).To(HaveLen(1))
		Expect(rule.Filters[0].Type).To(Equal(gatewayv1.HTTPRouteFilterURLRewrite))
		Expect(*rule.Filters[0].URLRewrite.Path.ReplacePrefixMatch).To(Equal("/"))
		Expect(string(rule.BackendRefs[0].Name)).To(Equal("cp"))
		Expect(*rule.BackendRefs[0].Port).To(BeEquivalentTo(6443))
	})
})
