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

// Package sshctrl allows tracking SSH connections from the bastion host to target hosts.
package sshctrl

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ti-mo/conntrack"
	"github.com/ti-mo/netfilter"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	// ProtoTCP is the protocol number for TCP.
	ProtoTCP = 6
)

// SSHTracker tracks SSH connections and emits metrics. Use NewSSHTracker to create a new instance.
type SSHTracker struct {
	port    uint16
	log     logr.Logger
	metric  *prometheus.CounterVec
	running atomic.Bool
}

// IsRunning returns true if the tracker is currently running, false otherwise.
func (t *SSHTracker) IsRunning() bool {
	return t.running.Load()
}

// NewSSHTracker creates and initializes a new SSH tracker.
func NewSSHTracker(log logr.Logger, port uint16, metricName string) *SSHTracker {
	obj := &SSHTracker{
		port: port,
		log:  log,
		metric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: metricName,
				Help: "SSH connections detected from bastion to a target",
			},
			[]string{"destination_ip", "destination_port"},
		),
	}
	obj.running.Store(false)
	metrics.Registry.MustRegister(obj.metric)
	return obj
}

// HandleEvent processes an SSH connection event and updates corresponding metrics.
func (t *SSHTracker) HandleEvent(destIP, destPort string) {
	t.log.Info("Incoming connection", "destIP", destIP, "destPort", destPort)
	t.metric.WithLabelValues(destIP, destPort).Inc()
}

// Start begins tracking SSH connections to the specified port using conntrack netfilter events.
// It listens for new conntrack entries (NFNLGRP_CONNTRACK_NEW) and increments Prometheus
// metrics for each new TCP connection whose destination port matches port.
func (t *SSHTracker) Start(ctx context.Context) error {
	c, err := conntrack.Dial(nil)
	if err != nil {
		return fmt.Errorf("error dialing conntrack: %w", err)
	}
	defer c.Close()

	evChan := make(chan conntrack.Event, 1024)
	errChan, err := c.Listen(evChan, 4, []netfilter.NetlinkGroup{netfilter.GroupCTNew})
	if err != nil {
		return fmt.Errorf("error starting conntrack listener: %w", err)
	}

	t.log.Info("SSH tracker started, listening for new connections on port", "port", t.port)
	t.running.Store(true)
	defer t.running.Store(false)

	for {
		select {
		case ev := <-evChan:
			t.log.V(utils.LogDebugLevel).Info("Received conntrack event", "event", ev)
			if ev.Type != conntrack.EventNew || ev.Flow == nil {
				t.log.V(utils.LogDebugLevel).Info("Ignoring non-new or nil flow event", "event", ev)
				continue
			}
			to := ev.Flow.TupleOrig
			// Filter for TCP (protocol 6) connections to the target port.
			if to.Proto.Protocol != ProtoTCP || to.Proto.DestinationPort != t.port {
				t.log.V(utils.LogDebugLevel).Info("Ignoring non-TCP or non-target port connection", "protocol", to.Proto.Protocol, "destPort", to.Proto.DestinationPort)
				continue
			}
			t.HandleEvent(to.IP.DestinationAddress.String(), fmt.Sprintf("%d", to.Proto.DestinationPort))
		case err := <-errChan:
			if err != nil {
				return fmt.Errorf("conntrack listener error: %w", err)
			}
		case <-ctx.Done():
			t.log.Info("SSH tracker stopping due to context cancellation")
			return nil
		}
	}
}
