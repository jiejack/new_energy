package iec61850

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestControlState_String(t *testing.T) {
	assert.Equal(t, "Idle", ControlStateIdle.String())
	assert.Equal(t, "Selected", ControlStateSelected.String())
	assert.Equal(t, "Operating", ControlStateOperating.String())
	assert.Equal(t, "Completed", ControlStateCompleted.String())
	assert.Equal(t, "Cancelled", ControlStateCancelled.String())
	assert.Equal(t, "Unknown", ControlState(99).String())
}

func TestControlModel_String(t *testing.T) {
	assert.Equal(t, "Direct with normal security", ControlModelDirectNormal.String())
	assert.Equal(t, "Select-before-operate with normal security", ControlModelSBONormal.String())
	assert.Equal(t, "Direct with enhanced security", ControlModelDirectEnhanced.String())
	assert.Equal(t, "Select-before-operate with enhanced security", ControlModelSBOEnhanced.String())
	assert.Equal(t, "Unknown", ControlModel(99).String())
}

func TestConnectionState_String(t *testing.T) {
	assert.Equal(t, "Disconnected", ConnectionStateDisconnected.String())
	assert.Equal(t, "Connecting", ConnectionStateConnecting.String())
	assert.Equal(t, "Connected", ConnectionStateConnected.String())
	assert.Equal(t, "Associating", ConnectionStateAssociating.String())
	assert.Equal(t, "Associated", ConnectionStateAssociated.String())
	assert.Equal(t, "Unknown", ConnectionState(99).String())
}

func TestControlObject_Cancel_InvalidState(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelSBONormal)
	err := ctlObj.Cancel()
	assert.Error(t, err)

	ctlObj.State = ControlStateCompleted
	err = ctlObj.Cancel()
	assert.Error(t, err)
}

func TestControlObject_Cancel_Selected(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelSBONormal)
	ctlObj.Select()
	err := ctlObj.Cancel()
	assert.NoError(t, err)
	assert.Equal(t, ControlStateCancelled, ctlObj.State)
}

func TestControlObject_IsSelectTimeout_NotSelected(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelSBONormal)
	assert.False(t, ctlObj.IsSelectTimeout(time.Second))
}

func TestControlObject_IsControlTimeout_NotOperating(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelSBONormal)
	assert.False(t, ctlObj.IsControlTimeout(time.Second))
}

func TestControlObject_IsSelectTimeout_Active(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelSBONormal)
	ctlObj.Select()
	assert.False(t, ctlObj.IsSelectTimeout(time.Hour))
}

func TestControlObject_IsControlTimeout_Active(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelDirectNormal)
	ctlObj.Operate(true)
	assert.False(t, ctlObj.IsControlTimeout(time.Hour))
}

func TestControlHandler_GetControlObject(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)

	obj := NewControlObject("ref1", ControlModelSBONormal)
	handler.AddControlObject(obj)

	retrieved, exists := handler.GetControlObject("ref1")
	assert.True(t, exists)
	assert.Equal(t, obj, retrieved)

	_, exists = handler.GetControlObject("nonexistent")
	assert.False(t, exists)
}

func TestControlHandler_Operate_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	err := handler.Operate("nonexistent", true)
	assert.Error(t, err)
}

func TestControlHandler_Cancel_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	err := handler.Cancel("nonexistent")
	assert.Error(t, err)
}

func TestControlHandler_SetTerminalConfirmation_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	err := handler.SetTerminalConfirmation("nonexistent", true)
	assert.Error(t, err)
}

func TestControlHandler_GetTerminalConfirmation_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	result := handler.GetTerminalConfirmation("nonexistent")
	assert.False(t, result)
}

func TestControlHandler_IsSelectTimeout_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	assert.False(t, handler.IsSelectTimeout("nonexistent"))
}

func TestControlHandler_IsControlTimeout_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	assert.False(t, handler.IsControlTimeout("nonexistent"))
}

func TestControlHandler_GetControlState_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewControlHandler(config)
	assert.Equal(t, ControlStateIdle, handler.GetControlState("nonexistent"))
}

func TestControlObject_DirectEnhanced(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelDirectEnhanced)
	err := ctlObj.Operate(true)
	assert.NoError(t, err)
	assert.Equal(t, ControlStateCompleted, ctlObj.State)
	assert.Equal(t, true, ctlObj.Value)
}

func TestControlObject_Operate_SBOEnhanced(t *testing.T) {
	ctlObj := NewControlObject("ref", ControlModelSBOEnhanced)
	ctlObj.Select()
	err := ctlObj.Operate(42)
	assert.NoError(t, err)
	assert.Equal(t, ControlStateCompleted, ctlObj.State)
	assert.Equal(t, 42, ctlObj.Value)
}

func TestEncodeControlRequest_FalseValue(t *testing.T) {
	req := &ControlRequest{
		Reference: "LD0/XCBR1.Pos",
		Value:     false,
		Test:      true,
		Check:     false,
	}
	data, err := EncodeControlRequest(req)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestDecodeControlResponse_TooShort(t *testing.T) {
	_, err := DecodeControlResponse([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestDecodeControlResponse_WithAddCause(t *testing.T) {
	ref := "LD0/XCBR1.Pos"
	refLen := uint16(len(ref))
	data := make([]byte, 0)
	data = append(data, byte(refLen>>8), byte(refLen))
	data = append(data, []byte(ref)...)
	data = append(data, byte(MMSResultSuccess))
	data = append(data, 5)
	data = append(data, []byte("cause")...)

	resp, err := DecodeControlResponse(data)
	require.NoError(t, err)
	assert.Equal(t, ref, resp.Reference)
	assert.Equal(t, MMSResultSuccess, resp.Result)
	assert.Equal(t, "cause", resp.AddCause)
}

func TestDecodeControlResponse_ShortAddCause(t *testing.T) {
	ref := "AB"
	refLen := uint16(len(ref))
	data := make([]byte, 0)
	data = append(data, byte(refLen>>8), byte(refLen))
	data = append(data, []byte(ref)...)
	data = append(data, byte(MMSResultAccessDenied))
	data = append(data, byte(10))

	resp, err := DecodeControlResponse(data)
	require.NoError(t, err)
	assert.Equal(t, ref, resp.Reference)
	assert.Equal(t, "", resp.AddCause)
}

func TestMMSClient_Connect_WithServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		header := make([]byte, 4)
		for {
			_, err := conn.Read(header)
			if err != nil {
				return
			}
			length := binary.BigEndian.Uint32(header)
			data := make([]byte, length)
			_, err = conn.Read(data)
			if err != nil {
				return
			}

			domain := "LD0"
			item := "MMXU1.TotW"
			respData := make([]byte, 4+len(domain)+4+len(item))
			binary.BigEndian.PutUint32(respData[0:4], uint32(len(domain)))
			copy(respData[4:], domain)
			binary.BigEndian.PutUint32(respData[4+len(domain):], uint32(len(item)))
			copy(respData[4+len(domain)+4:], item)
			respHeader := make([]byte, 4)
			binary.BigEndian.PutUint32(respHeader, uint32(len(respData)))
			conn.Write(respHeader)
			conn.Write(respData)
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	config := DefaultConfig()
	config.IPAddress = addr.IP.String()
	config.Port = addr.Port
	config.ConnectTimeout = 5 * time.Second
	config.ReadTimeout = 5 * time.Second
	config.WriteTimeout = 5 * time.Second

	client := NewMMSClient(config)
	err = client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	assert.Equal(t, ConnectionStateConnected, client.GetState())

	err = client.Associate()
	require.NoError(t, err)
	assert.Equal(t, ConnectionStateAssociated, client.GetState())

	val, err := client.ReadVariable("LD0/MMXU1.TotW.mag")
	assert.NoError(t, err)
	assert.Nil(t, val)

	err = client.WriteVariable("LD0/MMXU1.TotW.mag", 123.45)
	assert.NoError(t, err)

	results, err := client.ReadVariables([]string{"LD0/MMXU1.TotW.mag"})
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestMMSClient_Associate_NotConnected(t *testing.T) {
	config := DefaultConfig()
	client := NewMMSClient(config)
	err := client.Associate()
	assert.Equal(t, ErrNotConnected, err)
}

func TestDecodeMMSInitiateResponse_TooShort(t *testing.T) {
	_, err := DecodeMMSInitiateResponse([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestDecodeMMSReadResponse_TooShort(t *testing.T) {
	_, err := DecodeMMSReadResponse([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestDecodeMMSReadResponse_DomainTooShort(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 100)
	_, err := DecodeMMSReadResponse(data)
	assert.Error(t, err)
}

func TestDecodeMMSReadResponse_ItemTooShort(t *testing.T) {
	domain := "LD0"
	data := make([]byte, 4+len(domain)+4)
	binary.BigEndian.PutUint32(data[0:4], uint32(len(domain)))
	copy(data[4:], domain)
	binary.BigEndian.PutUint32(data[4+len(domain):], 100)
	_, err := DecodeMMSReadResponse(data)
	assert.Error(t, err)
}

func TestDecodeMMSWriteResponse_TooShort(t *testing.T) {
	_, err := DecodeMMSWriteResponse([]byte{0x01})
	assert.Error(t, err)
}

func TestDecodeMMSWriteResponse_Valid(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data[0:4], uint32(MMSResultSuccess))
	resp, err := DecodeMMSWriteResponse(data)
	require.NoError(t, err)
	assert.Equal(t, MMSResultSuccess, resp.Result)
}

func TestEncodeMMSGetDirectoryRequest(t *testing.T) {
	req := &MMSGetDirectoryRequest{ObjectClass: "Domain"}
	data, err := EncodeMMSGetDirectoryRequest(req)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestDecodeMMSGetDirectoryResponse_TooShort(t *testing.T) {
	_, err := DecodeMMSGetDirectoryResponse([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestDecodeMMSGetDirectoryResponse_Valid(t *testing.T) {
	domain1 := "LD0"
	domain2 := "LD1"
	data := make([]byte, 4+4+len(domain1)+4+len(domain2))
	offset := 0
	binary.BigEndian.PutUint32(data[offset:offset+4], 2)
	offset += 4
	binary.BigEndian.PutUint32(data[offset:offset+4], uint32(len(domain1)))
	offset += 4
	copy(data[offset:], domain1)
	offset += len(domain1)
	binary.BigEndian.PutUint32(data[offset:offset+4], uint32(len(domain2)))
	offset += 4
	copy(data[offset:], domain2)

	resp, err := DecodeMMSGetDirectoryResponse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resp.Domains))
	assert.Equal(t, "LD0", resp.Domains[0])
	assert.Equal(t, "LD1", resp.Domains[1])
}

func TestDecodeMMSGetDirectoryResponse_DomainLenTooShort(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data[0:4], 1)
	_, err := DecodeMMSGetDirectoryResponse(data)
	assert.Error(t, err)
}

func TestDecodeMMSGetDirectoryResponse_DomainTooShort(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 1)
	binary.BigEndian.PutUint32(data[4:8], 100)
	_, err := DecodeMMSGetDirectoryResponse(data)
	assert.Error(t, err)
}

func TestEncodeMMSGetVariableDirectoryRequest(t *testing.T) {
	req := &MMSGetVariableDirectoryRequest{DomainID: "LD0"}
	data, err := EncodeMMSGetVariableDirectoryRequest(req)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestDecodeMMSGetVariableDirectoryResponse_TooShort(t *testing.T) {
	_, err := DecodeMMSGetVariableDirectoryResponse([]byte{0x01})
	assert.Error(t, err)
}

func TestDecodeMMSGetVariableDirectoryResponse_Valid(t *testing.T) {
	var1 := "TotW"
	var2 := "TotV"
	data := make([]byte, 4+4+len(var1)+4+len(var2))
	offset := 0
	binary.BigEndian.PutUint32(data[offset:offset+4], 2)
	offset += 4
	binary.BigEndian.PutUint32(data[offset:offset+4], uint32(len(var1)))
	offset += 4
	copy(data[offset:], var1)
	offset += len(var1)
	binary.BigEndian.PutUint32(data[offset:offset+4], uint32(len(var2)))
	offset += 4
	copy(data[offset:], var2)

	resp, err := DecodeMMSGetVariableDirectoryResponse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resp.Variables))
	assert.Equal(t, "TotW", resp.Variables[0])
	assert.Equal(t, "TotV", resp.Variables[1])
}

func TestDecodeMMSGetVariableDirectoryResponse_VarLenTooShort(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data[0:4], 1)
	_, err := DecodeMMSGetVariableDirectoryResponse(data)
	assert.Error(t, err)
}

func TestDecodeMMSGetVariableDirectoryResponse_VarTooShort(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 1)
	binary.BigEndian.PutUint32(data[4:8], 100)
	_, err := DecodeMMSGetVariableDirectoryResponse(data)
	assert.Error(t, err)
}

func TestLogicalNode_GetDataObject(t *testing.T) {
	ln := NewLogicalNode("MMXU", "1")
	do := NewDataObject("TotW", "MX")
	ln.AddDataObject(do)

	retrieved, exists := ln.GetDataObject("TotW")
	assert.True(t, exists)
	assert.Equal(t, do, retrieved)

	_, exists = ln.GetDataObject("nonexistent")
	assert.False(t, exists)
}

func TestDataObject_GetDataAttribute(t *testing.T) {
	do := NewDataObject("TotW", "MX")
	da := NewDataAttribute("mag", "Float32")
	do.AddDataAttribute(da)

	retrieved, exists := do.GetDataAttribute("mag")
	assert.True(t, exists)
	assert.Equal(t, da, retrieved)

	_, exists = do.GetDataAttribute("nonexistent")
	assert.False(t, exists)
}

func TestReportHandler_AddURCB(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)

	urcb := NewURCB("urcb01", "LD0")
	err := handler.AddURCB(urcb)
	require.NoError(t, err)
}

func TestReportHandler_EnableReport_URCB(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)

	urcb := NewURCB("urcb01", "LD0")
	handler.AddURCB(urcb)

	err := handler.EnableReport("urcb01")
	require.NoError(t, err)
	assert.True(t, handler.IsReportEnabled("urcb01"))
}

func TestReportHandler_DisableReport_URCB(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)

	urcb := NewURCB("urcb01", "LD0")
	handler.AddURCB(urcb)
	handler.EnableReport("urcb01")

	err := handler.DisableReport("urcb01")
	require.NoError(t, err)
	assert.False(t, handler.IsReportEnabled("urcb01"))
}

func TestReportHandler_EnableReport_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)
	err := handler.EnableReport("nonexistent")
	assert.Error(t, err)
}

func TestReportHandler_DisableReport_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)
	err := handler.DisableReport("nonexistent")
	assert.Error(t, err)
}

func TestReportHandler_IsReportEnabled_NotFound(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)
	assert.False(t, handler.IsReportEnabled("nonexistent"))
}

func TestBRCB_EnableDisable(t *testing.T) {
	brcb := NewBRCB("brcb01", "LD0")
	assert.False(t, brcb.IsEnabled())
	brcb.Enable()
	assert.True(t, brcb.IsEnabled())
	brcb.Disable()
	assert.False(t, brcb.IsEnabled())
}

func TestBRCB_AddReport_WhenEnabled(t *testing.T) {
	brcb := NewBRCB("brcb01", "LD0")
	brcb.Enable()
	report := &Report{RptID: "brcb01", SqNum: 1}
	err := brcb.AddReport(report)
	require.NoError(t, err)
}

func TestURCB_AddReport_WhenEnabled(t *testing.T) {
	urcb := NewURCB("urcb01", "LD0")
	urcb.RptEna = true
	report := &Report{RptID: "urcb01", SqNum: 1}
	err := urcb.AddReport(report)
	require.NoError(t, err)
}

func TestSVBuffer_Get_OutOfRange(t *testing.T) {
	buf := NewSVBuffer(10)
	assert.Nil(t, buf.Get(-1))
	assert.Nil(t, buf.Get(0))
	assert.Nil(t, buf.Get(100))
}

func TestSVFrame_EncodeDecode_Roundtrip(t *testing.T) {
	frame := &SVFrame{
		AppID:     0x4000,
		Length:    100,
		Reserved1: 0,
		Reserved2: 0,
		SmpCnt:    5678,
		ConfRev:   2,
		SmpMod:    1,
		SmpRate:   4000,
		Dataset:   "dsMeas",
		ASDUCount: 1,
		ASDUs: []SVASDU{
			{
				SV: []SVValue{
					{Inst: 1, Value: 220.5, Quality: QualityGood},
					{Inst: 2, Value: 50.0, Quality: QualityGood},
				},
			},
		},
	}

	data, err := EncodeSVFrame(frame)
	require.NoError(t, err)

	decoded, err := DecodeSVFrame(data)
	require.NoError(t, err)
	assert.Equal(t, frame.AppID, decoded.AppID)
	assert.Equal(t, frame.SmpCnt, decoded.SmpCnt)
	assert.Equal(t, frame.ConfRev, decoded.ConfRev)
	assert.Equal(t, frame.Dataset, decoded.Dataset)
}

func TestDecodeSVFrame_DatasetTooShort(t *testing.T) {
	data := make([]byte, 22)
	data[18] = 10
	frame, err := DecodeSVFrame(data)
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.Equal(t, "", frame.Dataset)
	}
}

func TestSampledValuesHandler_ProcessFrame_CallbackError(t *testing.T) {
	config := DefaultConfig()
	handler := NewSampledValuesHandler(config)

	callback := func(data *SVData) error {
		return fmt.Errorf("callback error")
	}
	handler.Subscribe("SV01", callback)

	frame := &SVFrame{
		SmpCnt:    1,
		ASDUCount: 1,
		ASDUs: []SVASDU{
			{SV: []SVValue{{Inst: 1, Value: 100.0, Quality: QualityGood}}},
		},
	}

	err := handler.ProcessFrame(frame)
	assert.NoError(t, err)
}

func TestMMSClient_Connect_IPv6(t *testing.T) {
	config := DefaultConfig()
	config.IPAddress = "::1"
	config.Port = 19999
	config.ConnectTimeout = 1 * time.Second

	client := NewMMSClient(config)
	err := client.Connect()
	assert.Error(t, err)
}

func TestMMSClient_Disconnect_WithConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	config := DefaultConfig()
	config.IPAddress = addr.IP.String()
	config.Port = addr.Port
	config.ConnectTimeout = 5 * time.Second

	client := NewMMSClient(config)
	err = client.Connect()
	require.NoError(t, err)

	err = client.Disconnect()
	assert.NoError(t, err)
	assert.Equal(t, ConnectionStateDisconnected, client.GetState())
}

func TestDecodeReport_InvalidData(t *testing.T) {
	_, err := DecodeReport([]byte{})
	assert.Error(t, err)
}

func TestEncodeReport_EmptyValues(t *testing.T) {
	report := &Report{
		RptID:   "brcb01",
		SqNum:   1,
		Time:    NewTimestamp(time.Now()),
		ConfRev: 1,
		Values:  []ReportValue{},
	}
	data, err := EncodeReport(report)
	require.NoError(t, err)
	assert.NotNil(t, data)
}

func TestReportBuffer_Get_OutOfRange(t *testing.T) {
	buf := NewReportBuffer(10)
	assert.Nil(t, buf.Get(-1))
	assert.Nil(t, buf.Get(100))
}

func TestReportBuffer_Overflow(t *testing.T) {
	buf := NewReportBuffer(2)
	for i := 0; i < 5; i++ {
		report := &Report{RptID: "test", SqNum: uint16(i)}
		buf.Add(report)
	}
	assert.Equal(t, 2, buf.Size())
}

func TestReportHandler_ProcessDataChange_DisabledReport(t *testing.T) {
	config := DefaultConfig()
	handler := NewReportHandler(config)

	brcb := NewBRCB("brcb01", "LD0")
	brcb.SetTriggerOptions(TriggerDataChange)
	handler.AddBRCB(brcb)

	change := &DataChange{
		Ref:     "LD0/MMXU1.TotW.mag",
		Value:   100.0,
		Quality: QualityGood,
	}

	err := handler.ProcessDataChange(change)
	require.NoError(t, err)

	reports := handler.GetPendingReports("brcb01")
	assert.Equal(t, 0, len(reports))
}

func TestDataAttribute_SetAndGetVarious(t *testing.T) {
	da := NewDataAttribute("mag", "Float32")
	err := da.SetFloatValue(123.45)
	assert.NoError(t, err)
	val, err := da.GetFloatValue()
	assert.NoError(t, err)
	assert.Equal(t, 123.45, val)

	da2 := NewDataAttribute("stVal", "Int32")
	err = da2.SetIntValue(42)
	assert.NoError(t, err)
	val2, err := da2.GetIntValue()
	assert.NoError(t, err)
	assert.Equal(t, int64(42), val2)

	da3 := NewDataAttribute("boolVal", "Boolean")
	err = da3.SetBoolValue(true)
	assert.NoError(t, err)
	val3, err := da3.GetBoolValue()
	assert.NoError(t, err)
	assert.True(t, val3)
}

func TestDataAttribute_GetFloatValue_NotSet(t *testing.T) {
	da := NewDataAttribute("mag", "Float32")
	_, err := da.GetFloatValue()
	assert.Error(t, err)
}

func TestDataAttribute_GetIntValue_NotSet(t *testing.T) {
	da := NewDataAttribute("stVal", "Int32")
	_, err := da.GetIntValue()
	assert.Error(t, err)
}

func TestDataAttribute_GetBoolValue_NotSet(t *testing.T) {
	da := NewDataAttribute("stVal", "Boolean")
	_, err := da.GetBoolValue()
	assert.Error(t, err)
}
