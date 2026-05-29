package iec61850

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionState_String(t *testing.T) {
	assert.Equal(t, "Disconnected", ConnectionStateDisconnected.String())
	assert.Equal(t, "Connecting", ConnectionStateConnecting.String())
	assert.Equal(t, "Connected", ConnectionStateConnected.String())
	assert.Equal(t, "Associating", ConnectionStateAssociating.String())
	assert.Equal(t, "Associated", ConnectionStateAssociated.String())
	assert.Equal(t, "Unknown", ConnectionState(99).String())
}

func TestControlModel_String(t *testing.T) {
	assert.Equal(t, "Direct with normal security", ControlModelDirectNormal.String())
	assert.Equal(t, "Direct with enhanced security", ControlModelDirectEnhanced.String())
	assert.Equal(t, "Select-before-operate with normal security", ControlModelSBONormal.String())
	assert.Equal(t, "Select-before-operate with enhanced security", ControlModelSBOEnhanced.String())
	assert.Equal(t, "Unknown", ControlModel(99).String())
}

func TestControlState_String(t *testing.T) {
	assert.Equal(t, "Idle", ControlStateIdle.String())
	assert.Equal(t, "Selected", ControlStateSelected.String())
	assert.Equal(t, "Operating", ControlStateOperating.String())
	assert.Equal(t, "Completed", ControlStateCompleted.String())
	assert.Equal(t, "Cancelled", ControlStateCancelled.String())
	assert.Equal(t, "Failed", ControlStateFailed.String())
	assert.Equal(t, "Unknown", ControlState(99).String())
}

func TestQuality_Methods(t *testing.T) {
	assert.True(t, QualityGood.IsGood())
	assert.False(t, QualityGood.IsInvalid())
	assert.False(t, QualityGood.IsQuestionable())

	assert.True(t, QualityInvalid.IsInvalid())
	assert.False(t, QualityInvalid.IsGood())

	assert.True(t, QualityQuestionable.IsQuestionable())
}

func TestTimestamp_NewAndConvert(t *testing.T) {
	now := time.Now()
	ts := NewTimestamp(now)
	assert.NotZero(t, ts.SecondsSinceEpoch)

	converted := ts.ToTime()
	assert.WithinDuration(t, now, converted, time.Second)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, 102, config.Port)
	assert.Equal(t, 10*time.Second, config.ConnectTimeout)
	assert.Equal(t, 5*time.Second, config.ReadTimeout)
	assert.Equal(t, 256, config.ReportBufferSize)
	assert.Equal(t, 4000, config.SVSamplingRate)
}

func TestLogicalDevice_CRUD(t *testing.T) {
	ld := NewLogicalDevice("LD0", "1")
	require.NotNil(t, ld)
	assert.Equal(t, "LD0", ld.Name)

	ln := NewLogicalNode("MMXU", "1")
	require.NoError(t, ld.AddLogicalNode(ln))

	err := ld.AddLogicalNode(ln)
	assert.Error(t, err)

	retrieved, ok := ld.GetLogicalNode("MMXU1")
	assert.True(t, ok)
	assert.Equal(t, "MMXU1", retrieved.Name)

	_, ok = ld.GetLogicalNode("nonexistent")
	assert.False(t, ok)
}

func TestLogicalNode_CRUD(t *testing.T) {
	ln := NewLogicalNode("MMXU", "1")
	require.NotNil(t, ln)
	assert.Equal(t, "MMXU1", ln.Name)

	do := NewDataObject("TotW", "MX")
	require.NoError(t, ln.AddDataObject(do))

	err := ln.AddDataObject(do)
	assert.Error(t, err)

	retrieved, ok := ln.GetDataObject("TotW")
	assert.True(t, ok)
	assert.Equal(t, "TotW", retrieved.Name)

	_, ok = ln.GetDataObject("nonexistent")
	assert.False(t, ok)
}

func TestDataObject_CRUD(t *testing.T) {
	do := NewDataObject("TotW", "MX")
	require.NotNil(t, do)

	da := NewDataAttribute("mag", "Float32")
	require.NoError(t, do.AddDataAttribute(da))

	err := do.AddDataAttribute(da)
	assert.Error(t, err)

	retrieved, ok := do.GetDataAttribute("mag")
	assert.True(t, ok)
	assert.Equal(t, "mag", retrieved.Name)

	_, ok = do.GetDataAttribute("nonexistent")
	assert.False(t, ok)
}

func TestDataAttribute_FloatValue(t *testing.T) {
	da := NewDataAttribute("mag", "Float32")
	require.NoError(t, da.SetFloatValue(42.5))

	val, err := da.GetFloatValue()
	require.NoError(t, err)
	assert.Equal(t, 42.5, val)

	da2 := NewDataAttribute("empty", "Float32")
	_, err = da2.GetFloatValue()
	assert.Error(t, err)

	da3 := NewDataAttribute("wrong", "Int32")
	da3.Value = "not a float"
	_, err = da3.GetFloatValue()
	assert.Error(t, err)
}

func TestDataAttribute_IntValue(t *testing.T) {
	da := NewDataAttribute("cnt", "Int32")
	require.NoError(t, da.SetIntValue(100))

	val, err := da.GetIntValue()
	require.NoError(t, err)
	assert.Equal(t, int64(100), val)

	da2 := NewDataAttribute("empty", "Int32")
	_, err = da2.GetIntValue()
	assert.Error(t, err)

	da3 := NewDataAttribute("wrong", "Float32")
	da3.Value = "not an int"
	_, err = da3.GetIntValue()
	assert.Error(t, err)
}

func TestDataAttribute_BoolValue(t *testing.T) {
	da := NewDataAttribute("stVal", "Boolean")
	require.NoError(t, da.SetBoolValue(true))

	val, err := da.GetBoolValue()
	require.NoError(t, err)
	assert.True(t, val)

	da2 := NewDataAttribute("empty", "Boolean")
	_, err = da2.GetBoolValue()
	assert.Error(t, err)

	da3 := NewDataAttribute("wrong", "Boolean")
	da3.Value = "not a bool"
	_, err = da3.GetBoolValue()
	assert.Error(t, err)
}

func TestParseReference(t *testing.T) {
	ref, err := ParseReference("LD0/MMXU1.TotW.mag")
	require.NoError(t, err)
	assert.Equal(t, "LD0", ref.DeviceName)
	assert.Equal(t, "MMXU1", ref.LNName)
	assert.Equal(t, "TotW", ref.DOName)
	assert.Equal(t, "mag", ref.DAName)

	ref2, err := ParseReference("LD0/MMXU1.TotW")
	require.NoError(t, err)
	assert.Equal(t, "LD0", ref2.DeviceName)
	assert.Empty(t, ref2.DAName)

	_, err = ParseReference("invalid")
	assert.Error(t, err)

	_, err = ParseReference("LD0/no_dot")
	assert.Error(t, err)
}

func TestReference_String(t *testing.T) {
	ref := &Reference{DeviceName: "LD0", LNName: "MMXU1", DOName: "TotW", DAName: "mag"}
	assert.Equal(t, "LD0/MMXU1.TotW.mag", ref.String())

	ref2 := &Reference{DeviceName: "LD0", LNName: "MMXU1", DOName: "TotW"}
	assert.Equal(t, "LD0/MMXU1.TotW", ref2.String())
}

func TestDataSetCov(t *testing.T) {
	ds := NewDataSet("ds1", "LD0")
	require.NotNil(t, ds)
	assert.Equal(t, "ds1", ds.Name)

	ref := Reference{DeviceName: "LD0", LNName: "MMXU1", DOName: "TotW"}
	require.NoError(t, ds.AddMember(ref))
	assert.Len(t, ds.Members, 1)
}

func TestDataModel(t *testing.T) {
	dm := NewDataModel()
	require.NotNil(t, dm)

	ld := NewLogicalDevice("LD0", "1")
	require.NoError(t, dm.AddLogicalDevice(ld))

	err := dm.AddLogicalDevice(ld)
	assert.Error(t, err)

	retrieved, ok := dm.GetLogicalDevice("LD0")
	assert.True(t, ok)
	assert.Equal(t, "LD0", retrieved.Name)

	_, ok = dm.GetLogicalDevice("nonexistent")
	assert.False(t, ok)
}

func TestParseSCL(t *testing.T) {
	sclData := `<?xml version="1.0" encoding="UTF-8"?>
<SCL>
  <IED name="IED1">
    <AccessPoint name="S1">
      <Server>
        <LDevice inst="LD0">
          <LN0 lnClass="LLN0">
            <DOI name="Mod">
              <DAI name="stVal"/>
            </DOI>
          </LN0>
          <LN lnClass="MMXU" inst="1">
            <DOI name="TotW">
              <DAI name="mag"/>
            </DOI>
          </LN>
        </LDevice>
      </Server>
    </AccessPoint>
  </IED>
</SCL>`

	model, err := ParseSCL([]byte(sclData))
	require.NoError(t, err)
	assert.Equal(t, "IED1", model.IEDName)
	assert.Len(t, model.LogicalDevices, 1)

	ld, ok := model.GetLogicalDevice("LD0")
	assert.True(t, ok)
	assert.NotNil(t, ld)
}

func TestParseSCL_InvalidXML(t *testing.T) {
	_, err := ParseSCL([]byte("invalid xml"))
	assert.Error(t, err)
}

func TestParseSCL_NoIED(t *testing.T) {
	sclData := `<?xml version="1.0" encoding="UTF-8"?><SCL></SCL>`
	_, err := ParseSCL([]byte(sclData))
	assert.Error(t, err)
}

func TestBRCB_Operations(t *testing.T) {
	brcb := NewBRCB("rpt1", "LD0")
	require.NotNil(t, brcb)
	assert.Equal(t, "rpt1", brcb.RptID)
	assert.True(t, brcb.Buffered)

	require.NoError(t, brcb.SetDataset("ds1"))
	assert.Equal(t, "ds1", brcb.DatSet)

	require.NoError(t, brcb.SetTriggerOptions(TriggerDataChange | TriggerQualityChange))
	assert.True(t, brcb.TrgOps&TriggerDataChange != 0)

	require.NoError(t, brcb.SetIntegrityPeriod(5000))
	assert.Equal(t, uint32(5000), brcb.IntgPd)

	require.NoError(t, brcb.Enable())
	assert.True(t, brcb.RptEna)

	require.NoError(t, brcb.Disable())
	assert.False(t, brcb.RptEna)
}

func TestURCB_OperationsCoverage(t *testing.T) {
	urcb := NewURCB("rpt2", "LD0")
	require.NotNil(t, urcb)
	assert.Equal(t, "rpt2", urcb.RptID)
	assert.False(t, urcb.Buffered)
}

func TestReportBufferCov(t *testing.T) {
	buf := NewReportBuffer(10)
	require.NotNil(t, buf)
	assert.Equal(t, 10, buf.Capacity())
	assert.Equal(t, 0, buf.Size())

	report := &Report{
		RptID:   "rpt1",
		SqNum:   1,
		Time:    NewTimestamp(time.Now()),
		ConfRev: 1,
		Values: []ReportValue{
			{Ref: "LD0/MMXU1.TotW", Value: 42.5, Quality: QualityGood},
		},
	}
	require.NoError(t, buf.Add(report))
	assert.Equal(t, 1, buf.Size())

	retrieved := buf.Get(0)
	require.NotNil(t, retrieved)
	assert.Equal(t, "rpt1", retrieved.RptID)

	assert.Nil(t, buf.Get(5))
}

func TestControlObject_DirectNormal(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelDirectNormal)
	require.NotNil(t, co)
	assert.Equal(t, ControlStateIdle, co.State)

	err := co.Operate(true)
	require.NoError(t, err)
	assert.Equal(t, ControlStateCompleted, co.State)
	assert.Equal(t, true, co.Value)
}

func TestControlObject_DirectEnhanced(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelDirectEnhanced)
	err := co.Operate(false)
	require.NoError(t, err)
	assert.Equal(t, ControlStateCompleted, co.State)
}

func TestControlObject_SBO(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelSBONormal)
	assert.Equal(t, ControlStateIdle, co.State)

	err := co.Operate(true)
	assert.Error(t, err)

	require.NoError(t, co.Select())
	assert.Equal(t, ControlStateSelected, co.State)

	err = co.Select()
	assert.Error(t, err)

	require.NoError(t, co.Operate(true))
	assert.Equal(t, ControlStateCompleted, co.State)
}

func TestControlObject_SBOEnhanced(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelSBOEnhanced)
	require.NoError(t, co.Select())
	require.NoError(t, co.Operate(42.5))
	assert.Equal(t, ControlStateCompleted, co.State)
}

func TestControlObject_Cancel(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelSBONormal)
	require.NoError(t, co.Select())
	require.NoError(t, co.Cancel())
	assert.Equal(t, ControlStateCancelled, co.State)

	co2 := NewControlObject("LD0/GGIO1.SPCSO2", ControlModelDirectNormal)
	err := co2.Cancel()
	assert.Error(t, err)

	co3 := NewControlObject("LD0/GGIO1.SPCSO3", ControlModelDirectNormal)
	co3.Operate(true)
	err = co3.Cancel()
	assert.Error(t, err)
}

func TestControlObject_Reset(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelSBONormal)
	co.Select()
	co.Cancel()
	co.Reset()
	assert.Equal(t, ControlStateIdle, co.State)
}

func TestControlObject_SelectInvalidState(t *testing.T) {
	co := NewControlObject("LD0/GGIO1.SPCSO1", ControlModelSBONormal)
	co.State = ControlStateOperating
	err := co.Select()
	assert.Error(t, err)
}

func TestMMSClient_New(t *testing.T) {
	config := DefaultConfig()
	client := NewMMSClient(config)
	require.NotNil(t, client)
	assert.Equal(t, ConnectionStateDisconnected, client.GetState())
}

func TestMMSClient_DisconnectNotConnected(t *testing.T) {
	config := DefaultConfig()
	client := NewMMSClient(config)
	err := client.Disconnect()
	assert.NoError(t, err)
}

func TestMMSClient_AssociateNotConnected(t *testing.T) {
	config := DefaultConfig()
	client := NewMMSClient(config)
	err := client.Associate()
	assert.Equal(t, ErrNotConnected, err)
}

func TestEncodeMMSInitiateRequest(t *testing.T) {
	req := &MMSInitiateRequest{Version: 1}
	data, err := EncodeMMSInitiateRequest(req)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestDecodeMMSInitiateResponse(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 1)
	binary.BigEndian.PutUint32(data[4:8], 0)
	resp, err := DecodeMMSInitiateResponse(data)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Version)
}

func TestDecodeMMSInitiateResponse_InvalidLength(t *testing.T) {
	_, err := DecodeMMSInitiateResponse([]byte{0x01})
	assert.Error(t, err)
}

func TestSVBuffer_Operations(t *testing.T) {
	buf := NewSVBuffer(10)
	require.NotNil(t, buf)
	assert.Equal(t, 10, buf.Capacity())
	assert.Equal(t, 0, buf.Size())

	svData := &SVData{
		SmpCnt:  1,
		ConfRev: 1,
		Values:  []SVValue{{Inst: 1, Value: 3.14, Quality: QualityGood}},
	}
	require.NoError(t, buf.Add(svData))
	assert.Equal(t, 1, buf.Size())

	retrieved := buf.Get(0)
	require.NotNil(t, retrieved)
	assert.Equal(t, uint16(1), retrieved.SmpCnt)

	assert.Nil(t, buf.Get(5))
}

func TestSVBuffer_Full(t *testing.T) {
	buf := NewSVBuffer(3)
	for i := 0; i < 5; i++ {
		buf.Add(&SVData{SmpCnt: uint16(i)})
	}
	assert.Equal(t, 3, buf.Size())
}

func TestSVBuffer_LatestCov(t *testing.T) {
	buf := NewSVBuffer(10)
	buf.Add(&SVData{SmpCnt: 1})
	buf.Add(&SVData{SmpCnt: 2})

	latest := buf.Get(buf.Size() - 1)
	require.NotNil(t, latest)
	assert.Equal(t, uint16(2), latest.SmpCnt)
}

func TestSVBuffer_Latest_EmptyCov(t *testing.T) {
	buf := NewSVBuffer(10)
	assert.Nil(t, buf.Get(0))
}

func TestSVFrame_EncodeDecode(t *testing.T) {
	frame := &SVFrame{
		AppID:     0x4000,
		Length:    0,
		Reserved1: 0,
		Reserved2: 0,
		SmpCnt:    1,
		ConfRev:   1,
		SmpMod:    1,
		SmpRate:   4000,
		ASDUCount: 1,
		ASDUs: []SVASDU{
			{SV: []SVValue{
				{Inst: 1, Value: 3.14, Quality: QualityGood},
			}},
		},
	}

	data, err := EncodeSVFrame(frame)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	decoded, err := DecodeSVFrame(data)
	require.NoError(t, err)
	assert.Equal(t, frame.AppID, decoded.AppID)
	assert.Equal(t, frame.SmpCnt, decoded.SmpCnt)
}

func TestDecodeSVFrame_InvalidLength(t *testing.T) {
	_, err := DecodeSVFrame([]byte{0x01, 0x02})
	assert.Error(t, err)
}

func TestSVSyncStatus_Struct(t *testing.T) {
	status := SVSyncStatus{Synchronized: true, Precision: 0.001}
	assert.True(t, status.Synchronized)
	assert.InDelta(t, 0.001, status.Precision, 0.0001)
}

func TestSVMulticastConfig_Struct(t *testing.T) {
	config := SVMulticastConfig{
		GroupAddr: "224.0.0.1",
		Port:      102,
		Interface: "eth0",
	}
	assert.Equal(t, "224.0.0.1", config.GroupAddr)
}

func TestReportValue_Struct(t *testing.T) {
	rv := ReportValue{
		Ref:     "LD0/MMXU1.TotW",
		Value:   42.5,
		Quality: QualityGood,
	}
	assert.Equal(t, "LD0/MMXU1.TotW", rv.Ref)
	assert.Equal(t, 42.5, rv.Value)
}

func TestDataChange_Struct(t *testing.T) {
	dc := DataChange{
		Ref:     "LD0/MMXU1.TotW",
		Value:   100.0,
		Quality: QualityInvalid,
	}
	assert.Equal(t, "LD0/MMXU1.TotW", dc.Ref)
	assert.True(t, dc.Quality.IsInvalid())
}

func TestTriggerOptions_Bitmask(t *testing.T) {
	opts := TriggerDataChange | TriggerQualityChange
	assert.True(t, opts&TriggerDataChange != 0)
	assert.True(t, opts&TriggerQualityChange != 0)
	assert.False(t, opts&TriggerIntegrity != 0)
}

func TestMMSMessageType_Constants(t *testing.T) {
	assert.Equal(t, MMSMessageType(0), MMSMessageTypeConfirm)
	assert.Equal(t, MMSMessageType(1), MMSMessageTypeRequest)
	assert.Equal(t, MMSMessageType(2), MMSMessageTypeResponse)
	assert.Equal(t, MMSMessageType(3), MMSMessageTypeError)
}

func TestMMSResult_Constants(t *testing.T) {
	assert.Equal(t, MMSResult(0), MMSResultSuccess)
	assert.Equal(t, MMSResult(1), MMSResultAccessDenied)
	assert.Equal(t, MMSResult(2), MMSResultObjectNonExistent)
	assert.Equal(t, MMSResult(3), MMSResultParameterError)
	assert.Equal(t, MMSResult(4), MMSResultServiceError)
}

func TestConfig_Struct(t *testing.T) {
	config := &Config{
		IEDName:        "IED1",
		IPAddress:      "192.168.1.1",
		Port:           102,
		ConnectTimeout: 10 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		ReportBufferSize: 256,
		ControlTimeout: 5 * time.Second,
		SelectTimeout:  10 * time.Second,
		SVAppID:        0x4000,
		SVBufferSize:   1024,
		SVSamplingRate: 4000,
	}
	assert.Equal(t, "IED1", config.IEDName)
	assert.Equal(t, 102, config.Port)
}
