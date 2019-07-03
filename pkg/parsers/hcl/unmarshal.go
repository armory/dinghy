package hcl

import "github.com/hashicorp/hcl"

type DinghyHcl struct {}

func (d DinghyHcl) Unmarshal(data []byte, i interface{}) error {
	err := hcl.Unmarshal(data, i)
	if err != nil {
		return err
	}
	return nil
}