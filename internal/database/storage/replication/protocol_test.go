package replication_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlaveRequest_EncodeDecode(t *testing.T) {
	req := replication.NewSlaveRequest(42)
	var buf bytes.Buffer

	err := req.Encode(&buf)
	require.NoError(t, err, "Encoding should not produce an error")

	var decodedReq replication.SlaveRequest
	err = decodedReq.Decode(&buf)
	require.NoError(t, err, "Decoding should not produce an error")

	assert.Equal(t, req, decodedReq, "Decoded request should match original")
}

func TestMasterResponse_EncodeDecode(t *testing.T) {
	resp := replication.NewMasterResponse(true, 42, []byte("test data"))
	var buf bytes.Buffer

	err := resp.Encode(&buf)
	require.NoError(t, err, "Encoding should not produce an error")

	var decodedResp replication.MasterResponse
	err = decodedResp.Decode(&buf)
	require.NoError(t, err, "Decoding should not produce an error")

	assert.Equal(t, resp, decodedResp, "Decoded response should match original")
}

func TestSlaveRequestGobEncoding(t *testing.T) {
	req := replication.NewSlaveRequest(10)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(req)
	require.NoError(t, err, "Gob encoding should not produce an error")

	var decodedReq replication.SlaveRequest
	err = dec.Decode(&decodedReq)
	require.NoError(t, err, "Gob decoding should not produce an error")

	assert.Equal(t, req, decodedReq, "Decoded request should match original")
}

func TestMasterResponseGobEncoding(t *testing.T) {
	resp := replication.NewMasterResponse(false, 20, []byte("hello world"))
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(resp)
	require.NoError(t, err, "Gob encoding should not produce an error")

	var decodedResp replication.MasterResponse
	err = dec.Decode(&decodedResp)
	require.NoError(t, err, "Gob decoding should not produce an error")

	assert.Equal(t, resp, decodedResp, "Decoded response should match original")
}
