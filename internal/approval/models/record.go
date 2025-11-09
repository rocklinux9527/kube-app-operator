package models

import "time"

type Record struct {
    Who string `json:"who"`
    Action string `json:"action"`
    Comment string `json:"comment,omitempty"`
    StepIdx int `json:"step_idx"`
    TS time.Time `json:"ts"`
}

