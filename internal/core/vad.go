// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"encoding/json"
	"os/exec"
)

type TypeVAD struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

func VAD(file string) ([]TypeVAD, error) {
	cmd := exec.Command("wav2vad", file)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data []TypeVAD
	err = json.Unmarshal(output, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
