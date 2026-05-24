package vc

import "fmt"

type EQPreset struct {
	Name   string
	Label  string
	Filter string
}

var EQs = []EQPreset{
	{Name: "normal", Label: "Normal", Filter: ""},
	{Name: "nightcore", Label: "Nightcore", Filter: "aresample=48000,atempo=1.3,asetrate=48000*1.3"},
	{Name: "bass_boost", Label: "Bass Boost", Filter: "bass=g=10"},
	{Name: "high_treble", Label: "High Treble", Filter: "treble=g=10"},
	{Name: "slowed_reverb", Label: "Slowed+Reverb", Filter: "atempo=0.85,aecho=0.8:0.9:1000:0.3"},
	{Name: "concert", Label: "Concert", Filter: "aecho=0.8:0.9:100:0.5"},
	{Name: "vaporwave", Label: "Vaporwave", Filter: "aresample=48000,atempo=0.8,asetrate=48000*0.8"},
	{Name: "echo", Label: "Echo", Filter: "aecho=0.8:0.8:500:0.4"},
	{Name: "8d", Label: "8D", Filter: "apulsator=hz=0.1"},
	{Name: "soft", Label: "Soft", Filter: "equalizer=f=1000:t=q:w=1:g=-5"},
	{Name: "dance", Label: "Dance", Filter: "equalizer=f=40:t=q:w=1:g=8,equalizer=f=10000:t=q:w=1:g=5"},
}

func EQByName(name string) *EQPreset {
	for _, eq := range EQs {
		if eq.Name == name {
			return &eq
		}
	}
	return nil
}

func EQByLabel(label string) *EQPreset {
	for _, eq := range EQs {
		if eq.Label == label {
			return &eq
		}
	}
	return nil
}

func BuildAudioFilterChain(eqName string) string {
	if eqName == "" || eqName == "normal" {
		return ""
	}
	eq := EQByName(eqName)
	if eq == nil {
		return ""
	}
	return eq.Filter
}

func BuildAudioFilterFlag(eqName string) string {
	chain := BuildAudioFilterChain(eqName)
	if chain == "" {
		return ""
	}
	return fmt.Sprintf("-af \"%s,dynaudnorm=peak=0.95:maxgain=30\"", chain)
}

func BuildAudioFilterFlagLegacy(eqName string) string {
	chain := BuildAudioFilterChain(eqName)
	if chain == "" {
		return ""
	}
	return fmt.Sprintf("-af \"%s\"", chain)
}

func IsEQActive(eqName string) bool {
	return eqName != "" && eqName != "normal"
}
