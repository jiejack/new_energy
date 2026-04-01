package iec61850

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

// LogicalDevice represents an IEC 61850 logical device
type LogicalDevice struct {
	Name          string
	Inst          string
	LogicalNodes  map[string]*LogicalNode
}

// NewLogicalDevice creates a new logical device
func NewLogicalDevice(name, inst string) *LogicalDevice {
	return &LogicalDevice{
		Name:         name,
		Inst:         inst,
		LogicalNodes: make(map[string]*LogicalNode),
	}
}

// AddLogicalNode adds a logical node to the device
func (ld *LogicalDevice) AddLogicalNode(ln *LogicalNode) error {
	if _, exists := ld.LogicalNodes[ln.Name]; exists {
		return fmt.Errorf("logical node %s already exists", ln.Name)
	}
	ld.LogicalNodes[ln.Name] = ln
	return nil
}

// GetLogicalNode retrieves a logical node by name
func (ld *LogicalDevice) GetLogicalNode(name string) (*LogicalNode, bool) {
	ln, exists := ld.LogicalNodes[name]
	return ln, exists
}

// LogicalNode represents an IEC 61850 logical node
type LogicalNode struct {
	LNClass     string
	LNInst      string
	Name        string
	DataObjects map[string]*DataObject
}

// NewLogicalNode creates a new logical node
func NewLogicalNode(lnClass, lnInst string) *LogicalNode {
	return &LogicalNode{
		LNClass:     lnClass,
		LNInst:      lnInst,
		Name:        lnClass + lnInst,
		DataObjects: make(map[string]*DataObject),
	}
}

// AddDataObject adds a data object to the logical node
func (ln *LogicalNode) AddDataObject(do *DataObject) error {
	if _, exists := ln.DataObjects[do.Name]; exists {
		return fmt.Errorf("data object %s already exists", do.Name)
	}
	ln.DataObjects[do.Name] = do
	return nil
}

// GetDataObject retrieves a data object by name
func (ln *LogicalNode) GetDataObject(name string) (*DataObject, bool) {
	do, exists := ln.DataObjects[name]
	return do, exists
}

// DataObject represents an IEC 61850 data object
type DataObject struct {
	Name           string
	DC             string // Data class
	DataAttributes map[string]*DataAttribute
}

// NewDataObject creates a new data object
func NewDataObject(name, dc string) *DataObject {
	return &DataObject{
		Name:           name,
		DC:             dc,
		DataAttributes: make(map[string]*DataAttribute),
	}
}

// AddDataAttribute adds a data attribute to the data object
func (do *DataObject) AddDataAttribute(da *DataAttribute) error {
	if _, exists := do.DataAttributes[da.Name]; exists {
		return fmt.Errorf("data attribute %s already exists", da.Name)
	}
	do.DataAttributes[da.Name] = da
	return nil
}

// GetDataAttribute retrieves a data attribute by name
func (do *DataObject) GetDataAttribute(name string) (*DataAttribute, bool) {
	da, exists := do.DataAttributes[name]
	return da, exists
}

// DataAttribute represents an IEC 61850 data attribute
type DataAttribute struct {
	Name  string
	FC    string // Functional constraint
	Type  string
	Value interface{}
}

// NewDataAttribute creates a new data attribute
func NewDataAttribute(name, dataType string) *DataAttribute {
	return &DataAttribute{
		Name: name,
		FC:   "MX", // Default to measurement
		Type: dataType,
	}
}

// SetFloatValue sets a float value
func (da *DataAttribute) SetFloatValue(value float64) error {
	da.Value = value
	return nil
}

// GetFloatValue gets a float value
func (da *DataAttribute) GetFloatValue() (float64, error) {
	if da.Value == nil {
		return 0, errors.New("value is nil")
	}
	val, ok := da.Value.(float64)
	if !ok {
		return 0, errors.New("value is not float64")
	}
	return val, nil
}

// SetIntValue sets an integer value
func (da *DataAttribute) SetIntValue(value int64) error {
	da.Value = value
	return nil
}

// GetIntValue gets an integer value
func (da *DataAttribute) GetIntValue() (int64, error) {
	if da.Value == nil {
		return 0, errors.New("value is nil")
	}
	val, ok := da.Value.(int64)
	if !ok {
		return 0, errors.New("value is not int64")
	}
	return val, nil
}

// SetBoolValue sets a boolean value
func (da *DataAttribute) SetBoolValue(value bool) error {
	da.Value = value
	return nil
}

// GetBoolValue gets a boolean value
func (da *DataAttribute) GetBoolValue() (bool, error) {
	if da.Value == nil {
		return false, errors.New("value is nil")
	}
	val, ok := da.Value.(bool)
	if !ok {
		return false, errors.New("value is not bool")
	}
	return val, nil
}

// Reference represents an IEC 61850 object reference
type Reference struct {
	DeviceName string
	LNName     string
	DOName     string
	DAName     string
}

// ParseReference parses an IEC 61850 reference string
func ParseReference(ref string) (*Reference, error) {
	// Format: LD0/MMXU1.TotW.mag
	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	deviceName := parts[0]
	remaining := parts[1]

	lnAndDO := strings.Split(remaining, ".")
	if len(lnAndDO) < 2 {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	lnName := lnAndDO[0]
	doName := lnAndDO[1]
	daName := ""
	if len(lnAndDO) > 2 {
		daName = lnAndDO[2]
	}

	return &Reference{
		DeviceName: deviceName,
		LNName:     lnName,
		DOName:     doName,
		DAName:     daName,
	}, nil
}

// String formats the reference as a string
func (r *Reference) String() string {
	if r.DAName != "" {
		return fmt.Sprintf("%s/%s.%s.%s", r.DeviceName, r.LNName, r.DOName, r.DAName)
	}
	return fmt.Sprintf("%s/%s.%s", r.DeviceName, r.LNName, r.DOName)
}

// DataSet represents an IEC 61850 dataset
type DataSet struct {
	Name    string
	Device  string
	Members []Reference
}

// NewDataSet creates a new dataset
func NewDataSet(name, device string) *DataSet {
	return &DataSet{
		Name:    name,
		Device:  device,
		Members: make([]Reference, 0),
	}
}

// AddMember adds a reference to the dataset
func (ds *DataSet) AddMember(ref Reference) error {
	ds.Members = append(ds.Members, ref)
	return nil
}

// DataModel represents the complete IEC 61850 data model
type DataModel struct {
	IEDName        string
	LogicalDevices map[string]*LogicalDevice
}

// NewDataModel creates a new data model
func NewDataModel() *DataModel {
	return &DataModel{
		LogicalDevices: make(map[string]*LogicalDevice),
	}
}

// AddLogicalDevice adds a logical device to the model
func (dm *DataModel) AddLogicalDevice(ld *LogicalDevice) error {
	if _, exists := dm.LogicalDevices[ld.Name]; exists {
		return fmt.Errorf("logical device %s already exists", ld.Name)
	}
	dm.LogicalDevices[ld.Name] = ld
	return nil
}

// GetLogicalDevice retrieves a logical device by name
func (dm *DataModel) GetLogicalDevice(name string) (*LogicalDevice, bool) {
	ld, exists := dm.LogicalDevices[name]
	return ld, exists
}

// SCL structures for XML parsing
type SCL struct {
	XMLName xml.Name `xml:"SCL"`
	IEDs    []IED    `xml:"IED"`
}

type IED struct {
	XMLName      xml.Name     `xml:"IED"`
	Name         string       `xml:"name,attr"`
	AccessPoints []AccessPoint `xml:"AccessPoint"`
}

type AccessPoint struct {
	XMLName xml.Name `xml:"AccessPoint"`
	Name    string   `xml:"name,attr"`
	Server  Server   `xml:"Server"`
}

type Server struct {
	XMLName       xml.Name        `xml:"Server"`
	LogicalDevices []LogicalDeviceXML `xml:"LDevice"`
}

type LogicalDeviceXML struct {
	XMLName xml.Name `xml:"LDevice"`
	Inst    string   `xml:"inst,attr"`
	LN0     LN0XML   `xml:"LN0"`
	LNs     []LNXML  `xml:"LN"`
}

type LN0XML struct {
	XMLName xml.Name `xml:"LN0"`
	LNClass string   `xml:"lnClass,attr"`
	DOIs    []DOIXML `xml:"DOI"`
}

type LNXML struct {
	XMLName xml.Name `xml:"LN"`
	LNClass string   `xml:"lnClass,attr"`
	Inst    string   `xml:"inst,attr"`
	DOIs    []DOIXML `xml:"DOI"`
}

type DOIXML struct {
	XMLName xml.Name `xml:"DOI"`
	Name    string   `xml:"name,attr"`
	DAIs    []DAIXML `xml:"DAI"`
}

type DAIXML struct {
	XMLName xml.Name `xml:"DAI"`
	Name    string   `xml:"name,attr"`
}

// ParseSCL parses an SCL (Substation Configuration Language) file
func ParseSCL(data []byte) (*DataModel, error) {
	var scl SCL
	if err := xml.Unmarshal(data, &scl); err != nil {
		return nil, fmt.Errorf("failed to parse SCL: %w", err)
	}

	if len(scl.IEDs) == 0 {
		return nil, errors.New("no IED found in SCL")
	}

	model := NewDataModel()
	model.IEDName = scl.IEDs[0].Name

	for _, ied := range scl.IEDs {
		for _, ap := range ied.AccessPoints {
			for _, ldXML := range ap.Server.LogicalDevices {
				ld := NewLogicalDevice(ldXML.Inst, ldXML.Inst)

				// Process LN0
				ln0 := NewLogicalNode(ldXML.LN0.LNClass, "0")
				for _, doi := range ldXML.LN0.DOIs {
					do := NewDataObject(doi.Name, "MX")
					for _, dai := range doi.DAIs {
						da := NewDataAttribute(dai.Name, "MX")
						_ = do.AddDataAttribute(da)
					}
					_ = ln0.AddDataObject(do)
				}
				_ = ld.AddLogicalNode(ln0)

				// Process other LNs
				for _, lnXML := range ldXML.LNs {
					ln := NewLogicalNode(lnXML.LNClass, lnXML.Inst)
					for _, doi := range lnXML.DOIs {
						do := NewDataObject(doi.Name, "MX")
						for _, dai := range doi.DAIs {
							da := NewDataAttribute(dai.Name, "MX")
							_ = do.AddDataAttribute(da)
						}
						_ = ln.AddDataObject(do)
					}
					_ = ld.AddLogicalNode(ln)
				}

				_ = model.AddLogicalDevice(ld)
			}
		}
	}

	return model, nil
}
