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

package utils

import (
	"fmt"
	"net"
	"strings"
)

// ParseIPPool parses a comma-separated list of IPs, CIDRs, or ranges into a slice of IP strings.
func ParseIPPool(env string) ([]string, error) {
	var ips []string
	seen := make(map[string]bool)
	for _, entry := range strings.Split(env, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// CIDR
		if strings.Contains(entry, "/") {
			ip, ipnet, err := net.ParseCIDR(entry)
			if err != nil {
				return nil, err
			}
			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
				ipStr := ip.String()
				if !seen[ipStr] {
					ips = append(ips, ipStr)
					seen[ipStr] = true
				}
			}
			continue
		}
		// Range
		if strings.Contains(entry, "-") {
			parts := strings.Split(entry, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid IP range: %s", entry)
			}
			start := net.ParseIP(strings.TrimSpace(parts[0]))
			end := net.ParseIP(strings.TrimSpace(parts[1]))
			if start == nil || end == nil {
				return nil, fmt.Errorf("invalid IP in range: %s", entry)
			}
			for ip := start; !ipAfter(ip, end); incIP(ip) {
				ipStr := ip.String()
				if !seen[ipStr] {
					ips = append(ips, ipStr)
					seen[ipStr] = true
				}
			}
			continue
		}
		// Single IP
		if net.ParseIP(entry) != nil {
			ipStr := entry
			if !seen[ipStr] {
				ips = append(ips, ipStr)
				seen[ipStr] = true
			}
			continue
		}
		return nil, fmt.Errorf("invalid IP entry: %s", entry)
	}
	return ips, nil
}

// incIP increments an IP address (IPv4 only).
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ipAfter returns true if a > b.
func ipAfter(a, b net.IP) bool {
	for i := range a {
		if a[i] > b[i] {
			return true
		}
		if a[i] < b[i] {
			return false
		}
	}
	return false
}
