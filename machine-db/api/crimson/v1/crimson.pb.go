// Code generated by protoc-gen-go. DO NOT EDIT.
// source: go.chromium.org/luci/machine-db/api/crimson/v1/crimson.proto

/*
Package crimson is a generated protocol buffer package.

It is generated from these files:
	go.chromium.org/luci/machine-db/api/crimson/v1/crimson.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/datacenters.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/hosts.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/machines.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/nics.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/oses.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/platforms.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/racks.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/switches.proto
	go.chromium.org/luci/machine-db/api/crimson/v1/vlans.proto

It has these top-level messages:
	ListDatacentersRequest
	Datacenter
	ListDatacentersResponse
	Host
	CreateHostRequest
	ListHostsRequest
	ListHostsResponse
	Machine
	CreateMachineRequest
	ListMachinesRequest
	ListMachinesResponse
	NIC
	CreateNICRequest
	ListNICsRequest
	ListNICsResponse
	ListOSesRequest
	OS
	ListOSesResponse
	ListPlatformsRequest
	Platform
	ListPlatformsResponse
	ListRacksRequest
	Rack
	ListRacksResponse
	ListSwitchesRequest
	Switch
	ListSwitchesResponse
	ListVLANsRequest
	VLAN
	ListVLANsResponse
*/
package crimson

import prpc "go.chromium.org/luci/grpc/prpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Crimson service

type CrimsonClient interface {
	// Lists datacenters in the database.
	ListDatacenters(ctx context.Context, in *ListDatacentersRequest, opts ...grpc.CallOption) (*ListDatacentersResponse, error)
	// Lists operating systems in the database.
	ListOSes(ctx context.Context, in *ListOSesRequest, opts ...grpc.CallOption) (*ListOSesResponse, error)
	// Lists platforms in the database.
	ListPlatforms(ctx context.Context, in *ListPlatformsRequest, opts ...grpc.CallOption) (*ListPlatformsResponse, error)
	// Lists racks in the database.
	ListRacks(ctx context.Context, in *ListRacksRequest, opts ...grpc.CallOption) (*ListRacksResponse, error)
	// Lists switches in the database.
	ListSwitches(ctx context.Context, in *ListSwitchesRequest, opts ...grpc.CallOption) (*ListSwitchesResponse, error)
	// Lists VLANs in the database.
	ListVLANs(ctx context.Context, in *ListVLANsRequest, opts ...grpc.CallOption) (*ListVLANsResponse, error)
	// Creates a new machine in the database.
	CreateMachine(ctx context.Context, in *CreateMachineRequest, opts ...grpc.CallOption) (*Machine, error)
	// Lists machines in the database.
	ListMachines(ctx context.Context, in *ListMachinesRequest, opts ...grpc.CallOption) (*ListMachinesResponse, error)
	// Creates a new NIC in the database.
	CreateNIC(ctx context.Context, in *CreateNICRequest, opts ...grpc.CallOption) (*NIC, error)
	// Lists NICs in the database.
	ListNICs(ctx context.Context, in *ListNICsRequest, opts ...grpc.CallOption) (*ListNICsResponse, error)
	// Creates a new host in the database.
	CreateHost(ctx context.Context, in *CreateHostRequest, opts ...grpc.CallOption) (*Host, error)
	// Lists hosts in the database.
	ListHosts(ctx context.Context, in *ListHostsRequest, opts ...grpc.CallOption) (*ListHostsResponse, error)
}
type crimsonPRPCClient struct {
	client *prpc.Client
}

func NewCrimsonPRPCClient(client *prpc.Client) CrimsonClient {
	return &crimsonPRPCClient{client}
}

func (c *crimsonPRPCClient) ListDatacenters(ctx context.Context, in *ListDatacentersRequest, opts ...grpc.CallOption) (*ListDatacentersResponse, error) {
	out := new(ListDatacentersResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListDatacenters", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListOSes(ctx context.Context, in *ListOSesRequest, opts ...grpc.CallOption) (*ListOSesResponse, error) {
	out := new(ListOSesResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListOSes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListPlatforms(ctx context.Context, in *ListPlatformsRequest, opts ...grpc.CallOption) (*ListPlatformsResponse, error) {
	out := new(ListPlatformsResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListPlatforms", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListRacks(ctx context.Context, in *ListRacksRequest, opts ...grpc.CallOption) (*ListRacksResponse, error) {
	out := new(ListRacksResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListRacks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListSwitches(ctx context.Context, in *ListSwitchesRequest, opts ...grpc.CallOption) (*ListSwitchesResponse, error) {
	out := new(ListSwitchesResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListSwitches", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListVLANs(ctx context.Context, in *ListVLANsRequest, opts ...grpc.CallOption) (*ListVLANsResponse, error) {
	out := new(ListVLANsResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListVLANs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) CreateMachine(ctx context.Context, in *CreateMachineRequest, opts ...grpc.CallOption) (*Machine, error) {
	out := new(Machine)
	err := c.client.Call(ctx, "crimson.Crimson", "CreateMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListMachines(ctx context.Context, in *ListMachinesRequest, opts ...grpc.CallOption) (*ListMachinesResponse, error) {
	out := new(ListMachinesResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListMachines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) CreateNIC(ctx context.Context, in *CreateNICRequest, opts ...grpc.CallOption) (*NIC, error) {
	out := new(NIC)
	err := c.client.Call(ctx, "crimson.Crimson", "CreateNIC", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListNICs(ctx context.Context, in *ListNICsRequest, opts ...grpc.CallOption) (*ListNICsResponse, error) {
	out := new(ListNICsResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListNICs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) CreateHost(ctx context.Context, in *CreateHostRequest, opts ...grpc.CallOption) (*Host, error) {
	out := new(Host)
	err := c.client.Call(ctx, "crimson.Crimson", "CreateHost", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonPRPCClient) ListHosts(ctx context.Context, in *ListHostsRequest, opts ...grpc.CallOption) (*ListHostsResponse, error) {
	out := new(ListHostsResponse)
	err := c.client.Call(ctx, "crimson.Crimson", "ListHosts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type crimsonClient struct {
	cc *grpc.ClientConn
}

func NewCrimsonClient(cc *grpc.ClientConn) CrimsonClient {
	return &crimsonClient{cc}
}

func (c *crimsonClient) ListDatacenters(ctx context.Context, in *ListDatacentersRequest, opts ...grpc.CallOption) (*ListDatacentersResponse, error) {
	out := new(ListDatacentersResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListDatacenters", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListOSes(ctx context.Context, in *ListOSesRequest, opts ...grpc.CallOption) (*ListOSesResponse, error) {
	out := new(ListOSesResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListOSes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListPlatforms(ctx context.Context, in *ListPlatformsRequest, opts ...grpc.CallOption) (*ListPlatformsResponse, error) {
	out := new(ListPlatformsResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListPlatforms", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListRacks(ctx context.Context, in *ListRacksRequest, opts ...grpc.CallOption) (*ListRacksResponse, error) {
	out := new(ListRacksResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListRacks", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListSwitches(ctx context.Context, in *ListSwitchesRequest, opts ...grpc.CallOption) (*ListSwitchesResponse, error) {
	out := new(ListSwitchesResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListSwitches", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListVLANs(ctx context.Context, in *ListVLANsRequest, opts ...grpc.CallOption) (*ListVLANsResponse, error) {
	out := new(ListVLANsResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListVLANs", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) CreateMachine(ctx context.Context, in *CreateMachineRequest, opts ...grpc.CallOption) (*Machine, error) {
	out := new(Machine)
	err := grpc.Invoke(ctx, "/crimson.Crimson/CreateMachine", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListMachines(ctx context.Context, in *ListMachinesRequest, opts ...grpc.CallOption) (*ListMachinesResponse, error) {
	out := new(ListMachinesResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListMachines", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) CreateNIC(ctx context.Context, in *CreateNICRequest, opts ...grpc.CallOption) (*NIC, error) {
	out := new(NIC)
	err := grpc.Invoke(ctx, "/crimson.Crimson/CreateNIC", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListNICs(ctx context.Context, in *ListNICsRequest, opts ...grpc.CallOption) (*ListNICsResponse, error) {
	out := new(ListNICsResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListNICs", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) CreateHost(ctx context.Context, in *CreateHostRequest, opts ...grpc.CallOption) (*Host, error) {
	out := new(Host)
	err := grpc.Invoke(ctx, "/crimson.Crimson/CreateHost", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crimsonClient) ListHosts(ctx context.Context, in *ListHostsRequest, opts ...grpc.CallOption) (*ListHostsResponse, error) {
	out := new(ListHostsResponse)
	err := grpc.Invoke(ctx, "/crimson.Crimson/ListHosts", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Crimson service

type CrimsonServer interface {
	// Lists datacenters in the database.
	ListDatacenters(context.Context, *ListDatacentersRequest) (*ListDatacentersResponse, error)
	// Lists operating systems in the database.
	ListOSes(context.Context, *ListOSesRequest) (*ListOSesResponse, error)
	// Lists platforms in the database.
	ListPlatforms(context.Context, *ListPlatformsRequest) (*ListPlatformsResponse, error)
	// Lists racks in the database.
	ListRacks(context.Context, *ListRacksRequest) (*ListRacksResponse, error)
	// Lists switches in the database.
	ListSwitches(context.Context, *ListSwitchesRequest) (*ListSwitchesResponse, error)
	// Lists VLANs in the database.
	ListVLANs(context.Context, *ListVLANsRequest) (*ListVLANsResponse, error)
	// Creates a new machine in the database.
	CreateMachine(context.Context, *CreateMachineRequest) (*Machine, error)
	// Lists machines in the database.
	ListMachines(context.Context, *ListMachinesRequest) (*ListMachinesResponse, error)
	// Creates a new NIC in the database.
	CreateNIC(context.Context, *CreateNICRequest) (*NIC, error)
	// Lists NICs in the database.
	ListNICs(context.Context, *ListNICsRequest) (*ListNICsResponse, error)
	// Creates a new host in the database.
	CreateHost(context.Context, *CreateHostRequest) (*Host, error)
	// Lists hosts in the database.
	ListHosts(context.Context, *ListHostsRequest) (*ListHostsResponse, error)
}

func RegisterCrimsonServer(s prpc.Registrar, srv CrimsonServer) {
	s.RegisterService(&_Crimson_serviceDesc, srv)
}

func _Crimson_ListDatacenters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDatacentersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListDatacenters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListDatacenters",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListDatacenters(ctx, req.(*ListDatacentersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListOSes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOSesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListOSes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListOSes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListOSes(ctx, req.(*ListOSesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListPlatforms_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPlatformsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListPlatforms(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListPlatforms",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListPlatforms(ctx, req.(*ListPlatformsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListRacks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRacksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListRacks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListRacks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListRacks(ctx, req.(*ListRacksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListSwitches_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSwitchesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListSwitches(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListSwitches",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListSwitches(ctx, req.(*ListSwitchesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListVLANs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListVLANsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListVLANs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListVLANs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListVLANs(ctx, req.(*ListVLANsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_CreateMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateMachineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).CreateMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/CreateMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).CreateMachine(ctx, req.(*CreateMachineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListMachines_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMachinesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListMachines(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListMachines",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListMachines(ctx, req.(*ListMachinesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_CreateNIC_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateNICRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).CreateNIC(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/CreateNIC",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).CreateNIC(ctx, req.(*CreateNICRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListNICs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListNICsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListNICs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListNICs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListNICs(ctx, req.(*ListNICsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_CreateHost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateHostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).CreateHost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/CreateHost",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).CreateHost(ctx, req.(*CreateHostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Crimson_ListHosts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListHostsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrimsonServer).ListHosts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crimson.Crimson/ListHosts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrimsonServer).ListHosts(ctx, req.(*ListHostsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Crimson_serviceDesc = grpc.ServiceDesc{
	ServiceName: "crimson.Crimson",
	HandlerType: (*CrimsonServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListDatacenters",
			Handler:    _Crimson_ListDatacenters_Handler,
		},
		{
			MethodName: "ListOSes",
			Handler:    _Crimson_ListOSes_Handler,
		},
		{
			MethodName: "ListPlatforms",
			Handler:    _Crimson_ListPlatforms_Handler,
		},
		{
			MethodName: "ListRacks",
			Handler:    _Crimson_ListRacks_Handler,
		},
		{
			MethodName: "ListSwitches",
			Handler:    _Crimson_ListSwitches_Handler,
		},
		{
			MethodName: "ListVLANs",
			Handler:    _Crimson_ListVLANs_Handler,
		},
		{
			MethodName: "CreateMachine",
			Handler:    _Crimson_CreateMachine_Handler,
		},
		{
			MethodName: "ListMachines",
			Handler:    _Crimson_ListMachines_Handler,
		},
		{
			MethodName: "CreateNIC",
			Handler:    _Crimson_CreateNIC_Handler,
		},
		{
			MethodName: "ListNICs",
			Handler:    _Crimson_ListNICs_Handler,
		},
		{
			MethodName: "CreateHost",
			Handler:    _Crimson_CreateHost_Handler,
		},
		{
			MethodName: "ListHosts",
			Handler:    _Crimson_ListHosts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "go.chromium.org/luci/machine-db/api/crimson/v1/crimson.proto",
}

func init() {
	proto.RegisterFile("go.chromium.org/luci/machine-db/api/crimson/v1/crimson.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 406 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xcb, 0x4e, 0xeb, 0x30,
	0x10, 0x86, 0x77, 0xed, 0xa9, 0xd5, 0xea, 0x20, 0xaf, 0x20, 0xa2, 0xc0, 0x0b, 0x90, 0x88, 0x8b,
	0x84, 0x40, 0x5c, 0x0a, 0x61, 0x41, 0x45, 0x09, 0xa8, 0x45, 0xec, 0x5d, 0xd7, 0x34, 0x16, 0x4d,
	0x1c, 0x6c, 0xb7, 0xbc, 0x22, 0x8f, 0x85, 0x9c, 0x8c, 0xd3, 0x26, 0x4a, 0x17, 0xc9, 0x2e, 0xfe,
	0xbf, 0x99, 0x7f, 0x3c, 0x9e, 0x09, 0xba, 0x9e, 0x0b, 0x97, 0x86, 0x52, 0x44, 0x7c, 0x19, 0xb9,
	0x42, 0xce, 0xbd, 0xc5, 0x92, 0x72, 0x2f, 0x22, 0x34, 0xe4, 0x31, 0x3b, 0x9e, 0x4d, 0x3d, 0x92,
	0x70, 0x8f, 0x4a, 0x1e, 0x29, 0x11, 0x7b, 0xab, 0x13, 0xfb, 0xe9, 0x26, 0x52, 0x68, 0x81, 0xdb,
	0x70, 0x74, 0x06, 0x35, 0x6d, 0x66, 0x44, 0x13, 0xca, 0x62, 0xcd, 0xa4, 0xca, 0xac, 0x9c, 0x9b,
	0x9a, 0x0e, 0x40, 0x6c, 0xfa, 0x65, 0xcd, 0xf4, 0x98, 0xd3, 0xa6, 0xa9, 0x42, 0xe5, 0x55, 0xaf,
	0x6a, 0xa6, 0x86, 0x42, 0x69, 0x9b, 0x7b, 0x5b, 0x33, 0x37, 0x59, 0x10, 0xfd, 0x29, 0x64, 0xd4,
	0xb4, 0xb6, 0x24, 0xf4, 0xab, 0xe9, 0x63, 0xab, 0x1f, 0xae, 0x69, 0xd8, 0xb8, 0xed, 0xd5, 0x82,
	0xc4, 0x90, 0x7b, 0xfa, 0xdb, 0x42, 0x6d, 0x3f, 0x43, 0xf8, 0x1d, 0xfd, 0x1f, 0x71, 0xa5, 0x1f,
	0xd7, 0xcb, 0x80, 0x0f, 0x5d, 0xbb, 0x61, 0x25, 0x32, 0x66, 0xdf, 0x4b, 0xa6, 0xb4, 0x73, 0xb4,
	0x3d, 0x40, 0x25, 0x22, 0x56, 0x0c, 0xdf, 0xa1, 0x7f, 0x06, 0xbd, 0x4e, 0x98, 0xc2, 0xbb, 0x85,
	0x68, 0x23, 0x59, 0x9f, 0xbd, 0x0a, 0x02, 0x06, 0x01, 0xea, 0x19, 0xed, 0xcd, 0x3e, 0x38, 0xee,
	0x17, 0x62, 0x73, 0xdd, 0x5a, 0x1d, 0x6c, 0xc3, 0xe0, 0xf7, 0x80, 0x3a, 0x06, 0x8c, 0xcd, 0x00,
	0x70, 0xb1, 0x6e, 0xaa, 0x59, 0x1f, 0xa7, 0x0a, 0x81, 0xc7, 0x33, 0xea, 0x1a, 0x71, 0x02, 0x83,
	0xc0, 0xfb, 0x85, 0x58, 0x2b, 0x5b, 0xa7, 0xfe, 0x16, 0x5a, 0xbc, 0xd0, 0xc7, 0xe8, 0x3e, 0x28,
	0x5f, 0x28, 0xd5, 0xaa, 0x2f, 0x04, 0x08, 0x3c, 0x06, 0xa8, 0xe7, 0x4b, 0x46, 0x34, 0x7b, 0xc9,
	0xc6, 0xbe, 0xf1, 0x48, 0x05, 0xdd, 0x7a, 0xed, 0xe4, 0xd8, 0x26, 0x40, 0x4b, 0x70, 0x2c, 0xb7,
	0x64, 0xe5, 0xea, 0x96, 0xd6, 0x14, 0xae, 0x73, 0x8e, 0x3a, 0x59, 0xd9, 0x60, 0xe8, 0x6f, 0xb4,
	0x94, 0x6b, 0xd6, 0xa6, 0x9b, 0x23, 0x13, 0x08, 0xab, 0x12, 0x0c, 0xfd, 0xf2, 0xaa, 0x18, 0xa9,
	0x7a, 0x55, 0x32, 0x02, 0x65, 0x2f, 0x10, 0xca, 0x4a, 0x3c, 0x09, 0xa5, 0xb1, 0x53, 0xaa, 0x6b,
	0x44, 0x6b, 0xd2, 0xcb, 0x59, 0x1a, 0x0a, 0x23, 0x30, 0xdf, 0xe5, 0x11, 0xa4, 0x5a, 0xf5, 0x08,
	0x00, 0x65, 0xc5, 0xa7, 0xad, 0xf4, 0x8f, 0x3a, 0xfb, 0x0b, 0x00, 0x00, 0xff, 0xff, 0xe7, 0x25,
	0x6f, 0xaa, 0xc4, 0x05, 0x00, 0x00,
}
