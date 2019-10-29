// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	// "io"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// TestBlockHeader tests the BlockHeader API.
func TestBlockHeader(t *testing.T) {
	nonce64, err := RandomUint64()
	if err != nil {
		t.Errorf("RandomUint64: Error generating nonce: %v", err)
	}
	nonce := uint32(nonce64)

	hash := mainNetGenesisHash
	merkleHash := mainNetGenesisMerkleRoot
	bits := uint32(0x1d00ffff)
	bh := NewBlockHeader(1, &hash, &merkleHash, bits, nonce)

	// Ensure we get the same data back out.
	if !bh.PrevBlock.IsEqual(&hash) {
		t.Errorf("NewBlockHeader: wrong prev hash - got %v, want %v",
			spew.Sprint(bh.PrevBlock), spew.Sprint(hash))
	}
	if !bh.MerkleRoot.IsEqual(&merkleHash) {
		t.Errorf("NewBlockHeader: wrong merkle root - got %v, want %v",
			spew.Sprint(bh.MerkleRoot), spew.Sprint(merkleHash))
	}
	if bh.Bits != bits {
		t.Errorf("NewBlockHeader: wrong bits - got %v, want %v",
			bh.Bits, bits)
	}
	if bh.Nonce != nonce {
		t.Errorf("NewBlockHeader: wrong nonce - got %v, want %v",
			bh.Nonce, nonce)
	}
}

// TestBlockHeaderWire tests the BlockHeader wire encode and decode for various
// protocol versions.
func TestBlockHeaderWire(t *testing.T) {
	nonce := uint32(123123) // 0x1e0f3
	pver := uint32(70001)

	// baseBlockHdr is used in the various tests as a baseline BlockHeader.
	bits := uint32(0x1d00ffff)
	baseBlockHdr := &BlockHeader{
		Version:    1,
		PrevBlock:  mainNetGenesisHash,
		MerkleRoot: mainNetGenesisMerkleRoot,
		Timestamp:  time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST
		Bits:       bits,
		Nonce:      nonce,
	}

	// baseBlockHdrEncoded is the wire encoded bytes of baseBlockHdr.
	baseBlockHdrEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
		0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
		0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
		0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00, // PrevBlock
		0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2,
		0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
		0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
		0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a, // MerkleRoot
		0x29, 0xab, 0x5f, 0x49, // Timestamp
		0xff, 0xff, 0x00, 0x1d, // Bits
		0xf3, 0xe0, 0x01, 0x00, // Nonce
	}

	tests := []struct {
		in   *BlockHeader    // Data to encode
		out  *BlockHeader    // Expected decoded data
		buf  []byte          // Wire encoding
		pver uint32          // Protocol version for wire encoding
		enc  MessageEncoding // Message encoding variant to use
	}{
		// Latest protocol version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			ProtocolVersion,
			BaseEncoding,
		},

		// Protocol version BIP0035Version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			BIP0035Version,
			BaseEncoding,
		},

		// Protocol version BIP0031Version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			BIP0031Version,
			BaseEncoding,
		},

		// Protocol version NetAddressTimeVersion.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			NetAddressTimeVersion,
			BaseEncoding,
		},

		// Protocol version MultipleAddressVersion.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			MultipleAddressVersion,
			BaseEncoding,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		err := writeBlockHeader(&buf, test.pver, test.in)
		if err != nil {
			t.Errorf("writeBlockHeader #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("writeBlockHeader #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		buf.Reset()
		err = test.in.BtcEncode(&buf, pver, 0)
		if err != nil {
			t.Errorf("BtcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("BtcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the block header from wire format.
		var bh BlockHeader
		rbuf := bytes.NewReader(test.buf)
		err = readBlockHeader(rbuf, test.pver, &bh)
		if err != nil {
			t.Errorf("readBlockHeader #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("readBlockHeader #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}

		rbuf = bytes.NewReader(test.buf)
		err = bh.BtcDecode(rbuf, pver, test.enc)
		if err != nil {
			t.Errorf("BtcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("BtcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}
	}
}

// TestBlockHeaderSerialize tests BlockHeader serialize and deserialize.
func TestBlockHeaderSerialize(t *testing.T) {
	nonce := uint32(123123) // 0x1e0f3

	// baseBlockHdr is used in the various tests as a baseline BlockHeader.
	bits := uint32(0x1d00ffff)
	baseBlockHdr := &BlockHeader{
		Version:    1,
		PrevBlock:  mainNetGenesisHash,
		MerkleRoot: mainNetGenesisMerkleRoot,
		Timestamp:  time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST
		Bits:       bits,
		Nonce:      nonce,
	}

	// baseBlockHdrEncoded is the wire encoded bytes of baseBlockHdr.
	baseBlockHdrEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
		0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
		0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
		0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00, // PrevBlock
		0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2,
		0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
		0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
		0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a, // MerkleRoot
		0x29, 0xab, 0x5f, 0x49, // Timestamp
		0xff, 0xff, 0x00, 0x1d, // Bits
		0xf3, 0xe0, 0x01, 0x00, // Nonce
	}

	tests := []struct {
		in  *BlockHeader // Data to encode
		out *BlockHeader // Expected decoded data
		buf []byte       // Serialized data
	}{
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the block header.
		var buf bytes.Buffer
		err := test.in.Serialize(&buf)
		if err != nil {
			t.Errorf("Serialize #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("Serialize #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Deserialize the block header.
		var bh BlockHeader
		rbuf := bytes.NewReader(test.buf)
		err = bh.Deserialize(rbuf)
		if err != nil {
			t.Errorf("Deserialize #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("Deserialize #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}
	}
}

func Test_auxpowreadBlockHeader(t *testing.T) {
	buf := []byte{
		2, 1, 98, 0, 141, 225, 251, 22, 113, 5, 247, 204, 63, 34, 242, 129, 133, 100, 178, 215, 159, 144, 246, 137, 220, 244, 162, 177, 158, 123, 221, 193, 85, 171, 114, 155, 70, 29, 140, 46, 123, 3, 23, 94, 195, 7, 94, 163, 246, 117, 242, 93, 101, 35, 108, 119, 141, 59, 135, 17, 173, 160, 198, 125, 179, 230, 240, 193, 156, 246, 237, 84, 77, 240, 3, 27, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 87, 3, 68, 57, 11, 6, 47, 80, 50, 83, 72, 47, 4, 159, 246, 237, 84, 8, 250, 190, 109, 109, 84, 104, 74, 84, 206, 51, 55, 252, 178, 126, 135, 96, 234, 54, 27, 168, 128, 44, 56, 130, 37, 229, 109, 56, 253, 237, 201, 188, 254, 237, 228, 130, 8, 0, 0, 0, 0, 0, 0, 0, 8, 5, 226, 0, 15, 0, 0, 0, 17, 47, 67, 77, 115, 102, 105, 114, 101, 51, 50, 50, 50, 54, 49, 50, 57, 47, 0, 0, 0, 0, 1, 96, 10, 41, 42, 1, 0, 0, 0, 25, 118, 169, 20, 106, 44, 168, 149, 5, 124, 59, 148, 19, 108, 135, 118, 231, 49, 206, 206, 24, 180, 108, 134, 136, 172, 0, 0, 0, 0, 142, 76, 140, 73, 155, 2, 213, 169, 194, 87, 66, 208, 10, 160, 73, 118, 47, 224, 153, 169, 197, 55, 31, 223, 40, 200, 221, 50, 250, 27, 67, 86, 4, 250, 226, 43, 193, 126, 121, 189, 213, 185, 211, 162, 30, 65, 127, 118, 178, 99, 40, 108, 49, 179, 15, 86, 120, 174, 35, 92, 26, 38, 15, 100, 6, 131, 206, 101, 10, 47, 30, 140, 27, 11, 130, 75, 67, 191, 145, 27, 222, 139, 155, 61, 41, 5, 126, 197, 254, 87, 218, 177, 103, 189, 169, 57, 126, 118, 85, 250, 103, 13, 172, 160, 232, 187, 134, 132, 25, 221, 26, 119, 33, 206, 45, 228, 164, 26, 159, 166, 59, 12, 4, 39, 138, 125, 51, 202, 254, 184, 61, 176, 5, 168, 47, 155, 70, 45, 24, 154, 5, 220, 48, 191, 168, 103, 67, 39, 11, 142, 113, 251, 74, 30, 114, 243, 205, 36, 20, 209, 246, 0, 0, 0, 0, 3, 4, 126, 206, 156, 226, 196, 42, 12, 14, 244, 137, 162, 204, 12, 188, 30, 188, 171, 96, 201, 125, 218, 154, 30, 117, 183, 20, 20, 189, 147, 136, 183, 172, 75, 210, 35, 242, 123, 254, 181, 155, 180, 217, 193, 128, 77, 164, 16, 76, 85, 26, 149, 114, 157, 33, 26, 215, 58, 22, 90, 141, 17, 129, 4, 152, 182, 159, 209, 211, 222, 92, 163, 213, 117, 67, 34, 66, 88, 251, 229, 228, 57, 54, 210, 62, 74, 164, 30, 50, 150, 14, 107, 187, 187, 99, 144, 0, 0, 0, 0, 2, 0, 0, 0, 170, 219, 42, 83, 23, 24, 8, 232, 168, 123, 244, 145, 94, 2, 225, 68, 148, 207, 11, 61, 28, 84, 107, 64, 193, 16, 25, 143, 34, 74, 15, 31, 4, 80, 78, 212, 135, 173, 80, 162, 227, 138, 3, 127, 193, 71, 203, 195, 107, 245, 6, 159, 146, 209, 91, 246, 25, 53, 136, 71, 6, 186, 66, 175, 85, 246, 237, 84, 44, 157, 1, 27, 210, 27, 64, 0,
	}
	r := bytes.NewReader(buf)
	var header BlockHeader
	readBlockHeader(r, 0, &header)
	wantheader := BlockHeader{Version: 6422786, 
		PrevBlock: [32]byte{141, 225, 251, 22, 113, 5, 247, 204, 63, 34, 242, 129, 133, 100, 178, 215, 159, 144, 246, 137, 220, 244, 162, 177, 158, 123, 221, 193, 85, 171, 114, 155},
		MerkleRoot: [32]byte{70, 29, 140, 46, 123, 3, 23, 94, 195, 7, 94, 163, 246, 117, 242, 93, 101, 35, 108, 119, 141, 59, 135, 17, 173, 160, 198, 125, 179, 230, 240, 193},
		Timestamp:  time.Unix(1424881308, 0),
		Bits:       453242957,
		Nonce:      0,
	}
	if !reflect.DeepEqual(header, wantheader) {
		t.Errorf("aux() got1 = %v, want %v", header, wantheader)
	}

}
