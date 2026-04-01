package modbus

// CalculateCRC 计算CRC16校验码（Modbus RTU使用）
// 使用CRC-16-IBM (also known as CRC-16-ANSI)多项式：0x8005
func CalculateCRC(data []byte) uint16 {
	var crc uint16 = 0xFFFF

	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

// CalculateLRC 计算LRC校验码（Modbus ASCII使用）
// LRC是所有字节的二进制补码
func CalculateLRC(data []byte) byte {
	var lrc byte

	for _, b := range data {
		lrc += b
	}

	// 取补码
	lrc = ^lrc
	lrc++

	return lrc
}

// VerifyCRC 验证CRC校验码
func VerifyCRC(data []byte, expectedCRC uint16) bool {
	calculatedCRC := CalculateCRC(data)
	return calculatedCRC == expectedCRC
}

// VerifyLRC 验证LRC校验码
func VerifyLRC(data []byte, expectedLRC byte) bool {
	calculatedLRC := CalculateLRC(data)
	return calculatedLRC == expectedLRC
}
