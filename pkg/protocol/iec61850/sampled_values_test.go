package iec61850

import (
	"testing"
)

// TestSampledValuesHandler tests sampled values handler
func TestSampledValuesHandler(t *testing.T) {
	t.Run("create handler", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		if handler == nil {
			t.Fatal("handler should not be nil")
		}
	})

	t.Run("subscribe to SV stream", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		streamID := "SV01"
		callback := func(data *SVData) error {
			return nil
		}
		
		err := handler.Subscribe(streamID, callback)
		if err != nil {
			t.Errorf("failed to subscribe: %v", err)
		}
		
		if !handler.IsSubscribed(streamID) {
			t.Error("should be subscribed to stream")
		}
	})

	t.Run("unsubscribe from SV stream", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		streamID := "SV01"
		callback := func(data *SVData) error {
			return nil
		}
		
		_ = handler.Subscribe(streamID, callback)
		err := handler.Unsubscribe(streamID)
		
		if err != nil {
			t.Errorf("failed to unsubscribe: %v", err)
		}
		
		if handler.IsSubscribed(streamID) {
			t.Error("should not be subscribed to stream")
		}
	})

	t.Run("get subscription list", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		callback := func(data *SVData) error {
			return nil
		}
		
		_ = handler.Subscribe("SV01", callback)
		_ = handler.Subscribe("SV02", callback)
		
		list := handler.GetSubscriptionList()
		
		if len(list) != 2 {
			t.Errorf("expected 2 subscriptions, got %d", len(list))
		}
	})
}

// TestSVDataParsing tests SV data parsing
func TestSVDataParsing(t *testing.T) {
	t.Run("parse SV frame", func(t *testing.T) {
		// Create sample SV frame data
		frame := &SVFrame{
			AppID:      0x4000,
			Length:     100,
			Reserved1:  0,
			Reserved2:  0,
			SmpCnt:     1234,
			ConfRev:    1,
			SmpMod:     0,
			SmpRate:    4000,
			Dataset:    "SV01",
			ASDUCount:  1,
			ASDUs: []SVASDU{
				{
					SV: []SVValue{
						{Inst: 1, Value: 100.5, Quality: QualityGood},
						{Inst: 2, Value: 200.3, Quality: QualityGood},
					},
				},
			},
		}
		
		data, err := EncodeSVFrame(frame)
		if err != nil {
			t.Errorf("failed to encode SV frame: %v", err)
		}
		
		parsed, err := DecodeSVFrame(data)
		if err != nil {
			t.Errorf("failed to decode SV frame: %v", err)
		}
		
		if parsed.AppID != frame.AppID {
			t.Errorf("expected AppID %d, got %d", frame.AppID, parsed.AppID)
		}
		
		if parsed.SmpCnt != frame.SmpCnt {
			t.Errorf("expected SmpCnt %d, got %d", frame.SmpCnt, parsed.SmpCnt)
		}
	})

	t.Run("parse invalid SV frame", func(t *testing.T) {
		data := []byte{0x00, 0x01} // Too short
		
		_, err := DecodeSVFrame(data)
		if err == nil {
			t.Error("should return error for invalid frame")
		}
	})
}

// TestSVValue tests SV value operations
func TestSVValue(t *testing.T) {
	t.Run("create SV value", func(t *testing.T) {
		val := SVValue{
			Inst:    1,
			Value:   123.45,
			Quality: QualityGood,
		}
		
		if val.Inst != 1 {
			t.Errorf("expected inst 1, got %d", val.Inst)
		}
		if val.Value != 123.45 {
			t.Errorf("expected value 123.45, got %f", val.Value)
		}
		if !val.Quality.IsGood() {
			t.Error("quality should be good")
		}
	})
}

// TestSVMulticast tests SV multicast configuration
func TestSVMulticast(t *testing.T) {
	t.Run("configure multicast", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		mcConfig := &SVMulticastConfig{
			GroupAddr: "224.0.0.1",
			Port:      102,
			Interface: "eth0",
		}
		
		err := handler.ConfigureMulticast(mcConfig)
		if err != nil {
			t.Errorf("failed to configure multicast: %v", err)
		}
	})

	t.Run("get multicast config", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		mcConfig := &SVMulticastConfig{
			GroupAddr: "224.0.0.1",
			Port:      102,
			Interface: "eth0",
		}
		
		_ = handler.ConfigureMulticast(mcConfig)
		retrieved := handler.GetMulticastConfig()
		
		if retrieved.GroupAddr != mcConfig.GroupAddr {
			t.Errorf("expected group addr %s, got %s", mcConfig.GroupAddr, retrieved.GroupAddr)
		}
	})
}

// TestSVSynchronization tests SV synchronization
func TestSVSynchronization(t *testing.T) {
	t.Run("check synchronization status", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		status := handler.GetSyncStatus()
		
		// Initially not synchronized
		if status.Synchronized {
			t.Error("should not be synchronized initially")
		}
	})

	t.Run("set synchronization status", func(t *testing.T) {
		config := DefaultConfig()
		handler := NewSampledValuesHandler(config)
		
		handler.SetSyncStatus(true, 1e-9)
		
		status := handler.GetSyncStatus()
		
		if !status.Synchronized {
			t.Error("should be synchronized")
		}
		if status.Precision != 1e-9 {
			t.Errorf("expected precision 1e-9, got %e", status.Precision)
		}
	})
}

// TestSVBuffer tests SV buffer management
func TestSVBuffer(t *testing.T) {
	t.Run("create buffer", func(t *testing.T) {
		buffer := NewSVBuffer(100)
		
		if buffer == nil {
			t.Fatal("buffer should not be nil")
		}
		if buffer.Capacity() != 100 {
			t.Errorf("expected capacity 100, got %d", buffer.Capacity())
		}
	})

	t.Run("add data to buffer", func(t *testing.T) {
		buffer := NewSVBuffer(10)
		
		data := &SVData{
			SmpCnt: 1,
			Values: []SVValue{
				{Inst: 1, Value: 100.0, Quality: QualityGood},
			},
		}
		
		err := buffer.Add(data)
		if err != nil {
			t.Errorf("failed to add data: %v", err)
		}
		
		if buffer.Size() != 1 {
			t.Errorf("expected size 1, got %d", buffer.Size())
		}
	})

	t.Run("get data from buffer", func(t *testing.T) {
		buffer := NewSVBuffer(10)
		
		data := &SVData{
			SmpCnt: 1,
			Values: []SVValue{
				{Inst: 1, Value: 100.0, Quality: QualityGood},
			},
		}
		
		_ = buffer.Add(data)
		retrieved := buffer.Get(0)
		
		if retrieved == nil {
			t.Fatal("retrieved data should not be nil")
		}
		if retrieved.SmpCnt != 1 {
			t.Errorf("expected SmpCnt 1, got %d", retrieved.SmpCnt)
		}
	})

	t.Run("buffer overflow", func(t *testing.T) {
		buffer := NewSVBuffer(2)
		
		for i := 0; i < 3; i++ {
			data := &SVData{
				SmpCnt: uint16(i),
				Values: []SVValue{
					{Inst: 1, Value: float64(i), Quality: QualityGood},
				},
			}
			_ = buffer.Add(data)
		}
		
		// Buffer should have wrapped around
		if buffer.Size() != 2 {
			t.Errorf("expected size 2, got %d", buffer.Size())
		}
	})
}
