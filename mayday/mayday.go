package mayday

import (
	"github.com/flatcar/mayday/mayday/plugins/symlink"
	"github.com/flatcar/mayday/mayday/tar"
	"github.com/flatcar/mayday/mayday/tarable"
)

func Run(t tar.Tar, tarables []tarable.Tarable) error {

	for _, tb := range tarables {
		// Skip symlinks which would be added as empty files
		if _, ok := tb.(*symlink.MaydaySymlink); !ok {
			t.Add(tb)
		}
		t.MaybeMakeLink(tb.Link(), tb.Name())
	}

	return nil
}
