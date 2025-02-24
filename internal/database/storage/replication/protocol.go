package replication

import (
	"encoding/gob"
	"fmt"
	"io"
)

// SlaveRequest - struct for request from slave node
type SlaveRequest struct {
	SegmentNum int
}

// NewSlaveRequest - returns new slave request
func NewSlaveRequest(num int) SlaveRequest {
	return SlaveRequest{SegmentNum: num}
}

// Encode - encodes a SlaveRequest.
func (e SlaveRequest) Encode(w io.Writer) error {
	if err := gob.NewEncoder(w).Encode(e); err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	return nil
}

// Decode - decodes a SlaveRequest.
func (e *SlaveRequest) Decode(r io.Reader) error {
	if err := gob.NewDecoder(r).Decode(e); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	return nil
}

// MasterResponse is a struct for response from master node
type MasterResponse struct {
	Succeed bool
	Data    []byte
}

// Encode - encodes a MasterResponse.
func (e MasterResponse) Encode(w io.Writer) error {
	if err := gob.NewEncoder(w).Encode(e); err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	return nil
}

// Decode - decodes a MasterResponse.
func (e *MasterResponse) Decode(r io.Reader) error {
	if err := gob.NewDecoder(r).Decode(e); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	return nil
}

// NewMasterResponse returns new master response
func NewMasterResponse(succeed bool, segmentNum int, segmentData []byte) MasterResponse {
	return MasterResponse{
		Succeed: succeed,
		Data:    segmentData,
	}
}
