package packages

import (
	"io/fs"
	"reflect"
)

func init() {
	Packages.Insert("io/fs", PackageMap{
		"ErrClosed":          fs.ErrClosed,
		"ErrExist":           fs.ErrExist,
		"ErrInvalid":         fs.ErrInvalid,
		"ErrNotExist":        fs.ErrNotExist,
		"ErrPermission":      fs.ErrPermission,
		"FileInfoToDirEntry": fs.FileInfoToDirEntry,
		"FormatDirEntry":     fs.FormatDirEntry,
		"FormatFileInfo":     fs.FormatFileInfo,
		"Glob":               fs.Glob,
		"ModeAppend":         fs.ModeAppend,
		"ModeCharDevice":     fs.ModeCharDevice,
		"ModeDevice":         fs.ModeDevice,
		"ModeDir":            fs.ModeDir,
		"ModeExclusive":      fs.ModeExclusive,
		"ModeIrregular":      fs.ModeIrregular,
		"ModeNamedPipe":      fs.ModeNamedPipe,
		"ModePerm":           fs.ModePerm,
		"ModeSetgid":         fs.ModeSetgid,
		"ModeSetuid":         fs.ModeSetuid,
		"ModeSocket":         fs.ModeSocket,
		"ModeSticky":         fs.ModeSticky,
		"ModeSymlink":        fs.ModeSymlink,
		"ModeTemporary":      fs.ModeTemporary,
		"ModeType":           fs.ModeType,
		"ReadDir":            fs.ReadDir,
		"ReadFile":           fs.ReadFile,
		"SkipAll":            fs.SkipAll,
		"SkipDir":            fs.SkipDir,
		"Stat":               fs.Stat,
		"Sub":                fs.Sub,
		"ValidPath":          fs.ValidPath,
		"WalkDir":            fs.WalkDir,
	})
	PackageTypes.Insert("io/fs", PackageMap{
		"DirEntry":    reflect.TypeOf((*fs.DirEntry)(nil)).Elem(),
		"FS":          reflect.TypeOf((*fs.FS)(nil)).Elem(),
		"File":        reflect.TypeOf((*fs.File)(nil)).Elem(),
		"FileInfo":    reflect.TypeOf((*fs.FileInfo)(nil)).Elem(),
		"FileMode":    reflect.TypeOf((*fs.FileMode)(nil)).Elem(),
		"PathError":   fs.PathError{},
		"ReadDirFile": reflect.TypeOf((*fs.ReadDirFile)(nil)).Elem(),
	})
}
