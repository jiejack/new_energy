package iec61850

import (
	"net"
	"testing"
	"time"
)

// TestMMSClient tests MMS client operations
func TestMMSClient(t *testing.T) {
	t.Run("create client with config", func(t *testing.T) {
		config := DefaultConfig()
		config.IEDName = "TestIED"
		config.IPAddress = "192.168.1.100"
		
		client := NewMMSClient(config)
		
		if client == nil {
			t.Fatal("client should not be nil")
		}
		if client.config.IEDName != "TestIED" {
			t.Errorf("expected IED name TestIED, got %s", client.config.IEDName)
		}
		if client.GetState() != ConnectionStateDisconnected {
			t.Errorf("expected state Disconnected, got %s", client.GetState())
		}
	})

	t.Run("connect to invalid address", func(t *testing.T) {
		config := DefaultConfig()
		config.IPAddress = "invalid-address"
		config.ConnectTimeout = 1 * time.Second
		
		client := NewMMSClient(config)
		err := client.Connect()
		
		if err == nil {
			t.Error("should return error for invalid address")
		}
		if client.GetState() != ConnectionStateDisconnected {
			t.Errorf("expected state Disconnected, got %s", client.GetState())
		}
	})

	t.Run("disconnect when not connected", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		err := client.Disconnect()
		if err != nil {
			t.Errorf("disconnect should not return error when not connected: %v", err)
		}
	})
}

// TestMMSMessageEncoding tests MMS message encoding and decoding
func TestMMSMessageEncoding(t *testing.T) {
	t.Run("encode initiate request", func(t *testing.T) {
		req := &MMSInitiateRequest{
			Version: 1,
		}
		
		data, err := EncodeMMSInitiateRequest(req)
		if err != nil {
			t.Errorf("failed to encode initiate request: %v", err)
		}
		
		if len(data) == 0 {
			t.Error("encoded data should not be empty")
		}
	})

	t.Run("decode initiate response", func(t *testing.T) {
		// First encode a request
		req := &MMSInitiateRequest{
			Version: 1,
		}
		data, _ := EncodeMMSInitiateRequest(req)
		
		// Then decode it
		resp, err := DecodeMMSInitiateResponse(data)
		if err != nil {
			t.Errorf("failed to decode initiate response: %v", err)
		}
		
		if resp.Version != 1 {
			t.Errorf("expected version 1, got %d", resp.Version)
		}
	})

	t.Run("encode read request", func(t *testing.T) {
		req := &MMSReadRequest{
			DomainID: "LD0",
			ItemID:   "MMXU1.TotW.mag",
		}
		
		data, err := EncodeMMSReadRequest(req)
		if err != nil {
			t.Errorf("failed to encode read request: %v", err)
		}
		
		if len(data) == 0 {
			t.Error("encoded data should not be empty")
		}
	})

	t.Run("decode read response", func(t *testing.T) {
		// First encode a request
		req := &MMSReadRequest{
			DomainID: "LD0",
			ItemID:   "MMXU1.TotW.mag",
		}
		data, _ := EncodeMMSReadRequest(req)
		
		// Then decode it
		resp, err := DecodeMMSReadResponse(data)
		if err != nil {
			t.Errorf("failed to decode read response: %v", err)
		}
		
		if resp.DomainID != "LD0" {
			t.Errorf("expected domain ID LD0, got %s", resp.DomainID)
		}
	})

	t.Run("encode write request", func(t *testing.T) {
		req := &MMSWriteRequest{
			DomainID: "LD0",
			ItemID:   "MMXU1.TotW.mag",
			Value:    123.45,
		}
		
		data, err := EncodeMMSWriteRequest(req)
		if err != nil {
			t.Errorf("failed to encode write request: %v", err)
		}
		
		if len(data) == 0 {
			t.Error("encoded data should not be empty")
		}
	})
}

// TestAssociation tests MMS association
func TestAssociation(t *testing.T) {
	t.Run("association state transitions", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		// Initial state
		if client.GetState() != ConnectionStateDisconnected {
			t.Errorf("expected initial state Disconnected, got %s", client.GetState())
		}
		
		// Set state to connecting
		client.setState(ConnectionStateConnecting)
		if client.GetState() != ConnectionStateConnecting {
			t.Errorf("expected state Connecting, got %s", client.GetState())
		}
		
		// Set state to connected
		client.setState(ConnectionStateConnected)
		if client.GetState() != ConnectionStateConnected {
			t.Errorf("expected state Connected, got %s", client.GetState())
		}
	})
}

// TestVariableAccess tests variable access operations
func TestVariableAccess(t *testing.T) {
	t.Run("read variable when not connected", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		_, err := client.ReadVariable("LD0/MMXU1.TotW.mag")
		if err == nil {
			t.Error("should return error when not connected")
		}
		if err != ErrNotConnected {
			t.Errorf("expected ErrNotConnected, got %v", err)
		}
	})

	t.Run("write variable when not connected", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		err := client.WriteVariable("LD0/MMXU1.TotW.mag", 123.45)
		if err == nil {
			t.Error("should return error when not connected")
		}
		if err != ErrNotConnected {
			t.Errorf("expected ErrNotConnected, got %v", err)
		}
	})

	t.Run("read multiple variables when not connected", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		refs := []string{
			"LD0/MMXU1.TotW.mag",
			"LD0/MMXU1.TotV.mag",
		}
		
		_, err := client.ReadVariables(refs)
		if err == nil {
			t.Error("should return error when not connected")
		}
		if err != ErrNotConnected {
			t.Errorf("expected ErrNotConnected, got %v", err)
		}
	})
}

// TestMMSClientWithMockServer tests MMS client with mock server
func TestMMSClientWithMockServer(t *testing.T) {
	t.Run("connect to mock server", func(t *testing.T) {
		// Start mock server
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to start mock server: %v", err)
		}
		defer listener.Close()
		
		// Get actual port
		addr := listener.Addr().(*net.TCPAddr)
		
		// Configure client
		config := DefaultConfig()
		config.IPAddress = addr.IP.String()
		config.Port = addr.Port
		config.ConnectTimeout = 2 * time.Second
		
		client := NewMMSClient(config)
		
		// Accept connection in background
		go func() {
			conn, err := listener.Accept()
			if err == nil {
				conn.Close()
			}
		}()
		
		// Connect
		err = client.Connect()
		if err != nil {
			t.Errorf("failed to connect: %v", err)
		}
		
		// Cleanup
		_ = client.Disconnect()
	})
}

// TestGetServerDirectory tests server directory operations
func TestGetServerDirectory(t *testing.T) {
	t.Run("get directory when not connected", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		_, err := client.GetServerDirectory()
		if err == nil {
			t.Error("should return error when not connected")
		}
		if err != ErrNotConnected {
			t.Errorf("expected ErrNotConnected, got %v", err)
		}
	})
}

// TestGetVariableDirectory tests variable directory operations
func TestGetVariableDirectory(t *testing.T) {
	t.Run("get variable directory when not connected", func(t *testing.T) {
		config := DefaultConfig()
		client := NewMMSClient(config)
		
		_, err := client.GetVariableDirectory("LD0")
		if err == nil {
			t.Error("should return error when not connected")
		}
		if err != ErrNotConnected {
			t.Errorf("expected ErrNotConnected, got %v", err)
		}
	})
}
