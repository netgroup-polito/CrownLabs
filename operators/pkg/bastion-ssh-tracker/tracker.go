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

// Package bastion_ssh_tracker allows tracking SSH connections from the bastion host to target hosts.
package bastion_ssh_tracker

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// SSHTracker tracks SSH connections and emits metrics.
type SSHTracker struct {
	stopCh chan struct{}
	done   chan struct{}
}

// SSHConnection represents an SSH connection.
type SSHConnection struct {
	SourceIP   string
	SourcePort uint16
	DestIP     string
	DestPort   uint16
	StartTime  time.Time
}

// ConnectionEvent represents an event related to SSH connection.
type ConnectionEvent struct {
	Conn *SSHConnection
}

// processPacket is the handler called each time the BPF filter identifies a new
// TCP session (TCP packet with only the SYN flag set). It extracts SSH
// connection information from the packet layers and creates a ConnectionEvent
// that will be processed by handleEvent to update Prometheus metrics.
// This function acts as the bridge between raw network packet data and the
// metrics collection system.
func processPacket(packet gopacket.Packet, eventQueue chan ConnectionEvent) {
	// Get IP layer
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return
	}
	ip, _ := ipLayer.(*layers.IPv4)

	// Get TCP layer
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return
	}
	tcp, _ := tcpLayer.(*layers.TCP)

	srcIP := ip.SrcIP.String()
	dstIP := ip.DstIP.String()
	srcPort := uint16(tcp.SrcPort)
	dstPort := uint16(tcp.DstPort)

	newConn := &SSHConnection{
		SourceIP:   srcIP,
		SourcePort: srcPort,
		DestIP:     dstIP,
		DestPort:   dstPort,
		StartTime:  time.Now(),
	}

	// Send to event queue
	eventQueue <- ConnectionEvent{
		Conn: newConn,
	}
}

// handleEvent function is called by the event processing goroutine whenever a
// ConnectionEvent is received from the eventQueue channel. The eventQueue is
// populated by the processPacket function when new SSH connections are detected
// from network packet analysis.
// handleEvent processes an SSH connection event and updates corresponding metrics.
func handleEvent(event ConnectionEvent) {
	fmt.Print("New connection detected towards: ", event.Conn.DestIP, ":", event.Conn.DestPort, "\n")
	sshConnections.WithLabelValues(event.Conn.DestIP, strconv.Itoa(int(event.Conn.DestPort))).Inc()
}

// NewSSHTracker creates and initializes a new SSH tracker.
func NewSSHTracker() *SSHTracker {
	return &SSHTracker{
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// Start begins tracking SSH connections on the specified interface, port, and snaplen.
func (t *SSHTracker) Start(iface string, port, snaplen int) error {
	defer close(t.done)

	szFrame, szBlock, numBlocks, err := afpacketComputeSize(8, snaplen, os.Getpagesize())
	if err != nil {
		return fmt.Errorf("error computing afpacket size: %w", err)
	}

	timeout := time.Millisecond * 100

	afHandle, err := newAfpacketHandle(iface, szFrame, szBlock, numBlocks, false, timeout)
	if err != nil {
		return fmt.Errorf("error creating afpacket handle: %w", err)
	}
	defer afHandle.Close()

	// Filter for new outbound TCP packets on the specified port
	// Explicitly check for SYN packets to identify new connections
	// We want to track when a connection is established
	filter := fmt.Sprintf("tcp dst port %d and tcp[tcpflags] & tcp-syn != 0", port)
	if err := afHandle.SetBPFFilter(filter, snaplen); err != nil {
		return fmt.Errorf("error setting BPF filter: %w", err)
	}

	source := gopacket.ZeroCopyPacketDataSource(afHandle)

	eventQueue := make(chan ConnectionEvent, 100)

	var wg sync.WaitGroup
	stopWorkers := make(chan struct{})
	stopPackets := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event := <-eventQueue:
				handleEvent(event)
			case <-stopWorkers:
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopPackets:
				return
			default:
				data, _, err := source.ZeroCopyReadPacketData()
				if err != nil {
					continue
				}
				packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
				processPacket(packet, eventQueue)
			}
		}
	}()

	<-t.stopCh
	close(stopPackets)
	// Let the packet processing finish
	time.Sleep(2 * timeout)
	afHandle.Close()
	close(stopWorkers)
	wg.Wait()

	return nil
}

// Stop gracefully stops the SSH tracker.
func (t *SSHTracker) Stop() {
	close(t.stopCh)
	<-t.done
}
