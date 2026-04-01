package iec61850

import (
	"testing"
)

// TestLogicalDevice tests logical device creation and operations
func TestLogicalDevice(t *testing.T) {
	t.Run("create logical device", func(t *testing.T) {
		ld := NewLogicalDevice("LD0", "Metering")
		
		if ld.Name != "LD0" {
			t.Errorf("expected name LD0, got %s", ld.Name)
		}
		if ld.Inst != "Metering" {
			t.Errorf("expected inst Metering, got %s", ld.Inst)
		}
		if ld.LogicalNodes == nil {
			t.Error("logical nodes map should not be nil")
		}
	})

	t.Run("add logical node", func(t *testing.T) {
		ld := NewLogicalDevice("LD0", "Metering")
		ln := NewLogicalNode("MMXU", "1")
		
		err := ld.AddLogicalNode(ln)
		if err != nil {
			t.Errorf("failed to add logical node: %v", err)
		}
		
		if len(ld.LogicalNodes) != 1 {
			t.Errorf("expected 1 logical node, got %d", len(ld.LogicalNodes))
		}
		
		retrieved, exists := ld.GetLogicalNode("MMXU1")
		if !exists {
			t.Error("logical node MMXU1 should exist")
		}
		if retrieved != ln {
			t.Error("retrieved logical node should be the same")
		}
	})

	t.Run("add duplicate logical node", func(t *testing.T) {
		ld := NewLogicalDevice("LD0", "Metering")
		ln1 := NewLogicalNode("MMXU", "1")
		ln2 := NewLogicalNode("MMXU", "1")
		
		_ = ld.AddLogicalNode(ln1)
		err := ld.AddLogicalNode(ln2)
		
		if err == nil {
			t.Error("should return error for duplicate logical node")
		}
	})
}

// TestLogicalNode tests logical node creation and operations
func TestLogicalNode(t *testing.T) {
	t.Run("create logical node", func(t *testing.T) {
		ln := NewLogicalNode("MMXU", "1")
		
		if ln.LNClass != "MMXU" {
			t.Errorf("expected LNClass MMXU, got %s", ln.LNClass)
		}
		if ln.LNInst != "1" {
			t.Errorf("expected LNInst 1, got %s", ln.LNInst)
		}
		if ln.Name != "MMXU1" {
			t.Errorf("expected name MMXU1, got %s", ln.Name)
		}
	})

	t.Run("add data object", func(t *testing.T) {
		ln := NewLogicalNode("MMXU", "1")
		do := NewDataObject("TotW", "MX")
		
		err := ln.AddDataObject(do)
		if err != nil {
			t.Errorf("failed to add data object: %v", err)
		}
		
		if len(ln.DataObjects) != 1 {
			t.Errorf("expected 1 data object, got %d", len(ln.DataObjects))
		}
	})
}

// TestDataObject tests data object creation and operations
func TestDataObject(t *testing.T) {
	t.Run("create data object", func(t *testing.T) {
		do := NewDataObject("TotW", "MX")
		
		if do.Name != "TotW" {
			t.Errorf("expected name TotW, got %s", do.Name)
		}
		if do.DC != "MX" {
			t.Errorf("expected DC MX, got %s", do.DC)
		}
	})

	t.Run("add data attribute", func(t *testing.T) {
		do := NewDataObject("TotW", "MX")
		da := NewDataAttribute("mag", "Float32")
		
		err := do.AddDataAttribute(da)
		if err != nil {
			t.Errorf("failed to add data attribute: %v", err)
		}
		
		if len(do.DataAttributes) != 1 {
			t.Errorf("expected 1 data attribute, got %d", len(do.DataAttributes))
		}
	})
}

// TestDataAttribute tests data attribute creation and value operations
func TestDataAttribute(t *testing.T) {
	t.Run("create data attribute", func(t *testing.T) {
		da := NewDataAttribute("mag", "Float32")
		
		if da.Name != "mag" {
			t.Errorf("expected name mag, got %s", da.Name)
		}
		if da.FC != "MX" {
			t.Errorf("expected FC MX, got %s", da.FC)
		}
		if da.Type != "Float32" {
			t.Errorf("expected type Float32, got %s", da.Type)
		}
	})

	t.Run("set and get float value", func(t *testing.T) {
		da := NewDataAttribute("mag", "Float32")
		
		err := da.SetFloatValue(123.45)
		if err != nil {
			t.Errorf("failed to set float value: %v", err)
		}
		
		val, err := da.GetFloatValue()
		if err != nil {
			t.Errorf("failed to get float value: %v", err)
		}
		
		if val != 123.45 {
			t.Errorf("expected 123.45, got %f", val)
		}
	})

	t.Run("set and get integer value", func(t *testing.T) {
		da := NewDataAttribute("stVal", "Int32")
		
		err := da.SetIntValue(42)
		if err != nil {
			t.Errorf("failed to set int value: %v", err)
		}
		
		val, err := da.GetIntValue()
		if err != nil {
			t.Errorf("failed to get int value: %v", err)
		}
		
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	t.Run("set and get boolean value", func(t *testing.T) {
		da := NewDataAttribute("stVal", "Boolean")
		
		err := da.SetBoolValue(true)
		if err != nil {
			t.Errorf("failed to set bool value: %v", err)
		}
		
		val, err := da.GetBoolValue()
		if err != nil {
			t.Errorf("failed to get bool value: %v", err)
		}
		
		if val != true {
			t.Errorf("expected true, got %v", val)
		}
	})
}

// TestReference tests reference parsing and formatting
func TestReference(t *testing.T) {
	t.Run("parse valid reference", func(t *testing.T) {
		ref := "LD0/MMXU1.TotW.mag"
		
		parsed, err := ParseReference(ref)
		if err != nil {
			t.Errorf("failed to parse reference: %v", err)
		}
		
		if parsed.DeviceName != "LD0" {
			t.Errorf("expected device name LD0, got %s", parsed.DeviceName)
		}
		if parsed.LNName != "MMXU1" {
			t.Errorf("expected LN name MMXU1, got %s", parsed.LNName)
		}
		if parsed.DOName != "TotW" {
			t.Errorf("expected DO name TotW, got %s", parsed.DOName)
		}
		if parsed.DAName != "mag" {
			t.Errorf("expected DA name mag, got %s", parsed.DAName)
		}
	})

	t.Run("parse invalid reference", func(t *testing.T) {
		ref := "invalid"
		
		_, err := ParseReference(ref)
		if err == nil {
			t.Error("should return error for invalid reference")
		}
	})

	t.Run("format reference", func(t *testing.T) {
		ref := Reference{
			DeviceName: "LD0",
			LNName:     "MMXU1",
			DOName:     "TotW",
			DAName:     "mag",
		}
		
		formatted := ref.String()
		expected := "LD0/MMXU1.TotW.mag"
		
		if formatted != expected {
			t.Errorf("expected %s, got %s", expected, formatted)
		}
	})
}

// TestDataSet tests dataset operations
func TestDataSet(t *testing.T) {
	t.Run("create dataset", func(t *testing.T) {
		ds := NewDataSet("dsName", "LD0")
		
		if ds.Name != "dsName" {
			t.Errorf("expected name dsName, got %s", ds.Name)
		}
		if ds.Device != "LD0" {
			t.Errorf("expected device LD0, got %s", ds.Device)
		}
	})

	t.Run("add member to dataset", func(t *testing.T) {
		ds := NewDataSet("dsName", "LD0")
		ref := Reference{
			DeviceName: "LD0",
			LNName:     "MMXU1",
			DOName:     "TotW",
			DAName:     "mag",
		}
		
		err := ds.AddMember(ref)
		if err != nil {
			t.Errorf("failed to add member: %v", err)
		}
		
		if len(ds.Members) != 1 {
			t.Errorf("expected 1 member, got %d", len(ds.Members))
		}
	})
}

// TestSCLParsing tests SCL file parsing
func TestSCLParsing(t *testing.T) {
	t.Run("parse simple SCL", func(t *testing.T) {
		sclData := `<?xml version="1.0" encoding="UTF-8"?>
		<SCL xmlns="http://www.iec.ch/61850/2003/SCL">
			<IED name="IED1">
				<AccessPoint name="AP1">
					<Server>
						<LDevice inst="LD0">
							<LN0 lnClass="LLN0">
								<DOI name="Mod"/>
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
		if err != nil {
			t.Errorf("failed to parse SCL: %v", err)
		}
		
		if model == nil {
			t.Fatal("model should not be nil")
		}
		
		if model.IEDName != "IED1" {
			t.Errorf("expected IED name IED1, got %s", model.IEDName)
		}
		
		if len(model.LogicalDevices) != 1 {
			t.Errorf("expected 1 logical device, got %d", len(model.LogicalDevices))
		}
	})
}
