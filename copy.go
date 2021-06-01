package main

import (
	"fmt"
	"log"
)

func CopyImage(b *Bash, dest_dir, src, device, machine, bitstream string, clean bool) {
	src = fmt.Sprint(src, "/", machine)

	if !Exists(dest_dir) {
		CreateDir(dest_dir)
	}

	if !Exists(src) {
		log.Fatal("Source image directory could not be found")
	}

	if device != "" {
		b.Run("mount", device, dest_dir)
	}

	if clean {
		RemoveContents(dest_dir)
	}

	b.Run("../scripts/common-copy.sh", dest_dir, src, machine, bitstream)

	if device != "" {
		b.Run("udisksctl unmount -b", device)
		b.Run("udisksctl power-off -b", device)
	}
}



