package cronlib

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"strings"
)

type JobPolicy uint32

func (j JobPolicy) String() string {
	if v, e := j.MarshalText(); e == nil {
		return string(v)
	}
	return "known"
}

func (j *JobPolicy) UnmarshalText(text []byte) error {
	if v, err := ParseJobPolicy(string(text)); err != nil {
		return err
	} else {
		*j = v
	}
	return nil
}

func (j JobPolicy) MarshalText() (text []byte, err error) {
	if s, ok := policyNameMap[j]; ok {
		return []byte(s), nil

	} else {
		return nil, fmt.Errorf("unknown policy value:%d", j)
	}
}

type JobPolicies []JobPolicy

func (j JobPolicies) String() string {
	var ss []string
	for _, ij := range j {
		ss = append(ss, ij.String())
	}
	if len(ss) == 0 {
		return ""
	}
	return strings.Join(ss, ",")
}

func (j JobPolicies) ToTengoSlice() *tengo.Array {
	var ss []tengo.Object
	for _, ij := range j {
		ss = append(ss, &tengo.String{Value: ij.String()})
	}
	return &tengo.Array{Value: ss}
}

const (
	PolicyEmpty JobPolicy = iota
	PolicySkipIfRunning
	PolicyDelayIfRunning
	PolicyRecover
)

var (
	policyNameMap = map[JobPolicy]string{
		PolicyEmpty:          "empty",
		PolicySkipIfRunning:  "skipIfRunning",
		PolicyDelayIfRunning: "delayIfRunning",
		PolicyRecover:        "recover",
	}
	policyValueMap = map[string]JobPolicy{
		"empty":          PolicyEmpty,
		"skipIfRunning":  PolicySkipIfRunning,
		"delayIfRunning": PolicyDelayIfRunning,
		"recover":        PolicyRecover,
	}
)

func ParseJobPolicy(name string) (JobPolicy, error) {
	if name == "" {
		return PolicyEmpty, nil
	}
	policy, ok := policyValueMap[name]
	if !ok {
		return PolicyEmpty, fmt.Errorf("unknown policy:%s", name)
	}
	return policy, nil
}
