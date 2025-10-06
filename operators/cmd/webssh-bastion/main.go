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

// Package main contains the entrypoint for webSSH, a WebSocket SSH bridge for CrownLabs.
package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/stdr"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/webssh-bastion"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = crownlabsv1alpha1.AddToScheme(scheme)
	_ = crownlabsv1alpha2.AddToScheme(scheme)
}

func main() {
	sshUserFlag := flag.String("websshuser", "crownlabs", "The user to use for SSH connections.")
	websshprivatekeypathFlag := flag.String("websshprivatekeypath", "", "The path to the private key file for SSH authentication.")
	websshtimeoutdurationFlag := flag.String("websshtimeoutduration", "0", "The timeout duration for SSH connections. In minutes.")
	websshmaxconncountFlag := flag.String("websshmaxconncount", "1000", "The maximum number of concurrent SSH connections.")
	websshvmport := flag.String("websshvmport", "22", "The default SSH port for VMs.")
	websshwebsocketportFlag := flag.String("websshwebsocketport", "8085", "The port on which the WebSocket server listens.")

	flag.Parse()

	timeout64, err := strconv.ParseInt(*websshtimeoutdurationFlag, 10, 32) // parse directly as int32
	if err != nil {
		timeout64 = 30
	}

	maxConn64, err := strconv.ParseInt(*websshmaxconncountFlag, 10, 32)
	if err != nil {
		maxConn64 = 1000
	}

	stdLogger := log.New(os.Stderr, "", log.LstdFlags)
	baseLogger := stdr.New(stdLogger)

	webSSHCtx := &webssh.ServerContext{}
	webSSHCtx.BaseLogger = baseLogger
	webSSHCtx.SSHUser = *sshUserFlag
	webSSHCtx.PrivateKeyPath = *websshprivatekeypathFlag
	webSSHCtx.TimeoutDuration = time.Duration(timeout64) * time.Minute
	webSSHCtx.MaxConnectionCount = int32(maxConn64)
	webSSHCtx.WebsocketPort = *websshwebsocketportFlag
	webSSHCtx.VMSSHPort = *websshvmport
	webSSHCtx.BaseConfig, err = utils.GetRestConfig()

	if err != nil {
		webSSHCtx.BaseLogger.Error(err, "Failed to get REST config")
		return
	}

	info, err := os.Stat(webSSHCtx.PrivateKeyPath)
	if err != nil {
		webSSHCtx.BaseLogger.Error(err, "Cannot access private key file")
		return
	}

	if info.IsDir() {
		webSSHCtx.BaseLogger.Error(nil, "Private key path points to a directory, not a file")
		return
	}

	webSSHCtx.BaseLogger.Info("Config loaded",
		"SSHUser", webSSHCtx.SSHUser,
		"PrivateKeyPath", webSSHCtx.PrivateKeyPath,
		"TimeoutDuration", webSSHCtx.TimeoutDuration,
		"MaxConnectionCount", webSSHCtx.MaxConnectionCount,
		"WebsocketPort", webSSHCtx.WebsocketPort,
		"VMSSHPort", webSSHCtx.VMSSHPort)

	webSSHCtx.StartWebSSH()
}
