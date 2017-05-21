package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"io/ioutil"
	"os"
	"sync"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"salt": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alphabet": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "abcdefghijklmnopqrstuvwxyz0123456789",
			},
			"min_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4,
			},
			"state_file": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "./hashids-state.json",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"hashid": hashId(),
		},
		ConfigureFunc: providerConfigure,
	}
}

type HashIdsState struct {
	Alphabet  string
	MinLength int
	Salt      string
	Sequence  int
}

type HashIdsConfig struct {
	Mutex     *sync.Mutex
	StateFile string
}

func createSalt() (string, error) {
	byteLength := 64
	bytes := make([]byte, byteLength)

	n, err := rand.Reader.Read(bytes)
	if n != byteLength {
		return "", errors.New("generated insufficient random bytes")
	}
	if err != nil {
		return "", errwrap.Wrapf("error generating random bytes: {{err}}", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func writeStateFile(state *HashIdsState, config *HashIdsConfig) error {
	content, err := json.Marshal(state)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config.StateFile, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readState(config *HashIdsConfig) (*HashIdsState, error) {
	if _, err := os.Stat(config.StateFile); err == nil {
		dat, readErr := ioutil.ReadFile(config.StateFile)
		if readErr != nil {
			return nil, readErr
		}
		var state HashIdsState
		parseErr := json.Unmarshal(dat, &state)
		if parseErr != nil {
			return nil, parseErr
		} else {
			return &state, nil
		}
	} else {
		return nil, nil
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	salt, hasSalt := d.GetOk("salt")
	if !hasSalt {
		newSalt, err := createSalt()
		if err != nil {
			return nil, err
		}
		salt = newSalt
	}

	fileMutex := &sync.Mutex{}
	stateFile := d.Get("state_file").(string)

	config := &HashIdsConfig{
		StateFile: stateFile,
		Mutex:     fileMutex,
	}
	config.Mutex.Lock()
	defer config.Mutex.Unlock()

	state, err := readState(config)
	if err != nil {
		return nil, err
	}
	if state != nil {
		return config, nil
	}

	state = &HashIdsState{
		Alphabet:  d.Get("alphabet").(string),
		MinLength: d.Get("min_length").(int),
		Salt:      salt.(string),
		Sequence:  0,
	}
	err = writeStateFile(state, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
