package iec61850

import (
	"testing"
	"time"
)

// TestIEC61850Integration tests the complete IEC 61850 module integration
func TestIEC61850Integration(t *testing.T) {
	t.Run("complete workflow", func(t *testing.T) {
		// Create configuration
		config := DefaultConfig()
		config.IEDName = "TestIED"
		config.IPAddress = "192.168.1.100"
		
		// Create data model
		model := NewDataModel()
		model.IEDName = "TestIED"
		
		// Create logical device
		ld := NewLogicalDevice("LD0", "Metering")
		
		// Create logical nodes
		mmxu := NewLogicalNode("MMXU", "1")
		mmxu.AddDataObject(NewDataObject("TotW", "MX"))
		mmxu.AddDataObject(NewDataObject("TotV", "MX"))
		
		ld.AddLogicalNode(mmxu)
		model.AddLogicalDevice(ld)
		
		// Verify model
		if model.IEDName != "TestIED" {
			t.Errorf("expected IED name TestIED, got %s", model.IEDName)
		}
		
		retrievedLD, exists := model.GetLogicalDevice("LD0")
		if !exists {
			t.Fatal("logical device LD0 should exist")
		}
		
		retrievedLN, exists := retrievedLD.GetLogicalNode("MMXU1")
		if !exists {
			t.Fatal("logical node MMXU1 should exist")
		}
		
		if len(retrievedLN.DataObjects) != 2 {
			t.Errorf("expected 2 data objects, got %d", len(retrievedLN.DataObjects))
		}
	})

	t.Run("report and control integration", func(t *testing.T) {
		config := DefaultConfig()
		
		// Create handlers
		reportHandler := NewReportHandler(config)
		controlHandler := NewControlHandler(config)
		
		// Create and configure BRCB
		brcb := NewBRCB("brcb01", "LD0")
		brcb.SetDataset("dsStatus")
		brcb.SetTriggerOptions(TriggerDataChange | TriggerQualityChange)
		brcb.SetIntegrityPeriod(5000)
		
		reportHandler.AddBRCB(brcb)
		reportHandler.EnableReport("brcb01")
		
		// Create control object
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelSBONormal)
		controlHandler.AddControlObject(ctlObj)
		
		// Simulate data change
		dataChange := &DataChange{
			Ref:     "LD0/MMXU1.TotW.mag",
			Value:   100.5,
			Quality: QualityGood,
		}
		
		err := reportHandler.ProcessDataChange(dataChange)
		if err != nil {
			t.Errorf("failed to process data change: %v", err)
		}
		
		// Verify report was generated
		reports := reportHandler.GetPendingReports("brcb01")
		if len(reports) == 0 {
			t.Error("should have generated a report")
		}
		
		// Test control operation
		err = controlHandler.Select("LD0/XCBR1.Pos")
		if err != nil {
			t.Errorf("failed to select control object: %v", err)
		}
		
		err = controlHandler.Operate("LD0/XCBR1.Pos", true)
		if err != nil {
			t.Errorf("failed to operate control object: %v", err)
		}
		
		if controlHandler.GetControlState("LD0/XCBR1.Pos") != ControlStateCompleted {
			t.Error("control should be completed")
		}
	})

	t.Run("SV and report integration", func(t *testing.T) {
		config := DefaultConfig()
		
		// Create handlers
		svHandler := NewSampledValuesHandler(config)
		reportHandler := NewReportHandler(config)
		
		// Configure SV handler
		mcConfig := &SVMulticastConfig{
			GroupAddr: "224.0.0.1",
			Port:      102,
			Interface: "eth0",
		}
		svHandler.ConfigureMulticast(mcConfig)
		svHandler.SetSyncStatus(true, 1e-9)
		
		// Subscribe to SV stream
		svReceived := false
		callback := func(data *SVData) error {
			svReceived = true
			return nil
		}
		
		err := svHandler.Subscribe("SV01", callback)
		if err != nil {
			t.Errorf("failed to subscribe: %v", err)
		}
		
		// Process SV frame
		frame := &SVFrame{
			AppID:     0x4000,
			SmpCnt:    1234,
			ConfRev:   1,
			SmpRate:   4000,
			Dataset:   "SV01",
			ASDUCount: 1,
			ASDUs: []SVASDU{
				{
					SV: []SVValue{
						{Inst: 1, Value: 100.5, Quality: QualityGood},
						{Inst: 2, Value: 200.3, Quality: QualityGood},
					},
				},
			},
		}
		
		err = svHandler.ProcessFrame(frame)
		if err != nil {
			t.Errorf("failed to process SV frame: %v", err)
		}
		
		if !svReceived {
			t.Error("should have received SV data")
		}
		
		// Configure report for SV data
		brcb := NewBRCB("svReport", "LD0")
		brcb.SetTriggerOptions(TriggerDataUpdate)
		reportHandler.AddBRCB(brcb)
		reportHandler.EnableReport("svReport")
		
		// Verify synchronization status
		syncStatus := svHandler.GetSyncStatus()
		if !syncStatus.Synchronized {
			t.Error("should be synchronized")
		}
	})

	t.Run("timestamp handling", func(t *testing.T) {
		// Create timestamp
		now := time.Now()
		ts := NewTimestamp(now)
		
		// Verify timestamp
		if !ts.Quality.IsGood() {
			t.Error("timestamp quality should be good")
		}
		
		// Convert back to time
		converted := ts.ToTime()
		
		// Verify conversion (allow 1ms tolerance)
		diff := now.Sub(converted)
		if diff < 0 {
			diff = -diff
		}
		
		if diff > time.Millisecond {
			t.Errorf("timestamp conversion error: %v", diff)
		}
	})

	t.Run("reference parsing and formatting", func(t *testing.T) {
		refs := []string{
			"LD0/MMXU1.TotW.mag",
			"LD0/XCBR1.Pos.stVal",
			"LD1/MMTR1.TotWh.mag",
		}
		
		for _, ref := range refs {
			parsed, err := ParseReference(ref)
			if err != nil {
				t.Errorf("failed to parse reference %s: %v", ref, err)
				continue
			}
			
			formatted := parsed.String()
			if formatted != ref {
				t.Errorf("reference mismatch: expected %s, got %s", ref, formatted)
			}
		}
	})

	t.Run("quality flags", func(t *testing.T) {
		// Test quality flags
		qualities := []struct {
			q       Quality
			isGood  bool
			isInvalid bool
			isQuestionable bool
		}{
			{QualityGood, true, false, false},
			{QualityInvalid, false, true, false},
			{QualityQuestionable, false, false, true},
			{QualityGood | QualityOld, false, false, false},
		}
		
		for _, test := range qualities {
			if test.q.IsGood() != test.isGood {
				t.Errorf("quality %d: expected IsGood=%v, got %v", test.q, test.isGood, test.q.IsGood())
			}
			
			if test.q.IsInvalid() != test.isInvalid {
				t.Errorf("quality %d: expected IsInvalid=%v, got %v", test.q, test.isInvalid, test.q.IsInvalid())
			}
			
			if test.q.IsQuestionable() != test.isQuestionable {
				t.Errorf("quality %d: expected IsQuestionable=%v, got %v", test.q, test.isQuestionable, test.q.IsQuestionable())
			}
		}
	})
}

// TestIEC61850Performance tests performance characteristics
func TestIEC61850Performance(t *testing.T) {
	t.Run("report buffer performance", func(t *testing.T) {
		buffer := NewReportBuffer(1000)
		
		start := time.Now()
		
		for i := 0; i < 1000; i++ {
			report := &Report{
				RptID: "test",
				SqNum: uint16(i),
				Time:  NewTimestamp(time.Now()),
			}
			buffer.Add(report)
		}
		
		elapsed := time.Since(start)
		
		if elapsed > 100*time.Millisecond {
			t.Errorf("buffer operations too slow: %v", elapsed)
		}
		
		if buffer.Size() != 1000 {
			t.Errorf("expected buffer size 1000, got %d", buffer.Size())
		}
	})

	t.Run("SV buffer performance", func(t *testing.T) {
		buffer := NewSVBuffer(1000)
		
		start := time.Now()
		
		for i := 0; i < 1000; i++ {
			data := &SVData{
				SmpCnt: uint16(i),
				Values: []SVValue{
					{Inst: 1, Value: float64(i), Quality: QualityGood},
				},
			}
			buffer.Add(data)
		}
		
		elapsed := time.Since(start)
		
		if elapsed > 100*time.Millisecond {
			t.Errorf("buffer operations too slow: %v", elapsed)
		}
		
		if buffer.Size() != 1000 {
			t.Errorf("expected buffer size 1000, got %d", buffer.Size())
		}
	})

	t.Run("control state machine performance", func(t *testing.T) {
		ctlObj := NewControlObject("LD0/XCBR1.Pos", ControlModelSBONormal)
		
		start := time.Now()
		
		for i := 0; i < 1000; i++ {
			ctlObj.Select()
			ctlObj.Operate(true)
			ctlObj.Reset()
		}
		
		elapsed := time.Since(start)
		
		if elapsed > 100*time.Millisecond {
			t.Errorf("state machine operations too slow: %v", elapsed)
		}
	})
}
