package utils

import (
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

func (share *Share) UnmarshalText(b []byte) (err error) {
	*share, err = FromString(string(b))
	return err
}

func (share Share) MarshalText() ([]byte, error) {
	return []byte(share.ToString()), nil
}
