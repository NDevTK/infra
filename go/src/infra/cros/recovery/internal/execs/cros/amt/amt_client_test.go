// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package amt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindReturnValue(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?><a:Envelope xmlns:a="http://www.w3.org/2003/05/soap-envelope" xmlns:b="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:c="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:d="http://schemas.xmlsoap.org/ws/2005/02/trust" xmlns:e="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:f="http://schemas.dmtf.org/wbem/wsman/1/cimbinding.xsd" xmlns:g="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_PowerManagementService" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><a:Header><b:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</b:To><b:RelatesTo>uuid:2ff63a59-c845-4477-b633-02878ed55fd0</b:RelatesTo><b:Action a:mustUnderstand="true">http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_PowerManagementService/RequestPowerStateChangeResponse</b:Action><b:MessageID>uuid:00000000-8086-8086-8086-000000000010</b:MessageID><c:ResourceURI>http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_PowerManagementService</c:ResourceURI></a:Header><a:Body><g:RequestPowerStateChange_OUTPUT><g:ReturnValue>0</g:ReturnValue></g:RequestPowerStateChange_OUTPUT></a:Body></a:Envelope>`
	retvalue, _ := findReturnValue(response)

	assert.Equal(t, 0, retvalue)
}

func TestFindPowerState(t *testing.T) {
	response := `<?xml version="1.0" encoding="UTF-8"?><a:Envelope xmlns:a="http://www.w3.org/2003/05/soap-envelope" xmlns:b="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:c="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:d="http://schemas.xmlsoap.org/ws/2005/02/trust" xmlns:e="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:f="http://schemas.dmtf.org/wbem/wsman/1/cimbinding.xsd" xmlns:g="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_AssociatedPowerManagementService" xmlns:h="http://schemas.dmtf.org/wbem/wscim/1/common" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><a:Header><b:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</b:To><b:RelatesTo>uuid:e12c6258-06ec-4225-9516-5730f9b0ccd2</b:RelatesTo><b:Action a:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/GetResponse</b:Action><b:MessageID>uuid:00000000-8086-8086-8086-000000000080</b:MessageID><c:ResourceURI>http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_AssociatedPowerManagementService</c:ResourceURI></a:Header><a:Body><g:CIM_AssociatedPowerManagementService><g:AvailableRequestedPowerStates>10</g:AvailableRequestedPowerStates><g:AvailableRequestedPowerStates>8</g:AvailableRequestedPowerStates><g:AvailableRequestedPowerStates>5</g:AvailableRequestedPowerStates><g:AvailableRequestedPowerStates>11</g:AvailableRequestedPowerStates><g:PowerState>2</g:PowerState><g:RequestedPowerState>2</g:RequestedPowerState><g:ServiceProvided><b:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</b:Address><b:ReferenceParameters><c:ResourceURI>http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_PowerManagementService</c:ResourceURI><c:SelectorSet><c:Selector Name="CreationClassName">CIM_PowerManagementService</c:Selector><c:Selector Name="Name">Intel(r) AMT Power Management Service</c:Selector><c:Selector Name="SystemCreationClassName">CIM_ComputerSystem</c:Selector><c:Selector Name="SystemName">Intel(r) AMT</c:Selector></c:SelectorSet></b:ReferenceParameters></g:ServiceProvided><g:UserOfService><b:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</b:Address><b:ReferenceParameters><c:ResourceURI>http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ComputerSystem</c:ResourceURI><c:SelectorSet><c:Selector Name="CreationClassName">CIM_ComputerSystem</c:Selector><c:Selector Name="Name">ManagedSystem</c:Selector></c:SelectorSet></b:ReferenceParameters></g:UserOfService></g:CIM_AssociatedPowerManagementService></a:Body></a:Envelope>`
	pstate, _ := findPowerState(response)

	assert.Equal(t, 2, pstate)
}
