package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/speps/go-hashids"
)

func hashId() *schema.Resource {
	return &schema.Resource{
		Create: CreateHashId,
		Read:   ReadHashId,
		Delete: schema.RemoveFromState,

		Schema: map[string]*schema.Schema{
			"salt": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sequence": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hash_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreateHashId(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*HashIdsConfig)
	config.Mutex.Lock()
	defer config.Mutex.Unlock()

	hashIdsState, err := readState(config)
	if err != nil {
		return err
	}

	hd := hashids.NewData()
	hd.Alphabet = hashIdsState.Alphabet
	hd.Salt = hashIdsState.Salt
	hd.MinLength = hashIdsState.MinLength

	h := hashids.NewWithData(hd)
	sequence := hashIdsState.Sequence + 1
	id, encodeErr := h.Encode([]int{sequence})
	if encodeErr != nil {
		return encodeErr
	}
	hashIdsState.Sequence = sequence
	err = writeStateFile(hashIdsState, config)
	if err != nil {
		return err
	}

	d.Set("salt", hashIdsState.Salt)
	d.Set("sequence", sequence)
	d.Set("hash_id", id)
	d.SetId(id)

	return nil
}

func ReadHashId(d *schema.ResourceData, meta interface{}) error {
	return nil
}
