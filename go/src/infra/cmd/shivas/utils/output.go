// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"

	ufspb "infra/unifiedfleet/api/v1/proto"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// Titles for printing table format list
var (
	SwitchTitle              = []string{"Switch Name", "CapacityPort", "Zone", "Rack", "State", "UpdateTime"}
	KvmTitle                 = []string{"KVM Name", "MAC Address", "ChromePlatform", "CapacityPort", "Zone", "Rack", "State", "UpdateTime"}
	KvmFullTitle             = []string{"KVM Name", "MAC Address", "ChromePlatform", "CapacityPort", "IP", "Vlan", "State", "Zone", "Rack", "UpdateTime"}
	RpmTitle                 = []string{"RPM Name", "MAC Address", "CapacityPort", "UpdateTime"}
	DracTitle                = []string{"Drac Name", "Display name", "MAC Address", "Switch", "Switch Port", "Password", "Zone", "Rack", "Machine", "UpdateTime"}
	DracFullTitle            = []string{"Drac Name", "MAC Address", "Switch", "Switch Port", "Attached Host", "IP", "Vlan", "Zone", "Rack", "Machine", "UpdateTime"}
	NicTitle                 = []string{"Nic Name", "MAC Address", "Switch", "Switch Port", "Zone", "Rack", "Machine", "UpdateTime"}
	BrowserMachineTitle      = []string{"Machine Name", "Serial Number", "Zone", "Rack", "KVM", "KVM Port", "ChromePlatform", "DeploymentTicket", "Description", "State", "Realm", "UpdateTime"}
	OSMachineTitle           = []string{"Machine Name", "Zone", "Rack", "Barcode", "UpdateTime"}
	MachinelseprototypeTitle = []string{"Machine Prototype Name", "Occupied Capacity", "PeripheralTypes", "VirtualTypes", "Tags", "UpdateTime"}
	RacklseprototypeTitle    = []string{"Rack Prototype Name", "PeripheralTypes", "Tags", "UpdateTime"}
	ChromePlatformTitle      = []string{"Platform Name", "Manufacturer", "Description", "UpdateTime"}
	VlanTitle                = []string{"Vlan Name", "CIDR Block", "IP Capacity", "DHCP range", "Description", "State", "Zones", "UpdateTime"}
	VMTitle                  = []string{"VM Name", "OS Version", "MAC Address", "Zone", "Host", "Vlan", "IP", "State", "DeploymentTicket", "Description", "UpdateTime"}
	RackTitle                = []string{"Rack Name", "Zone", "Capacity", "State", "Realm", "UpdateTime"}
	MachineLSETitle          = []string{"Host", "OS Version", "Zone", "Virtual Datacenter", "Rack", "Machine(s)", "Nic", "Vlan", "IP", "State", "VM capacity", "DeploymentTicket", "Description", "UpdateTime"}
	MachineLSETFullitle      = []string{"Host", "OS Version", "Manufacturer", "Machine", "Zone", "Virtual Datacenter", "Rack", "Nic", "IP", "Vlan", "MAC Address", "State", "VM capacity", "Description", "UpdateTime"}
	ZoneTitle                = []string{"Name", "EnumName", "Department"}
	StateTitle               = []string{"Name", "EnumName", "Description"}
)

// TimeFormat for all timestamps handled by shivas
var timeFormat = "2006-01-02_15:04:05_MST"

// The tab writer which defines the write format
var tw = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

// The io writer for json output
var bw = bufio.NewWriter(os.Stdout)

type listAll func(ctx context.Context, ic ufsAPI.FleetClient, pageSize int32, pageToken, filter string, keysOnly bool) ([]proto.Message, string, error)
type printJSONFunc func(res []proto.Message, emit bool)
type printFullFunc func(ctx context.Context, ic ufsAPI.FleetClient, res []proto.Message, tsv bool) error
type printNormalFunc func(res []proto.Message, tsv, keysOnly bool) error
type printAll func(context.Context, ufsAPI.FleetClient, bool, int32, string, string, bool, bool, bool) (string, error)
type getSingleFunc func(ctx context.Context, ic ufsAPI.FleetClient, name string) (proto.Message, error)

// PrintEntities a batch of entities based on user parameters
func PrintEntities(ctx context.Context, ic ufsAPI.FleetClient, res []proto.Message, printJSON printJSONFunc, printFull printFullFunc, printNormal printNormalFunc, json, emit, full, tsv, keysOnly bool) error {
	if json {
		printJSON(res, emit)
		return nil
	}
	if full {
		return printFull(ctx, ic, res, tsv)
	}
	printNormal(res, tsv, keysOnly)
	return nil
}

// BatchList returns the all listed entities by filters
func BatchList(ctx context.Context, ic ufsAPI.FleetClient, listFunc listAll, filters []string, pageSize int, keysOnly bool) ([]proto.Message, error) {
	errs := make(map[string]error)
	res := make([]proto.Message, 0)
	if len(filters) == 0 {
		// No filters, single DoList call
		protos, err := DoList(ctx, ic, listFunc, int32(pageSize), "", keysOnly)
		if err != nil {
			errs["emptyFilter"] = err
		}
		res = append(res, protos...)
		if pageSize > 0 && len(res) >= pageSize {
			res = res[0:pageSize]
		}
	} else if pageSize > 0 {
		// Filters with a pagesize limit
		// If user specifies a limit, calling DoList without concrrency avoids non-required list calls to UFS
		for _, filter := range filters {
			protos, err := DoList(ctx, ic, listFunc, int32(pageSize), filter, keysOnly)
			if err != nil {
				errs[filter] = err
			} else {
				res = append(res, protos...)
				if len(res) >= pageSize {
					res = res[0:pageSize]
					break
				}
			}
		}
	} else {
		// Filters without pagesize limit
		// If user doesnt specify any limit, call DoList for each filter concurrently to improve latency
		res, errs = concurrentList(ctx, ic, listFunc, filters, pageSize, keysOnly)
	}

	if len(errs) > 0 {
		fmt.Println("Fail to do some queries:")
		resErr := make([]error, 0, len(errs))
		for f, err := range errs {
			fmt.Printf("Filter %s: %s\n", f, err.Error())
			resErr = append(resErr, err)
		}
		return nil, errors.MultiError(resErr)
	}
	return res, nil
}

// concurrentList calls Dolist concurrently for each filter
func concurrentList(ctx context.Context, ic ufsAPI.FleetClient, listFunc listAll, filters []string, pageSize int, keysOnly bool) ([]proto.Message, map[string]error) {
	// buffered channel to append data to a slice in a thread safe env
	queue := make(chan []proto.Message, 1)
	// waitgroup for multiple goroutines
	var wg sync.WaitGroup
	// number of goroutines/threads in the wait group to run concurrently
	wg.Add(len(filters))
	// sync map to store the errors
	var merr sync.Map
	errs := make(map[string]error)
	res := make([]proto.Message, 0)
	for i := 0; i < len(filters); i++ {
		// goroutine for each filter
		go func(i int) {
			protos, err := DoList(ctx, ic, listFunc, int32(pageSize), filters[i], keysOnly)
			if err != nil {
				// store the err in sync map
				merr.Store(filters[i], err)
				// inform waitgroup that thread is completed
				wg.Done()
			} else {
				// send the protos to the buffered channel
				queue <- protos
			}
		}(i)
	}

	// goroutine to append data to slice
	go func() {
		// receive protos on queue channel
		for pm := range queue {
			// append proto messages to slice
			res = append(res, pm...)
			// inform waitgroup that one more goroutine/thread is completed.
			wg.Done()
		}
	}()

	// defer closing the channel
	defer close(queue)
	// wait for all goroutines in the waitgroup to complete
	wg.Wait()

	// iterate over sync map to copy data to a normal map for filter->errors
	merr.Range(func(key, value interface{}) bool {
		errs[fmt.Sprint(key)] = value.(error)
		return true
	})
	return res, errs
}

// DoList lists the outputs
func DoList(ctx context.Context, ic ufsAPI.FleetClient, listFunc listAll, pageSize int32, filter string, keysOnly bool) ([]proto.Message, error) {
	var pageToken string
	res := make([]proto.Message, 0)
	if pageSize == 0 {
		for {
			protos, token, err := listFunc(ctx, ic, ufsUtil.MaxPageSize, pageToken, filter, keysOnly)
			if err != nil {
				return nil, err
			}
			res = append(res, protos...)
			if token == "" {
				break
			}
			pageToken = token
		}
	} else {
		for i := int32(0); i < pageSize; i = i + ufsUtil.MaxPageSize {
			var size int32
			if pageSize-i < ufsUtil.MaxPageSize {
				size = pageSize % ufsUtil.MaxPageSize
			} else {
				size = ufsUtil.MaxPageSize
			}
			protos, token, err := listFunc(ctx, ic, size, pageToken, filter, keysOnly)
			if err != nil {
				return nil, err
			}
			res = append(res, protos...)
			if token == "" {
				break
			}
			pageToken = token
		}
	}
	return res, nil
}

// ConcurrentGet runs multiple goroutines making Get calls to UFS
func ConcurrentGet(ctx context.Context, ic ufsAPI.FleetClient, names []string, getSingle getSingleFunc) []proto.Message {
	var res []proto.Message
	// buffered channel to append data to a slice in a thread safe env
	queue := make(chan proto.Message, 1)
	// waitgroup for multiple goroutines
	var wg sync.WaitGroup
	// number of goroutines/threads in the wait group to run concurrently
	wg.Add(len(names))
	for i := 0; i < len(names); i++ {
		// goroutine for each id/name
		go func(i int) {
			// single Get request call to UFS
			m, err := getSingle(ctx, ic, names[i])
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error()+" => "+names[i])
				// inform waitgroup that thread is completed
				wg.Done()
			} else {
				// send the proto to the buffered channel
				queue <- m
			}
		}(i)
	}

	// goroutine to append data to slice
	go func() {
		// receive proto on queue channel
		for pm := range queue {
			// append proto message to slice
			res = append(res, pm)
			// inform waitgroup that one more goroutine/thread is completed.
			wg.Done()
		}
	}()

	// defer closing the channel
	defer close(queue)
	// wait for all goroutines in the waitgroup to complete
	wg.Wait()
	return res
}

// PrintListJSONFormat prints the list output in JSON format
func PrintListJSONFormat(ctx context.Context, ic ufsAPI.FleetClient, f printAll, json bool, pageSize int32, filter string, keysOnly, emit bool) error {
	var pageToken string
	fmt.Print("[")
	if pageSize == 0 {
		for {
			token, err := f(ctx, ic, json, ufsUtil.MaxPageSize, pageToken, filter, keysOnly, false, emit)
			if err != nil {
				return err
			}
			if token == "" {
				break
			}
			fmt.Print(",")
			pageToken = token
		}
	} else {
		for i := int32(0); i < pageSize; i = i + ufsUtil.MaxPageSize {
			var size int32
			if pageSize-i < ufsUtil.MaxPageSize {
				size = pageSize % ufsUtil.MaxPageSize
			} else {
				size = ufsUtil.MaxPageSize
			}
			token, err := f(ctx, ic, json, size, pageToken, filter, keysOnly, false, emit)
			if err != nil {
				return err
			}
			if token == "" {
				break
			} else if i+ufsUtil.MaxPageSize < pageSize {
				fmt.Print(",")
			}
			pageToken = token
		}
	}
	fmt.Println("]")
	return nil
}

// PrintTableTitle prints the table title with parameters
func PrintTableTitle(title []string, tsv, keysOnly bool) {
	if !tsv && !keysOnly {
		PrintTitle(title)
	}
}

// PrintListTableFormat prints list output in Table format
func PrintListTableFormat(ctx context.Context, ic ufsAPI.FleetClient, f printAll, json bool, pageSize int32, filter string, keysOnly bool, title []string, tsv bool) error {
	if !tsv {
		if keysOnly {
			PrintTitle(title[0:1])
		} else {
			PrintTitle(title)
		}
	}
	var pageToken string
	if pageSize == 0 {
		for {
			token, err := f(ctx, ic, json, ufsUtil.MaxPageSize, pageToken, filter, keysOnly, tsv, false)
			if err != nil {
				return err
			}
			if token == "" {
				break
			}
			pageToken = token
		}
	} else {
		for i := int32(0); i < pageSize; i = i + ufsUtil.MaxPageSize {
			var size int32
			if pageSize-i < ufsUtil.MaxPageSize {
				size = pageSize % ufsUtil.MaxPageSize
			} else {
				size = ufsUtil.MaxPageSize
			}
			token, err := f(ctx, ic, json, size, pageToken, filter, keysOnly, tsv, false)
			if err != nil {
				return err
			}
			if token == "" {
				break
			}
			pageToken = token
		}
	}
	return nil
}

// PrintJSON prints the interface output as json
func PrintJSON(t interface{}) error {
	switch reflect.TypeOf(t).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(t)
		fmt.Print("[")
		for i := 0; i < s.Len(); i++ {
			e, err := json.MarshalIndent(s.Index(i).Interface(), "", "\t")
			if err != nil {
				return err
			}
			fmt.Println(string(e))
			if i != s.Len()-1 {
				fmt.Println(",")
			}
		}
		fmt.Println("]")
	}
	return nil
}

// PrintProtoJSON prints the output as json
func PrintProtoJSON(pm proto.Message, emit bool) {
	defer bw.Flush()
	m := protojson.MarshalOptions{
		EmitUnpopulated: emit,
		Indent:          "\t",
	}
	json, err := m.Marshal(proto.MessageV2(pm))
	if err != nil {
		fmt.Println("Failed to marshal JSON")
	} else {
		bw.Write(json)
		fmt.Fprintf(bw, "\n")
	}
}

// PrintTitle prints the title fields in table form.
func PrintTitle(title []string) {
	for _, s := range title {
		fmt.Fprint(tw, fmt.Sprintf("%s\t", s))
	}
	fmt.Fprintln(tw)
}

// PrintSwitches prints the all switches in table form.
func PrintSwitches(res []proto.Message, keysOnly bool) {
	switches := make([]*ufspb.Switch, len(res))
	for i, r := range res {
		switches[i] = r.(*ufspb.Switch)
	}
	defer tw.Flush()
	for _, s := range switches {
		printSwitch(s, keysOnly)
	}
}

func switchOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.Switch)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		fmt.Sprintf("%d", m.GetCapacityPort()),
		m.GetZone(),
		m.GetRack(),
		m.GetResourceState().String(),
		ts,
	}
}

func printSwitch(sw *ufspb.Switch, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(sw.Name))
		return
	}
	var out string
	for _, s := range switchOutputStrs(sw) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintSwitchesJSON prints the switch details in json format.
func PrintSwitchesJSON(res []proto.Message, emit bool) {
	switches := make([]*ufspb.Switch, len(res))
	for i, r := range res {
		switches[i] = r.(*ufspb.Switch)
	}
	fmt.Print("[")
	for i, s := range switches {
		s.Name = ufsUtil.RemovePrefix(s.Name)
		PrintProtoJSON(s, emit)
		if i < len(switches)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func kvmFullOutputStrs(kvm *ufspb.KVM, dhcp *ufspb.DHCPConfig) []string {
	var ts string
	if t, err := ptypes.Timestamp(kvm.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(kvm.Name),
		kvm.GetMacAddress(),
		kvm.GetChromePlatform(),
		fmt.Sprintf("%d", kvm.GetCapacityPort()),
		dhcp.GetIp(),
		dhcp.GetVlan(),
		kvm.GetResourceState().String(),
		kvm.GetZone(),
		kvm.GetRack(),
		ts,
	}
}

// PrintKVMFull prints the full info for kvm
func PrintKVMFull(kvms []*ufspb.KVM, dhcps map[string]*ufspb.DHCPConfig) {
	defer tw.Flush()
	for i := range kvms {
		var out string
		for _, s := range kvmFullOutputStrs(kvms[i], dhcps[kvms[i].GetName()]) {
			out += fmt.Sprintf("%s\t", s)
		}
		fmt.Fprintln(tw, out)
	}
}

// PrintKVMs prints the all kvms in table form.
func PrintKVMs(res []proto.Message, keysOnly bool) {
	kvms := make([]*ufspb.KVM, len(res))
	for i, r := range res {
		kvms[i] = r.(*ufspb.KVM)
	}
	defer tw.Flush()
	for _, kvm := range kvms {
		printKVM(kvm, keysOnly)
	}
}

func kvmOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.KVM)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetMacAddress(),
		m.GetChromePlatform(),
		fmt.Sprintf("%d", m.GetCapacityPort()),
		m.GetZone(),
		m.GetRack(),
		m.GetResourceState().String(),
		ts,
	}
}

func printKVM(kvm *ufspb.KVM, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(kvm.Name))
		return
	}
	var out string
	for _, s := range kvmOutputStrs(kvm) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintKVMsJSON prints the kvm details in json format.
func PrintKVMsJSON(res []proto.Message, emit bool) {
	kvms := make([]*ufspb.KVM, len(res))
	for i, r := range res {
		kvms[i] = r.(*ufspb.KVM)
	}
	fmt.Print("[")
	for i, s := range kvms {
		s.Name = ufsUtil.RemovePrefix(s.Name)
		PrintProtoJSON(s, emit)
		if i < len(kvms)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintRPMs prints the all rpms in table form.
func PrintRPMs(res []proto.Message, keysOnly bool) {
	rpms := make([]*ufspb.RPM, len(res))
	for i, r := range res {
		rpms[i] = r.(*ufspb.RPM)
	}
	defer tw.Flush()
	for _, rpm := range rpms {
		printRPM(rpm, keysOnly)
	}
}

func rpmOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.RPM)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetMacAddress(),
		fmt.Sprintf("%d", m.GetCapacityPort()),
		m.GetZone(),
		m.GetRack(),
		m.GetResourceState().String(),
		ts,
	}
}

func printRPM(rpm *ufspb.RPM, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(rpm.Name))
		return
	}
	var out string
	for _, s := range rpmOutputStrs(rpm) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintRPMsJSON prints the rpm details in json format.
func PrintRPMsJSON(res []proto.Message, emit bool) {
	rpms := make([]*ufspb.RPM, len(res))
	for i, r := range res {
		rpms[i] = r.(*ufspb.RPM)
	}
	fmt.Print("[")
	for i, s := range rpms {
		s.Name = ufsUtil.RemovePrefix(s.Name)
		PrintProtoJSON(s, emit)
		if i < len(rpms)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func dracFullOutputStrs(m *ufspb.Drac, dhcp *ufspb.DHCPConfig) []string {
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.Name),
		m.GetMacAddress(),
		m.GetSwitchInterface().GetSwitch(),
		m.GetSwitchInterface().GetPortName(),
		dhcp.GetHostname(),
		dhcp.GetIp(),
		dhcp.GetVlan(),
		m.GetZone(),
		m.GetRack(),
		m.GetMachine(),
		ts,
	}
}

// PrintDracFull prints the full related msg for drac
func PrintDracFull(entities []*ufspb.Drac, dhcps map[string]*ufspb.DHCPConfig) {
	defer tw.Flush()
	for i := range entities {
		var out string
		for _, s := range dracFullOutputStrs(entities[i], dhcps[entities[i].GetName()]) {
			out += fmt.Sprintf("%s\t", s)
		}
		fmt.Fprintln(tw, out)
	}
}

// PrintDracs prints the all dracs in table form.
func PrintDracs(res []proto.Message, keysOnly bool) {
	dracs := make([]*ufspb.Drac, len(res))
	for i, r := range res {
		dracs[i] = r.(*ufspb.Drac)
	}
	defer tw.Flush()
	for _, drac := range dracs {
		printDrac(drac, keysOnly)
	}
}

func dracOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.Drac)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.Name),
		m.GetDisplayName(),
		m.GetMacAddress(),
		m.GetSwitchInterface().GetSwitch(),
		m.GetSwitchInterface().GetPortName(),
		m.GetPassword(),
		m.GetZone(),
		m.GetRack(),
		m.GetMachine(),
		ts,
	}
}

func printDrac(drac *ufspb.Drac, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(drac.Name))
		return
	}
	var out string
	for _, s := range dracOutputStrs(drac) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintDracsJSON prints the drac details in json format.
func PrintDracsJSON(res []proto.Message, emit bool) {
	dracs := make([]*ufspb.Drac, len(res))
	for i, r := range res {
		dracs[i] = r.(*ufspb.Drac)
	}
	fmt.Print("[")
	for i, s := range dracs {
		s.Name = ufsUtil.RemovePrefix(s.Name)
		PrintProtoJSON(s, emit)
		if i < len(dracs)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintNics prints the all nics in table form.
func PrintNics(res []proto.Message, keysOnly bool) {
	nics := make([]*ufspb.Nic, len(res))
	for i, r := range res {
		nics[i] = r.(*ufspb.Nic)
	}
	defer tw.Flush()
	for _, nic := range nics {
		printNic(nic, keysOnly)
	}
}

func nicOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.Nic)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetMacAddress(),
		m.GetSwitchInterface().GetSwitch(),
		m.GetSwitchInterface().GetPortName(),
		m.GetZone(),
		m.GetRack(),
		m.GetMachine(),
		ts,
	}
}

func printNic(nic *ufspb.Nic, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(nic.Name))
		return
	}
	var out string
	for _, s := range nicOutputStrs(nic) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintNicsJSON prints the nic details in json format.
func PrintNicsJSON(res []proto.Message, emit bool) {
	nics := make([]*ufspb.Nic, len(res))
	for i, r := range res {
		nics[i] = r.(*ufspb.Nic)
	}
	fmt.Print("[")
	for i, s := range nics {
		s.Name = ufsUtil.RemovePrefix(s.Name)
		PrintProtoJSON(s, emit)
		if i < len(nics)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func machineOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.Machine)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	if m.GetChromeBrowserMachine() != nil {
		return []string{
			ufsUtil.RemovePrefix(m.GetName()),
			m.GetSerialNumber(),
			m.GetLocation().GetZone().String(),
			m.GetLocation().GetRack(),
			m.GetChromeBrowserMachine().GetKvmInterface().GetKvm(),
			m.GetChromeBrowserMachine().GetKvmInterface().GetPortName(),
			m.GetChromeBrowserMachine().GetChromePlatform(),
			m.GetChromeBrowserMachine().GetDeploymentTicket(),
			m.GetChromeBrowserMachine().GetDescription(),
			m.GetResourceState().String(),
			m.GetRealm(),
			ts,
		}
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetLocation().GetZone().String(),
		m.GetLocation().GetRack(),
		m.GetLocation().GetBarcodeName(),
		ts,
	}
}

// PrintMachines prints the all machines in table form.
func PrintMachines(res []proto.Message, keysOnly bool) {
	machines := make([]*ufspb.Machine, len(res))
	for i, r := range res {
		machines[i] = r.(*ufspb.Machine)
	}
	defer tw.Flush()
	for _, m := range machines {
		printMachine(m, keysOnly)
	}
}

func printMachine(m *ufspb.Machine, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(m.Name))
		return
	}
	var out string
	for _, s := range machineOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintMachinesJSON prints the machine details in json format.
func PrintMachinesJSON(res []proto.Message, emit bool) {
	machines := make([]*ufspb.Machine, len(res))
	for i, r := range res {
		machines[i] = r.(*ufspb.Machine)
	}
	fmt.Print("[")
	for i, m := range machines {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(machines)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintMachineLSEPrototypes prints the all msleps in table form.
func PrintMachineLSEPrototypes(res []proto.Message, keysOnly bool) {
	entities := make([]*ufspb.MachineLSEPrototype, len(res))
	for i, r := range res {
		entities[i] = r.(*ufspb.MachineLSEPrototype)
	}
	defer tw.Flush()
	for _, m := range entities {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		printMachineLSEPrototype(m, keysOnly)
	}
}

func machineLSEPrototypeOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.MachineLSEPrototype)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	res := []string{
		ufsUtil.RemovePrefix(m.GetName()),
		fmt.Sprintf("%d", m.GetOccupiedCapacityRu()),
	}
	prs := m.GetPeripheralRequirements()
	var peripheralTypes string
	for _, pr := range prs {
		peripheralTypes += fmt.Sprintf("%s,", pr.GetPeripheralType())
	}
	res = append(res, strings.TrimSuffix(peripheralTypes, ","))
	var virtualTypes string
	for _, vm := range m.GetVirtualRequirements() {
		virtualTypes += fmt.Sprintf("%s,", vm.GetVirtualType())
	}
	res = append(res, strings.TrimSuffix(virtualTypes, ","))
	res = append(res, fmt.Sprintf("%s", m.GetTags()))
	res = append(res, ts)
	return res
}

func printMachineLSEPrototype(m *ufspb.MachineLSEPrototype, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(m.Name))
		return
	}
	var out string
	for _, s := range machineLSEPrototypeOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintMachineLSEPrototypesJSON prints the mslep details in json format.
func PrintMachineLSEPrototypesJSON(res []proto.Message, emit bool) {
	entities := make([]*ufspb.MachineLSEPrototype, len(res))
	for i, r := range res {
		entities[i] = r.(*ufspb.MachineLSEPrototype)
	}
	fmt.Print("[")
	for i, m := range entities {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(entities)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintRackLSEPrototypes prints the all msleps in table form.
func PrintRackLSEPrototypes(res []proto.Message, keysOnly bool) {
	rlseps := make([]*ufspb.RackLSEPrototype, len(res))
	for i, r := range res {
		rlseps[i] = r.(*ufspb.RackLSEPrototype)
	}
	defer tw.Flush()
	for _, m := range rlseps {
		printRackLSEPrototype(m, keysOnly)
	}
}

func rackLSEPrototypeOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.RackLSEPrototype)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	res := []string{ufsUtil.RemovePrefix(m.GetName())}
	var peripheralTypes string
	for _, pr := range m.GetPeripheralRequirements() {
		peripheralTypes += fmt.Sprintf("%s,", pr.GetPeripheralType())
	}
	res = append(res, strings.TrimSuffix(peripheralTypes, ","))
	res = append(res, fmt.Sprintf("%s", m.GetTags()))
	res = append(res, ts)
	return res
}

func printRackLSEPrototype(m *ufspb.RackLSEPrototype, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(m.Name))
		return
	}
	var out string
	for _, s := range rackLSEPrototypeOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintRackLSEPrototypesJSON prints the mslep details in json format.
func PrintRackLSEPrototypesJSON(res []proto.Message, emit bool) {
	rlseps := make([]*ufspb.RackLSEPrototype, len(res))
	for i, r := range res {
		rlseps[i] = r.(*ufspb.RackLSEPrototype)
	}
	fmt.Print("[")
	for i, m := range rlseps {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(rlseps)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintVlansJSON prints the vlan details in json format.
func PrintVlansJSON(res []proto.Message, emit bool) {
	vlans := make([]*ufspb.Vlan, len(res))
	for i, r := range res {
		vlans[i] = r.(*ufspb.Vlan)
	}
	fmt.Print("[")
	for i, m := range vlans {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(vlans)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintVlans prints the all vlans in table form.
func PrintVlans(res []proto.Message, keysOnly bool) {
	vlans := make([]*ufspb.Vlan, len(res))
	for i, r := range res {
		vlans[i] = r.(*ufspb.Vlan)
	}
	defer tw.Flush()
	for _, v := range vlans {
		printVlan(v, keysOnly)
	}
}

func vlanOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.Vlan)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	zones := make([]string, len(m.GetZones()))
	for i, z := range m.GetZones() {
		zones[i] = z.String()
	}
	var dhcpRange string
	if m.GetFreeStartIpv4Str() != "" {
		dhcpRange = fmt.Sprintf("%s-%s", m.GetFreeStartIpv4Str(), m.GetFreeEndIpv4Str())
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetVlanAddress(),
		fmt.Sprintf("%d", m.GetCapacityIp()),
		dhcpRange,
		m.GetDescription(),
		m.GetResourceState().String(),
		strSlicesToStr(zones),
		ts,
	}
}

func printVlan(m *ufspb.Vlan, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(m.Name))
		return
	}
	var out string
	for _, s := range vlanOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintChromePlatforms prints the all msleps in table form.
func PrintChromePlatforms(res []proto.Message, keysOnly bool) {
	platforms := make([]*ufspb.ChromePlatform, len(res))
	for i, r := range res {
		platforms[i] = r.(*ufspb.ChromePlatform)
	}
	defer tw.Flush()
	for _, m := range platforms {
		printChromePlatform(m, keysOnly)
	}
}

func platformOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.ChromePlatform)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetManufacturer(),
		m.GetDescription(),
		ts,
	}
}

func printChromePlatform(m *ufspb.ChromePlatform, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(m.Name))
		return
	}
	var out string
	for _, s := range platformOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintChromePlatformsJSON prints the mslep details in json format.
func PrintChromePlatformsJSON(res []proto.Message, emit bool) {
	platforms := make([]*ufspb.ChromePlatform, len(res))
	for i, r := range res {
		platforms[i] = r.(*ufspb.ChromePlatform)
	}
	fmt.Print("[")
	for i, m := range platforms {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(platforms)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintMachineLSEsJSON prints the machinelse details in json format.
func PrintMachineLSEsJSON(res []proto.Message, emit bool) {
	machinelses := make([]*ufspb.MachineLSE, len(res))
	for i, r := range res {
		machinelses[i] = r.(*ufspb.MachineLSE)
	}
	fmt.Print("[")
	for i, m := range machinelses {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(machinelses)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func machineLSEFullOutputStrs(lse *ufspb.MachineLSE, dhcp *ufspb.DHCPConfig) []string {
	var ts string
	if t, err := ptypes.Timestamp(lse.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(lse.GetName()),
		lse.GetChromeBrowserMachineLse().GetOsVersion().GetValue(),
		lse.GetManufacturer(),
		strSlicesToStr(lse.GetMachines()),
		lse.GetZone(),
		lse.GetChromeBrowserMachineLse().GetVirtualDatacenter(),
		lse.GetRack(),
		lse.GetNic(),
		dhcp.GetIp(),
		dhcp.GetVlan(),
		dhcp.GetMacAddress(),
		lse.GetResourceState().String(),
		fmt.Sprintf("%d", lse.GetChromeBrowserMachineLse().GetVmCapacity()),
		lse.GetDescription(),
		ts,
	}
}

// PrintMachineLSEFull prints the full info for a host
func PrintMachineLSEFull(entities []*ufspb.MachineLSE, dhcps map[string]*ufspb.DHCPConfig) {
	defer tw.Flush()
	for i := range entities {
		var out string
		for _, s := range machineLSEFullOutputStrs(entities[i], dhcps[entities[i].GetName()]) {
			out += fmt.Sprintf("%s\t", s)
		}
		fmt.Fprintln(tw, out)
	}
}

// PrintMachineLSEs prints the all machinelses in table form.
func PrintMachineLSEs(res []proto.Message, keysOnly bool) {
	entities := make([]*ufspb.MachineLSE, len(res))
	for i, r := range res {
		entities[i] = r.(*ufspb.MachineLSE)
	}
	defer tw.Flush()
	for _, m := range entities {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		printMachineLSE(m, keysOnly)
	}
}

func machineLSEOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.MachineLSE)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	machine := ""
	if len(m.GetMachines()) == 1 {
		machine = m.GetMachines()[0]
	}
	if len(m.GetMachines()) > 1 {
		machine = strSlicesToStr(m.GetMachines())
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetChromeBrowserMachineLse().GetOsVersion().GetValue(),
		m.GetZone(),
		m.GetChromeBrowserMachineLse().GetVirtualDatacenter(),
		m.GetRack(),
		machine,
		m.GetNic(),
		m.GetVlan(),
		m.GetIp(),
		m.GetResourceState().String(),
		fmt.Sprintf("%d", m.GetChromeBrowserMachineLse().GetVmCapacity()),
		m.GetDeploymentTicket(),
		m.GetDescription(),
		ts,
	}
}

func printMachineLSE(m *ufspb.MachineLSE, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(m.Name))
		return
	}
	var out string
	for _, s := range machineLSEOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintFreeVMs prints the all free slots in table form.
func PrintFreeVMs(entities []*ufspb.MachineLSE, dhcps map[string]*ufspb.DHCPConfig) {
	defer tw.Flush()
	PrintTitle([]string{"Host", "Os Version", "Manufacturer", "Vlan", "Zone", "Free slots", "State"})
	for _, h := range entities {
		h.Name = ufsUtil.RemovePrefix(h.Name)
		printFreeVM(h, dhcps[h.Name])
	}
}

func printFreeVM(host *ufspb.MachineLSE, dhcp *ufspb.DHCPConfig) {
	out := fmt.Sprintf("%s\t", host.GetName())
	out += fmt.Sprintf("%s\t", host.GetChromeBrowserMachineLse().GetOsVersion().GetValue())
	out += fmt.Sprintf("%s\t", host.GetManufacturer())
	out += fmt.Sprintf("%s\t", dhcp.GetVlan())
	out += fmt.Sprintf("%s\t", host.GetZone())
	out += fmt.Sprintf("%d\t", host.GetChromeBrowserMachineLse().GetVmCapacity())
	out += fmt.Sprintf("%s\t", host.GetResourceState().String())
	fmt.Fprintln(tw, out)
}

// PrintVMs prints the all vms in table form.
func PrintVMs(res []proto.Message, keysOnly bool) {
	vms := make([]*ufspb.VM, len(res))
	for i, r := range res {
		vms[i] = r.(*ufspb.VM)
	}
	defer tw.Flush()
	for _, vm := range vms {
		printVM(vm, keysOnly)
	}
}

func vmOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.VM)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetOsVersion().GetValue(),
		m.GetMacAddress(),
		m.GetZone(),
		m.GetMachineLseId(),
		m.GetVlan(),
		m.GetIp(),
		m.GetResourceState().String(),
		m.GetDeploymentTicket(),
		m.GetDescription(),
		ts,
	}
}

func printVM(vm *ufspb.VM, keysOnly bool) {
	if keysOnly {
		fmt.Fprintln(tw, ufsUtil.RemovePrefix(vm.Name))
		return
	}
	var out string
	for _, s := range vmOutputStrs(vm) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintVMsJSON prints the vm details in json format.
func PrintVMsJSON(res []proto.Message, emit bool) {
	vms := make([]*ufspb.VM, len(res))
	for i, r := range res {
		vms[i] = r.(*ufspb.VM)
	}
	fmt.Print("[")
	for i, m := range vms {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(vms)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// PrintRacks prints the all racks in table form.
func PrintRacks(res []proto.Message, keysOnly bool) {
	racks := make([]*ufspb.Rack, len(res))
	for i, r := range res {
		racks[i] = r.(*ufspb.Rack)
	}
	defer tw.Flush()
	for _, m := range racks {
		printRack(m, keysOnly)
	}
}

func rackOutputStrs(pm proto.Message) []string {
	m := pm.(*ufspb.Rack)
	var ts string
	if t, err := ptypes.Timestamp(m.GetUpdateTime()); err == nil {
		ts = t.Local().Format(timeFormat)
	}
	return []string{
		ufsUtil.RemovePrefix(m.GetName()),
		m.GetLocation().GetZone().String(),
		fmt.Sprintf("%d", m.GetCapacityRu()),
		m.GetResourceState().String(),
		m.GetRealm(),
		ts,
	}
}

func printRack(m *ufspb.Rack, keysOnly bool) {
	m.Name = ufsUtil.RemovePrefix(m.Name)
	if keysOnly {
		fmt.Fprintln(tw, m.GetName())
		return
	}
	var out string
	for _, s := range rackOutputStrs(m) {
		out += fmt.Sprintf("%s\t", s)
	}
	fmt.Fprintln(tw, out)
}

// PrintRacksJSON prints the rack details in json format.
func PrintRacksJSON(res []proto.Message, emit bool) {
	racks := make([]*ufspb.Rack, len(res))
	for i, r := range res {
		racks[i] = r.(*ufspb.Rack)
	}
	fmt.Print("[")
	for i, m := range racks {
		m.Name = ufsUtil.RemovePrefix(m.Name)
		PrintProtoJSON(m, emit)
		if i < len(racks)-1 {
			fmt.Print(",")
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func strSlicesToStr(slices []string) string {
	return strings.Join(slices, ",")
}

// PrintAllNormal prints a 2D slice with tabwriter
func PrintAllNormal(title []string, res [][]string, keysOnly bool) {
	defer tw.Flush()
	PrintTableTitle(title, false, keysOnly)
	for i := 0; i < len(res); i++ {
		var out string
		for _, s := range res[i] {
			out += fmt.Sprintf("%s\t", s)
		}
		fmt.Fprintln(tw, out)
	}
}
