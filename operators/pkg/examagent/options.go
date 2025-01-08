// Copyright 2020-2025 Politecnico di Torino
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

package examagent

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"

	"k8s.io/klog/v2"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/restcfg"
)

type options struct {
	AllowedIPs       string
	Namespace        string
	BasePath         string
	ListenerAddr     string
	PrintRequestBody bool
	ipNets           []*net.IPNet
}

// Options object holds all the examagent parameters.
var Options options

// Initialize flags and associate each parameter to the given options object.
func (o *options) Init() {
	flag.StringVar(&o.ListenerAddr, "address", ":8888", "[address]:port of the landing server")
	flag.StringVar(&o.Namespace, "namespace", "", "Namespace in which Templates are stored and instances will be created")
	flag.StringVar(&o.AllowedIPs, "allowed-ips", "", "Comma separated list of CIDRs that are allowed to create new instances")
	flag.StringVar(&o.BasePath, "base-path", "/api", "Base path of the Exam Agent API")
	flag.BoolVar(&o.PrintRequestBody, "print-request-body", false, "Print the request body (WARNING: might be unstable)")

	restcfg.InitFlags(nil)

	klog.InitFlags(nil)
}

// Parse and normalize options.
func (o *options) Parse() error {
	flag.Parse()

	if o.Namespace == "" {
		return errors.New("missing argument: namespace")
	}

	if o.AllowedIPs == "" {
		klog.Infoln("No whitelist IPs have been specified: all IPs are allowed")
	} else {
		ips := strings.Split(o.AllowedIPs, ",")
		o.ipNets = make([]*net.IPNet, len(ips))
		for i, ip := range ips {
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				return fmt.Errorf("invalid CIDR: %w", err)
			}
			o.ipNets[i] = ipnet
		}
	}

	if o.BasePath == "" {
		return errors.New("missing argument: base-path")
	}

	return nil
}

// AllowedIPs contains a list of IPs that are allowed to create new instances.
type AllowedIPs string

// CheckAllowedIP checks if the given IP is allowed within the AllowedIPs.
func (o *options) CheckAllowedIP(rawIP string) error {
	if o.AllowedIPs == "" {
		return nil
	}

	addr := net.ParseIP(rawIP)
	if addr == nil {
		return fmt.Errorf("cannot parse IP (%v)", rawIP)
	}

	for _, ipnet := range o.ipNets {
		if ipnet.Contains(addr) {
			return nil
		}
	}

	return fmt.Errorf("IP %v not whitelisted", rawIP)
}
