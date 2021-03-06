// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"strings"

	"google.golang.org/api/compute/v1"

	"go.chromium.org/gae/service/datastore"

	"go.chromium.org/luci/gce/api/config/v1"
)

// VMsKind is a VMs entity's kind in the datastore.
const VMsKind = "VMs"

// VMs is a root entity representing a configured block of VMs.
// VM entities should be created for each VMs entity.
type VMs struct {
	// _extra is where unknown properties are put into memory.
	// Extra properties are not written to the datastore.
	_extra datastore.PropertyMap `gae:"-,extra"`
	// _kind is the entity's kind in the datastore.
	_kind string `gae:"$kind,VMs"`
	// ID is the unique identifier for this VMs block.
	ID string `gae:"$id"`
	// Config is the config.Block representation of this entity.
	Config config.Block `gae:"config"`
}

// VMKind is a VM entity's kind in the datastore.
const VMKind = "VM"

// VM is a root entity representing a configured VM.
// GCE instances should be created for each VM entity.
type VM struct {
	// _extra is where unknown properties are put into memory.
	// Extra properties are not written to the datastore.
	_extra datastore.PropertyMap `gae:"-,extra"`
	// _kind is the entity's kind in the datastore.
	_kind string `gae:"$kind,VM"`
	// ID is the unique identifier for this VM.
	ID string `gae:"$id"`
	// Attributes is the config.VM describing the GCE instance to create.
	Attributes config.VM `gae:"attributes"`
	// Deadline is the Unix time when the GCE instance should be deleted.
	// This time is in UTC and in seconds.
	Deadline int64 `gae:"deadline"`
	// Drained indicates whether or not this VM is drained.
	// A GCE instance should not be created for a drained VM.
	// Any existing GCE instance should be deleted regardless of deadline.
	Drained bool `gae:"drained"`
	// Hostname is the short hostname of the GCE instance to create.
	Hostname string `gae:"hostname"`
	// Index is this VM's number with respect to its VMs entity.
	Index int32 `gae:"index"`
	// Lifetime is the number of seconds the GCE instance should live for.
	Lifetime int64 `gae:"lifetime"`
	// Prefix is the prefix to use when naming the GCE instance.
	Prefix string `gae:"prefix"`
	// Swarming is hostname of the Swarming server the GCE instance connects to.
	Swarming string `gae:"swarming"`
	// URL is the URL of the created GCE instance.
	URL string `gae:"url"`
	// VMs is the ID of the VMs entity this VM was created from.
	VMs string `gae:"vms"`
}

// getDisks returns a []*compute.AttachedDisk representation of this VM's disks.
func (vm *VM) getDisks() []*compute.AttachedDisk {
	if len(vm.Attributes.GetDisk()) == 0 {
		return nil
	}
	disks := make([]*compute.AttachedDisk, len(vm.Attributes.Disk))
	for i, disk := range vm.Attributes.Disk {
		disks[i] = &compute.AttachedDisk{
			// AutoDelete deletes the disk when the instance is deleted.
			AutoDelete: true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb:  disk.Size,
				DiskType:    disk.Type,
				SourceImage: disk.Image,
			},
		}
	}
	// GCE requires the first disk to be the boot disk.
	disks[0].Boot = true
	return disks
}

// getMetadata returns a *compute.Metadata representation of this VM's metadata.
func (vm *VM) getMetadata() *compute.Metadata {
	if len(vm.Attributes.GetMetadata()) == 0 {
		return nil
	}
	meta := &compute.Metadata{
		Items: make([]*compute.MetadataItems, len(vm.Attributes.Metadata)),
	}
	for i, data := range vm.Attributes.Metadata {
		// Implicitly rejects FromFile, which is only supported in configs.
		spl := strings.SplitN(data.GetFromText(), ":", 2)
		// Per strings.SplitN semantics, len(spl) > 0 when splitting on a non-empty separator.
		// Therefore we can be sure the spl[0] exists (even if it's an empty string).
		key := spl[0]
		var val *string
		if len(spl) > 1 {
			val = &spl[1]
		}
		meta.Items[i] = &compute.MetadataItems{
			Key:   key,
			Value: val,
		}
	}
	return meta
}

// getNetworkInterfaces returns a []*compute.NetworkInterface representation of this VM's network interfaces.
func (vm *VM) getNetworkInterfaces() []*compute.NetworkInterface {
	if len(vm.Attributes.GetNetworkInterface()) == 0 {
		return nil
	}
	nics := make([]*compute.NetworkInterface, len(vm.Attributes.NetworkInterface))
	for i, nic := range vm.Attributes.NetworkInterface {
		nics[i] = &compute.NetworkInterface{
			Network: nic.Network,
		}
		if len(nic.GetAccessConfig()) > 0 {
			nics[i].AccessConfigs = make([]*compute.AccessConfig, len(nic.AccessConfig))
			for j, cfg := range nic.AccessConfig {
				nics[i].AccessConfigs[j] = &compute.AccessConfig{
					Type: cfg.Type.String(),
				}
			}
		}
	}
	return nics
}

// getServiceAccounts returns a []*compute.ServiceAccount representation of this VM's service accounts.
func (vm *VM) getServiceAccounts() []*compute.ServiceAccount {
	if len(vm.Attributes.GetServiceAccount()) == 0 {
		return nil
	}
	accts := make([]*compute.ServiceAccount, len(vm.Attributes.ServiceAccount))
	for i, sa := range vm.Attributes.ServiceAccount {
		accts[i] = &compute.ServiceAccount{
			Email: sa.Email,
		}
		if len(sa.GetScope()) > 0 {
			accts[i].Scopes = make([]string, len(sa.Scope))
			for j, s := range sa.Scope {
				accts[i].Scopes[j] = s
			}
		}
	}
	return accts
}

// GetInstance returns a *compute.Instance representation of this VM.
func (vm *VM) GetInstance() *compute.Instance {
	inst := &compute.Instance{
		Name:              vm.Hostname,
		Disks:             vm.getDisks(),
		MachineType:       vm.Attributes.GetMachineType(),
		Metadata:          vm.getMetadata(),
		NetworkInterfaces: vm.getNetworkInterfaces(),
		ServiceAccounts:   vm.getServiceAccounts(),
	}
	return inst
}
