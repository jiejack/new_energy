package iec61850

import (
	"testing"
	"time"
)

// TestReportHandler tests report handler
func TestReportHandler(t *testing.T) {
	t.Run("create handler", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewReportHandler(config)
		
		if handler == nil {
			t.Fatal("handler should not be nil")
		}
	})

	t.Run("enable report", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewReportHandler(config)
		
		rptID := "brcb01"
		brcb := NewBRCB(rptID, "LD0")
		_ = handler.AddBRCB(brcb)
		
		err := handler.EnableReport(rptID)
		
		if err != nil {
			t.Errorf("failed to enable report: %v", err)
		}
		
		if !handler.IsReportEnabled(rptID) {
			t.Error("report should be enabled")
		}
	})

	t.Run("disable report", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewReportHandler(config)
		
		rptID := "brcb01"
		brcb := NewBRCB(rptID, "LD0")
		_ = handler.AddBRCB(brcb)
		
		_ = handler.EnableReport(rptID)
		err := handler.DisableReport(rptID)
		
		if err != nil {
			t.Errorf("failed to disable report: %v", err)
		}
		
		if handler.IsReportEnabled(rptID) {
			t.Error("report should be disabled")
		}
	})
}

// TestReportControlBlock tests report control block operations
func TestReportControlBlock(t *testing.T) {
	t.Run("create BRCB", func(t *testing.T) {
		brcb := NewBRCB("brcb01", "LD0")
		
		if brcb.RptID != "brcb01" {
			t.Errorf("expected RptID brcb01, got %s", brcb.RptID)
		}
		if brcb.Domain != "LD0" {
			t.Errorf("expected domain LD0, got %s", brcb.Domain)
		}
		if brcb.RptEna != false {
			t.Error("RptEna should be false initially")
		}
	})

	t.Run("set dataset", func(t *testing.T) {
		brcb := NewBRCB("brcb01", "LD0")
		
		err := brcb.SetDataset("dsName")
		if err != nil {
			t.Errorf("failed to set dataset: %v", err)
		}
		
		if brcb.DatSet != "dsName" {
			t.Errorf("expected dataset dsName, got %s", brcb.DatSet)
		}
	})

	t.Run("set trigger options", func(t *testing.T) {
		brcb := NewBRCB("brcb01", "LD0")
		
		options := TriggerDataChange | TriggerQualityChange
		err := brcb.SetTriggerOptions(options)
		if err != nil {
			t.Errorf("failed to set trigger options: %v", err)
		}
		
		if brcb.TrgOps != options {
			t.Errorf("expected trigger options %d, got %d", options, brcb.TrgOps)
		}
	})

	t.Run("set integrity period", func(t *testing.T) {
		brcb := NewBRCB("brcb01", "LD0")
		
		period := uint32(5000) // 5 seconds
		err := brcb.SetIntegrityPeriod(period)
		if err != nil {
			t.Errorf("failed to set integrity period: %v", err)
		}
		
		if brcb.IntgPd != period {
			t.Errorf("expected integrity period %d, got %d", period, brcb.IntgPd)
		}
	})
}

// TestReportBuffer tests report buffer operations
func TestReportBuffer(t *testing.T) {
	t.Run("create buffer", func(t *testing.T) {
		buffer := NewReportBuffer(100)
		
		if buffer == nil {
			t.Fatal("buffer should not be nil")
		}
		if buffer.Capacity() != 100 {
			t.Errorf("expected capacity 100, got %d", buffer.Capacity())
		}
	})

	t.Run("add report to buffer", func(t *testing.T) {
		buffer := NewReportBuffer(10)
		
		report := &Report{
			RptID:  "brcb01",
			SqNum:  1,
			Time:   NewTimestamp(time.Now()),
			Values: []ReportValue{
				{Ref: "LD0/MMXU1.TotW.mag", Value: 100.0, Quality: QualityGood},
			},
		}
		
		err := buffer.Add(report)
		if err != nil {
			t.Errorf("failed to add report: %v", err)
		}
		
		if buffer.Size() != 1 {
			t.Errorf("expected size 1, got %d", buffer.Size())
		}
	})

	t.Run("get report from buffer", func(t *testing.T) {
		buffer := NewReportBuffer(10)
		
		report := &Report{
			RptID:  "brcb01",
			SqNum:  1,
			Time:   NewTimestamp(time.Now()),
			Values: []ReportValue{
				{Ref: "LD0/MMXU1.TotW.mag", Value: 100.0, Quality: QualityGood},
			},
		}
		
		_ = buffer.Add(report)
		retrieved := buffer.Get(0)
		
		if retrieved == nil {
			t.Fatal("retrieved report should not be nil")
		}
		if retrieved.RptID != "brcb01" {
			t.Errorf("expected RptID brcb01, got %s", retrieved.RptID)
		}
	})
}

// TestReportGeneration tests report generation
func TestReportGeneration(t *testing.T) {
	t.Run("generate report on data change", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewReportHandler(config)
		
		brcb := NewBRCB("brcb01", "LD0")
		brcb.SetTriggerOptions(TriggerDataChange)
		_ = handler.AddBRCB(brcb)
		_ = handler.EnableReport("brcb01")
		
		// Simulate data change
		change := &DataChange{
			Ref:     "LD0/MMXU1.TotW.mag",
			Value:   100.0,
			Quality: QualityGood,
		}
		
		err := handler.ProcessDataChange(change)
		if err != nil {
			t.Errorf("failed to process data change: %v", err)
		}
		
		// Check if report was generated
		reports := handler.GetPendingReports("brcb01")
		if len(reports) == 0 {
			t.Error("should have generated a report")
		}
	})

	t.Run("generate report on quality change", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewReportHandler(config)
		
		brcb := NewBRCB("brcb01", "LD0")
		brcb.SetTriggerOptions(TriggerQualityChange)
		_ = handler.AddBRCB(brcb)
		_ = handler.EnableReport("brcb01")
		
		// Simulate quality change
		change := &DataChange{
			Ref:     "LD0/MMXU1.TotW.mag",
			Value:   100.0,
			Quality: QualityInvalid,
		}
		
		err := handler.ProcessDataChange(change)
		if err != nil {
			t.Errorf("failed to process quality change: %v", err)
		}
		
		// Check if report was generated
		reports := handler.GetPendingReports("brcb01")
		if len(reports) == 0 {
			t.Error("should have generated a report")
		}
	})
}

// TestURCB tests unbuffered report control block
func TestURCB(t *testing.T) {
	t.Run("create URCB", func(t *testing.T) {
		urcb := NewURCB("urcb01", "LD0")
		
		if urcb.RptID != "urcb01" {
			t.Errorf("expected RptID urcb01, got %s", urcb.RptID)
		}
		if urcb.Buffered != false {
			t.Error("URCB should not be buffered")
		}
	})

	t.Run("URCB should not buffer reports", func(t *testing.T) {
		urcb := NewURCB("urcb01", "LD0")
		
		report := &Report{
			RptID: "urcb01",
			SqNum: 1,
		}
		
		// URCB should not buffer when disabled
		err := urcb.AddReport(report)
		if err == nil {
			t.Error("URCB should not accept reports when disabled")
		}
	})
}

// TestReportEncoding tests report encoding
func TestReportEncoding(t *testing.T) {
	t.Run("encode report", func(t *testing.T) {
		report := &Report{
			RptID:    "brcb01",
			SqNum:    1,
			Time:     NewTimestamp(time.Now()),
			ConfRev:  1,
			Values: []ReportValue{
				{Ref: "LD0/MMXU1.TotW.mag", Value: 100.0, Quality: QualityGood},
			},
		}
		
		data, err := EncodeReport(report)
		if err != nil {
			t.Errorf("failed to encode report: %v", err)
		}
		
		if len(data) == 0 {
			t.Error("encoded data should not be empty")
		}
	})

	t.Run("decode report", func(t *testing.T) {
		report := &Report{
			RptID:    "brcb01",
			SqNum:    1,
			Time:     NewTimestamp(time.Now()),
			ConfRev:  1,
			Values: []ReportValue{
				{Ref: "LD0/MMXU1.TotW.mag", Value: 100.0, Quality: QualityGood},
			},
		}
		
		data, _ := EncodeReport(report)
		decoded, err := DecodeReport(data)
		
		if err != nil {
			t.Errorf("failed to decode report: %v", err)
		}
		
		if decoded.RptID != report.RptID {
			t.Errorf("expected RptID %s, got %s", report.RptID, decoded.RptID)
		}
	})
}
