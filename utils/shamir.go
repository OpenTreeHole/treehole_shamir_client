package utils

import (
	"encoding/json"
	"fmt"
	. "math/big"
)

type Share struct {
	X *Int `json:"x"`
	Y *Int `json:"y"`
}

type Shares []Share

func (share Share) ToString() string {
	return fmt.Sprintf("%d\n%d", share.X, share.Y)
}

func FromString(rawShare string) (Share, error) {
	share := Share{new(Int), new(Int)}
	_, err := fmt.Sscanf(rawShare, "%v\n%v", share.X, share.Y)
	if err != nil {
		return share, err
	}
	return share, nil
}

func (share *Share) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	s, err := FromString(string(b))
	if err != nil {
		return err
	}
	*share = s
	return nil
}

func (share Share) MarshalJson() ([]byte, error) {
	return json.Marshal(share.ToString())
}
