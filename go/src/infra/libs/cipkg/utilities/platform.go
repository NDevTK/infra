package utilities

import (
	"fmt"
	"runtime"
	"strings"
)

type PlatformAttribute struct {
	Key string
	Val string
}

// Platform defines the environment outside the build system, which can't be
// identified in the environment variables or commands, but will affect the
// outputs. Platform must includes operationg system and architecture, but
// also support any attributes that may affect the results (e.g. glibcABI).
type Platform struct {
	os   string
	arch string

	attributes []PlatformAttribute
}

func NewPlatform(os, arch string) *Platform {
	var info Platform
	info.Add("os", os)
	info.Add("arch", arch)
	return &info
}

func CurrentPlatform() *Platform {
	return NewPlatform(runtime.GOOS, runtime.GOARCH)
}

// Get(...) will return the value of the attribute
func (p *Platform) Get(key string) string {
	for _, attr := range p.attributes {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// Add(...) will append the attribute. If multiple attributes have same key,
// only the first added one will be returned.
func (p *Platform) Add(key, val string) {
	switch key {
	case "os":
		p.os = val
	case "arch":
		p.arch = val
	}

	p.attributes = append(p.attributes, PlatformAttribute{
		Key: key,
		Val: val,
	})
}

func (p *Platform) OS() string {
	return p.os
}

func (p *Platform) Arch() string {
	return p.arch
}

func (p *Platform) String() string {
	attrs := make([]string, 0, len(p.attributes))
	for _, attr := range p.attributes {
		attrs = append(attrs, fmt.Sprintf("%s=%s", attr.Key, attr.Val))
	}
	return strings.Join(attrs, ",")
}
