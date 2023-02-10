package proxmox

import (
	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/qemu"
)

type wrappedScsi qemu.Scsi

func (w *wrappedScsi) GetFile() string {
	return w.File
}

func (w *wrappedScsi) SetFile(f string) {
	w.File = f
}

func (w *wrappedScsi) GetMedia() string {
	if w.Media == nil {
		return ""
	}
	return string(*w.Media)
}

func (w *wrappedScsi) SetMedia(m string) {
	w.Media = qemu.PtrScsiMedia(qemu.ScsiMedia(m))
}

func (w *wrappedScsi) UnSetMedia() {
	w.Media = nil
}

func (w *wrappedScsi) GetImportFrom() string {
	if w.ImportFrom == nil {
		return ""
	}
	return string(*w.ImportFrom)
}

func (w *wrappedScsi) SetImportFrom(m string) {
	w.ImportFrom = &m
}

func (w *wrappedScsi) GetSnapshot() bool {
	if w.Snapshot == nil {
		return false
	}
	return bool(*w.Snapshot)
}

func (w *wrappedScsi) SetSnapshot(m bool) {
	w.Snapshot = proxmox.PVEBool(m)
}

func (w *wrappedScsi) SetBackup(m bool) {
	w.Backup = proxmox.PVEBool(m)
}

type wrappedIde qemu.Ide

func (w *wrappedIde) GetFile() string {
	return w.File
}

func (w *wrappedIde) SetFile(f string) {
	w.File = f
}

func (w *wrappedIde) GetMedia() string {
	if w.Media == nil {
		return ""
	}
	return string(*w.Media)
}

func (w *wrappedIde) SetMedia(m string) {
	w.Media = (*qemu.IdeMedia)(&m)
}

func (w *wrappedIde) UnSetMedia() {
	w.Media = nil
}

func (w *wrappedIde) GetImportFrom() string {
	if w.ImportFrom == nil {
		return ""
	}
	return string(*w.ImportFrom)
}

func (w *wrappedIde) SetImportFrom(m string) {
	w.ImportFrom = &m
}

func (w *wrappedIde) GetSnapshot() bool {
	if w.Snapshot == nil {
		return false
	}
	return bool(*w.Snapshot)
}

func (w *wrappedIde) SetSnapshot(m bool) {
	w.Snapshot = proxmox.PVEBool(m)
}

func (w *wrappedIde) SetBackup(m bool) {
	w.Backup = proxmox.PVEBool(m)
}

type wrappedDisk interface {
	GetFile() string
	SetFile(string)
	GetMedia() string
	SetMedia(string)
	UnSetMedia()
	GetImportFrom() string
	SetImportFrom(string)
	GetSnapshot() bool
	SetSnapshot(bool)
	SetBackup(bool)
}
