// Copyright 2013-2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Google's Omaha application update protocol, version 3.
//
// Omaha is a poll based protocol using XML. Requests are made by clients to
// check for updates or report events of an update process. Responses are given
// by the server to provide update information, if any, or to simply
// acknowledge the receipt of event status.
//
// https://github.com/google/omaha/blob/wiki/ServerProtocol.md
package omaha

import (
	"encoding/xml"

	"github.com/coreos/mantle/version"
)

type Request struct {
	XMLName        xml.Name `xml:"request" json:"-"`
	OS             *OS      `xml:"os"`
	Apps           []*App   `xml:"app"`
	Protocol       string   `xml:"protocol,attr"`
	Version        string   `xml:"version,attr,omitempty"`
	IsMachine      string   `xml:"ismachine,attr,omitempty"`
	SessionId      string   `xml:"sessionid,attr,omitempty"`
	UserId         string   `xml:"userid,attr,omitempty"`
	InstallSource  string   `xml:"installsource,attr,omitempty"`
	TestSource     string   `xml:"testsource,attr,omitempty"`
	RequestId      string   `xml:"requestid,attr,omitempty"`
	UpdaterVersion string   `xml:"updaterversion,attr,omitempty"`
}

func NewRequest() *Request {
	return &Request{
		Protocol: "3.0",
		Version:  version.Version,
		OS: &OS{
			Platform: LocalPlatform(),
			Arch:     LocalArch(),
			// TODO(marineam): Version and ServicePack
		},
	}
}

func (r *Request) AddApp(id, version string) *App {
	a := &App{Id: id, Version: version}
	r.Apps = append(r.Apps, a)
	return a
}

type Response struct {
	XMLName  xml.Name `xml:"response" json:"-"`
	DayStart DayStart `xml:"daystart"`
	Apps     []*App   `xml:"app"`
	Protocol string   `xml:"protocol,attr"`
	Server   string   `xml:"server,attr"`
}

func NewResponse() *Response {
	return &Response{
		Protocol: "3.0",
		Server:   "mantle",
		DayStart: DayStart{ElapsedSeconds: "0"},
	}
}

type DayStart struct {
	ElapsedSeconds string `xml:"elapsed_seconds,attr"`
}

func (r *Response) AddApp(id string, status AppStatus) *App {
	a := &App{Id: id, Status: status}
	r.Apps = append(r.Apps, a)
	return a
}

type App struct {
	Ping        *Ping        `xml:"ping"`
	UpdateCheck *UpdateCheck `xml:"updatecheck"`
	Events      []*Event     `xml:"event" json:",omitempty"`
	Id          string       `xml:"appid,attr,omitempty"`
	Version     string       `xml:"version,attr,omitempty"`
	NextVersion string       `xml:"nextversion,attr,omitempty"`
	Lang        string       `xml:"lang,attr,omitempty"`
	Client      string       `xml:"client,attr,omitempty"`
	InstallAge  string       `xml:"installage,attr,omitempty"`
	Status      AppStatus    `xml:"status,attr,omitempty"`

	// update engine extensions
	Track     string `xml:"track,attr,omitempty"`
	FromTrack string `xml:"from_track,attr,omitempty"`

	// coreos update engine extensions
	BootId    string `xml:"bootid,attr,omitempty"`
	MachineID string `xml:"machineid,attr,omitempty"`
	OEM       string `xml:"oem,attr,omitempty"`
}

func (a *App) AddUpdateCheck() *UpdateCheck {
	a.UpdateCheck = new(UpdateCheck)
	return a.UpdateCheck
}

func (a *App) AddPing() *Ping {
	a.Ping = new(Ping)
	return a.Ping
}

func (a *App) AddEvent() *Event {
	event := new(Event)
	a.Events = append(a.Events, event)
	return event
}

type UpdateCheck struct {
	URLs                *URLs        `xml:"urls"`
	Manifest            *Manifest    `xml:"manifest"`
	TargetVersionPrefix string       `xml:"targetversionprefix,attr,omitempty"`
	Status              UpdateStatus `xml:"status,attr,omitempty"`
}

func (u *UpdateCheck) AddURL(codebase string) *URL {
	// An intermediate struct is used instead of a "urls>url" tag simply
	// to keep Go from generating <urls></urls> if the list is empty.
	if u.URLs == nil {
		u.URLs = new(URLs)
	}
	url := &URL{CodeBase: codebase}
	u.URLs.URLs = append(u.URLs.URLs, url)
	return url
}

func (u *UpdateCheck) AddManifest(version string) *Manifest {
	u.Manifest = &Manifest{Version: version}
	return u.Manifest
}

type Ping struct {
	LastReportDays string `xml:"r,attr,omitempty"`
	Status         string `xml:"status,attr,omitempty"`
}

type OS struct {
	Platform    string `xml:"platform,attr,omitempty"`
	Version     string `xml:"version,attr,omitempty"`
	ServicePack string `xml:"sp,attr,omitempty"`
	Arch        string `xml:"arch,attr,omitempty"`
}

type Event struct {
	Type            EventType   `xml:"eventtype,attr"`
	Result          EventResult `xml:"eventresult,attr"`
	PreviousVersion string      `xml:"previousversion,attr,omitempty"`
	ErrorCode       string      `xml:"errorcode,attr,omitempty"`
	Status          string      `xml:"status,attr,omitempty"`
}

type URLs struct {
	URLs []*URL `xml:"url" json:",omitempty"`
}

type URL struct {
	CodeBase string `xml:"codebase,attr"`
}

type Manifest struct {
	Packages []*Package `xml:"packages>package"`
	Actions  []*Action  `xml:"actions>action"`
	Version  string     `xml:"version,attr"`
}

type Package struct {
	Hash     string `xml:"hash,attr"`
	Name     string `xml:"name,attr"`
	Size     uint64 `xml:"size,attr"`
	Required bool   `xml:"required,attr"`
}

func (m *Manifest) AddPackage() *Package {
	p := &Package{}
	m.Packages = append(m.Packages, p)
	return p
}

func (m *Manifest) AddAction(event string) *Action {
	a := &Action{Event: event}
	m.Actions = append(m.Actions, a)
	return a
}

type Action struct {
	Event string `xml:"event,attr"`

	// Extensions added by update_engine
	ChromeOSVersion       string `xml:"ChromeOSVersion,attr"`
	Sha256                string `xml:"sha256,attr"`
	NeedsAdmin            bool   `xml:"needsadmin,attr"`
	IsDelta               bool   `xml:"IsDelta,attr"`
	DisablePayloadBackoff bool   `xml:"DisablePayloadBackoff,attr,omitempty"`
	MetadataSignatureRsa  string `xml:"MetadataSignatureRsa,attr,omitempty"`
	MetadataSize          string `xml:"MetadataSize,attr,omitempty"`
	Deadline              string `xml:"deadline,attr,omitempty"`
}
