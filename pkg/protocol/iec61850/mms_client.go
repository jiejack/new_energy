package iec61850

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// MMSClient represents an MMS client
type MMSClient struct {
	config     *Config
	conn       net.Conn
	state      ConnectionState
	stateMutex sync.RWMutex
}

// NewMMSClient creates a new MMS client
func NewMMSClient(config *Config) *MMSClient {
	return &MMSClient{
		config: config,
		state:  ConnectionStateDisconnected,
	}
}

// GetState returns the current connection state
func (c *MMSClient) GetState() ConnectionState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

// setState sets the connection state
func (c *MMSClient) setState(state ConnectionState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.state = state
}

// Connect establishes a connection to the IED
func (c *MMSClient) Connect() error {
	c.setState(ConnectionStateConnecting)

	address := fmt.Sprintf("%s:%d", c.config.IPAddress, c.config.Port)
	
	dialer := net.Dialer{
		Timeout: c.config.ConnectTimeout,
	}
	
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		c.setState(ConnectionStateDisconnected)
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	
	c.conn = conn
	c.setState(ConnectionStateConnected)
	
	return nil
}

// Disconnect closes the connection to the IED
func (c *MMSClient) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	
	err := c.conn.Close()
	c.conn = nil
	c.setState(ConnectionStateDisconnected)
	
	return err
}

// Associate establishes MMS association
func (c *MMSClient) Associate() error {
	if c.GetState() != ConnectionStateConnected {
		return ErrNotConnected
	}
	
	c.setState(ConnectionStateAssociating)
	
	// Send initiate request
	req := &MMSInitiateRequest{
		Version: 1,
	}
	
	data, err := EncodeMMSInitiateRequest(req)
	if err != nil {
		c.setState(ConnectionStateConnected)
		return fmt.Errorf("failed to encode initiate request: %w", err)
	}
	
	// Send request
	if err := c.send(data); err != nil {
		c.setState(ConnectionStateConnected)
		return fmt.Errorf("failed to send initiate request: %w", err)
	}
	
	// Receive response
	respData, err := c.receive()
	if err != nil {
		c.setState(ConnectionStateConnected)
		return fmt.Errorf("failed to receive initiate response: %w", err)
	}
	
	// Decode response
	_, err = DecodeMMSInitiateResponse(respData)
	if err != nil {
		c.setState(ConnectionStateConnected)
		return fmt.Errorf("failed to decode initiate response: %w", err)
	}
	
	c.setState(ConnectionStateAssociated)
	return nil
}

// ReadVariable reads a single variable from the IED
func (c *MMSClient) ReadVariable(ref string) (interface{}, error) {
	if c.GetState() != ConnectionStateAssociated {
		return nil, ErrNotConnected
	}
	
	// Parse reference
	parsed, err := ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}
	
	// Create read request
	req := &MMSReadRequest{
		DomainID: parsed.DeviceName,
		ItemID:   fmt.Sprintf("%s.%s", parsed.LNName, parsed.DOName),
	}
	
	data, err := EncodeMMSReadRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode read request: %w", err)
	}
	
	// Send request
	if err := c.send(data); err != nil {
		return nil, fmt.Errorf("failed to send read request: %w", err)
	}
	
	// Receive response
	respData, err := c.receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive read response: %w", err)
	}
	
	// Decode response
	resp, err := DecodeMMSReadResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode read response: %w", err)
	}
	
	return resp.Value, nil
}

// WriteVariable writes a value to a variable on the IED
func (c *MMSClient) WriteVariable(ref string, value interface{}) error {
	if c.GetState() != ConnectionStateAssociated {
		return ErrNotConnected
	}
	
	// Parse reference
	parsed, err := ParseReference(ref)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}
	
	// Create write request
	req := &MMSWriteRequest{
		DomainID: parsed.DeviceName,
		ItemID:   fmt.Sprintf("%s.%s", parsed.LNName, parsed.DOName),
		Value:    value,
	}
	
	data, err := EncodeMMSWriteRequest(req)
	if err != nil {
		return fmt.Errorf("failed to encode write request: %w", err)
	}
	
	// Send request
	if err := c.send(data); err != nil {
		return fmt.Errorf("failed to send write request: %w", err)
	}
	
	// Receive response
	respData, err := c.receive()
	if err != nil {
		return fmt.Errorf("failed to receive write response: %w", err)
	}
	
	// Decode response
	_, err = DecodeMMSWriteResponse(respData)
	if err != nil {
		return fmt.Errorf("failed to decode write response: %w", err)
	}
	
	return nil
}

// ReadVariables reads multiple variables from the IED
func (c *MMSClient) ReadVariables(refs []string) (map[string]interface{}, error) {
	if c.GetState() != ConnectionStateAssociated {
		return nil, ErrNotConnected
	}
	
	results := make(map[string]interface{})
	
	for _, ref := range refs {
		value, err := c.ReadVariable(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", ref, err)
		}
		results[ref] = value
	}
	
	return results, nil
}

// GetServerDirectory retrieves the server directory
func (c *MMSClient) GetServerDirectory() ([]string, error) {
	if c.GetState() != ConnectionStateAssociated {
		return nil, ErrNotConnected
	}
	
	// Create get directory request
	req := &MMSGetDirectoryRequest{
		ObjectClass: "Domain",
	}
	
	data, err := EncodeMMSGetDirectoryRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode directory request: %w", err)
	}
	
	// Send request
	if err := c.send(data); err != nil {
		return nil, fmt.Errorf("failed to send directory request: %w", err)
	}
	
	// Receive response
	respData, err := c.receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive directory response: %w", err)
	}
	
	// Decode response
	resp, err := DecodeMMSGetDirectoryResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode directory response: %w", err)
	}
	
	return resp.Domains, nil
}

// GetVariableDirectory retrieves the variable directory for a domain
func (c *MMSClient) GetVariableDirectory(domain string) ([]string, error) {
	if c.GetState() != ConnectionStateAssociated {
		return nil, ErrNotConnected
	}
	
	// Create get variable directory request
	req := &MMSGetVariableDirectoryRequest{
		DomainID: domain,
	}
	
	data, err := EncodeMMSGetVariableDirectoryRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode variable directory request: %w", err)
	}
	
	// Send request
	if err := c.send(data); err != nil {
		return nil, fmt.Errorf("failed to send variable directory request: %w", err)
	}
	
	// Receive response
	respData, err := c.receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive variable directory response: %w", err)
	}
	
	// Decode response
	resp, err := DecodeMMSGetVariableDirectoryResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode variable directory response: %w", err)
	}
	
	return resp.Variables, nil
}

// send sends data over the connection
func (c *MMSClient) send(data []byte) error {
	if c.conn == nil {
		return ErrNotConnected
	}
	
	// Set write deadline
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout)); err != nil {
		return err
	}
	
	// Write length prefix
	length := uint32(len(data))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, length)
	
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	
	// Write data
	_, err := c.conn.Write(data)
	return err
}

// receive receives data from the connection
func (c *MMSClient) receive() ([]byte, error) {
	if c.conn == nil {
		return nil, ErrNotConnected
	}
	
	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout)); err != nil {
		return nil, err
	}
	
	// Read length prefix
	header := make([]byte, 4)
	if _, err := c.conn.Read(header); err != nil {
		return nil, err
	}
	
	length := binary.BigEndian.Uint32(header)
	
	// Read data
	data := make([]byte, length)
	_, err := c.conn.Read(data)
	
	return data, err
}

// MMSInitiateRequest represents an MMS initiate request
type MMSInitiateRequest struct {
	Version int
}

// MMSInitiateResponse represents an MMS initiate response
type MMSInitiateResponse struct {
	Version int
}

// MMSReadRequest represents an MMS read request
type MMSReadRequest struct {
	DomainID string
	ItemID   string
}

// MMSReadResponse represents an MMS read response
type MMSReadResponse struct {
	DomainID string
	ItemID   string
	Value    interface{}
}

// MMSWriteRequest represents an MMS write request
type MMSWriteRequest struct {
	DomainID string
	ItemID   string
	Value    interface{}
}

// MMSWriteResponse represents an MMS write response
type MMSWriteResponse struct {
	Result MMSResult
}

// MMSGetDirectoryRequest represents an MMS get directory request
type MMSGetDirectoryRequest struct {
	ObjectClass string
}

// MMSGetDirectoryResponse represents an MMS get directory response
type MMSGetDirectoryResponse struct {
	Domains []string
}

// MMSGetVariableDirectoryRequest represents an MMS get variable directory request
type MMSGetVariableDirectoryRequest struct {
	DomainID string
}

// MMSGetVariableDirectoryResponse represents an MMS get variable directory response
type MMSGetVariableDirectoryResponse struct {
	Variables []string
}

// EncodeMMSInitiateRequest encodes an MMS initiate request
func EncodeMMSInitiateRequest(req *MMSInitiateRequest) ([]byte, error) {
	// Simplified encoding for testing
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], uint32(req.Version))
	return data, nil
}

// DecodeMMSInitiateResponse decodes an MMS initiate response
func DecodeMMSInitiateResponse(data []byte) (*MMSInitiateResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}
	
	version := int(binary.BigEndian.Uint32(data[0:4]))
	return &MMSInitiateResponse{
		Version: version,
	}, nil
}

// EncodeMMSReadRequest encodes an MMS read request
func EncodeMMSReadRequest(req *MMSReadRequest) ([]byte, error) {
	// Simplified encoding for testing
	data := make([]byte, 4+len(req.DomainID)+4+len(req.ItemID))
	offset := 0
	
	binary.BigEndian.PutUint32(data[offset:], uint32(len(req.DomainID)))
	offset += 4
	copy(data[offset:], req.DomainID)
	offset += len(req.DomainID)
	
	binary.BigEndian.PutUint32(data[offset:], uint32(len(req.ItemID)))
	offset += 4
	copy(data[offset:], req.ItemID)
	
	return data, nil
}

// DecodeMMSReadResponse decodes an MMS read response
func DecodeMMSReadResponse(data []byte) (*MMSReadResponse, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}
	
	offset := 0
	domainLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	
	if len(data) < offset+domainLen+4 {
		return nil, fmt.Errorf("data too short for domain")
	}
	
	domainID := string(data[offset : offset+domainLen])
	offset += domainLen
	
	itemLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	
	if len(data) < offset+itemLen {
		return nil, fmt.Errorf("data too short for item")
	}
	
	itemID := string(data[offset : offset+itemLen])
	
	return &MMSReadResponse{
		DomainID: domainID,
		ItemID:   itemID,
		Value:    nil,
	}, nil
}

// EncodeMMSWriteRequest encodes an MMS write request
func EncodeMMSWriteRequest(req *MMSWriteRequest) ([]byte, error) {
	// Simplified encoding for testing
	data := make([]byte, 4+len(req.DomainID)+4+len(req.ItemID)+8)
	offset := 0
	
	binary.BigEndian.PutUint32(data[offset:], uint32(len(req.DomainID)))
	offset += 4
	copy(data[offset:], req.DomainID)
	offset += len(req.DomainID)
	
	binary.BigEndian.PutUint32(data[offset:], uint32(len(req.ItemID)))
	offset += 4
	copy(data[offset:], req.ItemID)
	
	return data, nil
}

// DecodeMMSWriteResponse decodes an MMS write response
func DecodeMMSWriteResponse(data []byte) (*MMSWriteResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}
	
	result := MMSResult(binary.BigEndian.Uint32(data[0:4]))
	return &MMSWriteResponse{
		Result: result,
	}, nil
}

// EncodeMMSGetDirectoryRequest encodes an MMS get directory request
func EncodeMMSGetDirectoryRequest(req *MMSGetDirectoryRequest) ([]byte, error) {
	data := make([]byte, 4+len(req.ObjectClass))
	binary.BigEndian.PutUint32(data[0:4], uint32(len(req.ObjectClass)))
	copy(data[4:], req.ObjectClass)
	return data, nil
}

// DecodeMMSGetDirectoryResponse decodes an MMS get directory response
func DecodeMMSGetDirectoryResponse(data []byte) (*MMSGetDirectoryResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}
	
	count := int(binary.BigEndian.Uint32(data[0:4]))
	domains := make([]string, count)
	
	offset := 4
	for i := 0; i < count; i++ {
		if len(data) < offset+4 {
			return nil, fmt.Errorf("data too short for domain length")
		}
		length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		
		if len(data) < offset+length {
			return nil, fmt.Errorf("data too short for domain")
		}
		domains[i] = string(data[offset : offset+length])
		offset += length
	}
	
	return &MMSGetDirectoryResponse{
		Domains: domains,
	}, nil
}

// EncodeMMSGetVariableDirectoryRequest encodes an MMS get variable directory request
func EncodeMMSGetVariableDirectoryRequest(req *MMSGetVariableDirectoryRequest) ([]byte, error) {
	data := make([]byte, 4+len(req.DomainID))
	binary.BigEndian.PutUint32(data[0:4], uint32(len(req.DomainID)))
	copy(data[4:], req.DomainID)
	return data, nil
}

// DecodeMMSGetVariableDirectoryResponse decodes an MMS get variable directory response
func DecodeMMSGetVariableDirectoryResponse(data []byte) (*MMSGetVariableDirectoryResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}
	
	count := int(binary.BigEndian.Uint32(data[0:4]))
	variables := make([]string, count)
	
	offset := 4
	for i := 0; i < count; i++ {
		if len(data) < offset+4 {
			return nil, fmt.Errorf("data too short for variable length")
		}
		length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		
		if len(data) < offset+length {
			return nil, fmt.Errorf("data too short for variable")
		}
		variables[i] = string(data[offset : offset+length])
		offset += length
	}
	
	return &MMSGetVariableDirectoryResponse{
		Variables: variables,
	}, nil
}
