package push

// EncodeVarint32 prepends a varint32 length prefix to the given protobuf data.
// This is compatible with Netty's ProtobufVarint32LengthFieldPrepender.
func EncodeVarint32(data []byte) []byte {
	length := uint32(len(data))
	var header []byte
	for length > 0x7F {
		header = append(header, byte(length&0x7F)|0x80)
		length >>= 7
	}
	header = append(header, byte(length))
	result := make([]byte, len(header)+len(data))
	copy(result, header)
	copy(result[len(header):], data)
	return result
}

// DecodeVarint32 reads a varint32 length-prefixed frame from the buffer.
// It returns the decoded message, the remaining bytes, and whether decoding succeeded.
// If the buffer does not contain a complete frame, ok is false and the caller
// should wait for more data.
// This is compatible with Netty's ProtobufVarint32FrameDecoder.
func DecodeVarint32(buffer []byte) (msg []byte, remaining []byte, ok bool) {
	if len(buffer) == 0 {
		return nil, buffer, false
	}

	// Read varint32 length prefix (max 5 bytes for 32-bit value)
	var length uint32
	var shift uint
	headerLen := 0
	for i := 0; i < 5 && i < len(buffer); i++ {
		b := buffer[i]
		length |= uint32(b&0x7F) << shift
		shift += 7
		headerLen = i + 1
		if b&0x80 == 0 {
			// End of varint
			totalLen := headerLen + int(length)
			if len(buffer) < totalLen {
				// Not enough data for the full message yet
				return nil, buffer, false
			}
			msg = buffer[headerLen:totalLen]
			remaining = buffer[totalLen:]
			return msg, remaining, true
		}
	}

	// If we consumed 5 bytes and the last one still has the continuation bit,
	// the varint is malformed (exceeds 32-bit range). Return false so the
	// caller can handle the error.
	if headerLen == 5 && buffer[4]&0x80 != 0 {
		return nil, buffer, false
	}

	// Not enough bytes to read the complete varint header
	return nil, buffer, false
}
