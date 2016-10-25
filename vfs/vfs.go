// Package vfs is for storing files on the cozy, including binary ones like
// photos and movies. The range of possible operations with this endpoint goes
// from simple ones, like uploading a file, to more complex ones, like renaming
// a folder. It also ensure that an instance is not exceeding its quota, and
// keeps a trash to recover files recently deleted.
package vfs

import (
	"path"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// ForbiddenFilenameChars is the list of forbidden characters in a filename.
const ForbiddenFilenameChars = "/\x00"

// RootFolderID is the identifier of the root folder
const RootFolderID = "io.cozy.folders-root"

// DocType is the type of document, eg. file or folder
type DocType string

const (
	// FileDocType is document type
	FileDocType DocType = "io.cozy.files"
	// FolderDocType is document type
	FolderDocType = "io.cozy.folders"
)

// DocMetaAttributes is a struct containing modifiable fields from
// file and directory documents.
type DocMetaAttributes struct {
	Name       string     `json:"name,omitempty"`
	FolderID   *string    `json:"folder_id,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	Executable *bool      `json:"executable,omitempty"`
}

// Context is used to convey the afero.Fs object along with the
// CouchDb database prefix.
type Context struct {
	fs afero.Fs
	db string
}

// NewContext is the constructor function for Context
func NewContext(fs afero.Fs, dbprefix string) *Context {
	return &Context{fs, dbprefix}
}

// ParseDocType is used to transform a string to a DocType.
func ParseDocType(docType string) (result DocType, err error) {
	switch docType {
	case "io.cozy.files":
		result = FileDocType
	case "io.cozy.folders":
		result = FolderDocType
	default:
		err = ErrDocTypeInvalid
	}
	return
}

func checkFileName(str string) error {
	if str == "" || strings.ContainsAny(str, ForbiddenFilenameChars) {
		return ErrIllegalFilename
	}
	return nil
}

// getFilePath is used to generate the filepath of a new file or
// directory. It will check if the given parent folderID is well
// defined is the database and filesystem and it will generate the new
// path of the wanted file, checking if there is not colision with
// existing file.
func getFilePath(c *Context, name, folderID string) (pth string, parentDoc *DirDoc, err error) {
	if err = checkFileName(name); err != nil {
		return
	}

	var parentPath string

	if folderID == "" || folderID == RootFolderID {
		parentPath = "/"
	} else {
		parentDoc, err = GetDirectoryDoc(c, folderID)
		if err != nil {
			return
		}
		parentPath = parentDoc.Path
	}

	pth = path.Join(parentPath, name)
	return
}

func appendTags(oldtags, newtags []string) []string {
	mtags := make(map[string]struct{})
	var stags []string

	addTag := func(tag string) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return
		}
		if _, ok := mtags[tag]; !ok {
			stags = append(stags, tag)
			mtags[tag] = struct{}{}
		}
	}

	for _, tag := range oldtags {
		addTag(tag)
	}
	for _, tag := range newtags {
		addTag(tag)
	}

	return stags
}
