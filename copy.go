package main

import (
	"log"
	"os"
	"runtime"
)

type CopyImage struct {
	b         *Bash
	src       string
	machine   string
	bitstream string
}

func (c CopyImage) Local(destDir, device string, clean bool) {
	if !Exists(destDir) {
		CreateDir(destDir)
	}

	if !Exists(c.src) {
		log.Print("Source image directory could not be found: " + c.src)
		runtime.Goexit()
	}

	if device != "" {
		c.b.Run("mount", device, destDir)
	}

	if clean {
		RemoveContents(destDir)
	}

	c.commonCopy(destDir)

	if device != "" {
		c.b.Run("udisksctl unmount -b", device)
		c.b.Run("udisksctl power-off -b", device)
	}
}

func (c CopyImage) Remote(password, ssh_ip string, no_qspi bool) {
	//b.Run("../scripts/ssh-copy.sh", "build/tmp/deploy/images/", cfg.Machine, opt.Bitstream, opt.Password, opt.SshIP, strconv.FormatBool(opt.NoQSPI))
	dest := MakeTmpDir()
	defer os.RemoveAll(dest)

	if !Exists(c.src) {
		log.Print("Source image directory could not be found: " + c.src)
		runtime.Goexit()
	}

	c.commonCopy(dest)

	ssh := NewSshSession(ssh_ip, password)
	defer ssh.Close()

	// TODO use tar or something to transmit all files and after all files
	// have been transmited extract them. This way we avoid a condition
	// where one file is transmitted correctly and then, because of a network error
	// a second file can not be transmitted, which may cause the system to not be able
	// to boot if the system is restarted

	// TOOD: Update file names according to the board type
	ssh.Run("mount /dev/mmcblk0p1 /mnt")
	defer ssh.Run("umount /mnt")

	ssh.CopyFile(dest+"/boot.bin", "/mnt/boot.bin")
	ssh.CopyFile(dest+"/uramdisk", "/mnt/uramdisk")
	ssh.CopyFile(dest+"/uImage", "/mnt/uImage")
	ssh.CopyFile(dest+"/fpga.bin", "/mnt/fpga.bin")
	ssh.CopyFile(dest+"/devicetree.dtb", "/mnt/devicetree.dtb")
	ssh.CopyFile(dest+"/uEnv.txt", "/mnt/uEnv.txt")

	if !no_qspi {
		ssh.Run("flashcp -v /mnt/boot.bin /dev/mtd0")
		ssh.Run("flashcp -v /mnt/uImage /dev/mtd1")
		ssh.Run("flashcp -v /mnt/devicetree.dtb /dev/mtd2")
		ssh.Run("flashcp -v /mnt/uramdisk /dev/mtd5")
	}

	ssh.Run("killall gfaserverd gfaserver &> /dev/null")
	ssh.Run("reboot")
}

func (c CopyImage) commonCopy(dest string) {
	//Explained in https://github.com/Xilinx/meta-xilinx/blob/master/README.booting.md#loading-via-sd

	switch c.machine {
	case "mercury-zx5":
		// Generate boot.bin for enclustra
		Copy(c.src+"/u-boot.elf", "resources/binaries")
		Copy(c.bitstream, "resources/binaries/fpga.bit")

		old_dir, _ := os.Getwd()
		err := os.Chdir("resources/binaries")
		LogAndExit(err)

		c.b.Run("mkbootimage boot.bif /tmp/boot.bin")

		err = os.Chdir(old_dir)
		LogAndExit(err)

		Copy("/tmp/boot.bin", c.src+"/boot.bin")
		Copy(c.src+"/"+c.machine+".dtb", dest+"/devicetree.dtb")
		Copy(c.src+"/core-image-minimal-"+c.machine+".cpio.gz.u-boot", dest+"/uramdisk")
		Copy("resources/uEnv.txt", dest+"/uEnv.txt")

		// Check sizes
		c.b.Run("scripts/check_image_files_size.py", dest)
	case "zc702-zynq7":
		Copy(c.src+"/zynq-zc702.dtb", dest)
		Copy(c.src+"/boot.scr", dest)
		Copy(c.src+"/u-boot.img", dest)

		// Couple of hacks to be able to load the FPGA
		c.b.Run("cat " + c.src + "/uEnv.txt | tr -d '\t' | sed 's/bitstream_image=boot.bin/bitstream_image=fpga.bin/' > " + dest + "/uEnv.txt")
		c.b.Run("echo \"bootcmd=run loadfpga && run distro_bootcmd\" >> " + dest + "/uEnv.txt")
		// cp "$SRC/uEnv.txt" "$DEST" || exit 1
		Copy(c.src+"/core-image-minimal-"+c.machine+".cpio.gz.u-boot", dest+"/uramdisk.image.gz")
	case "picozed-zynq7":
		Copy(c.src+"/core-image-minimal-"+c.machine+".cpio.gz.u-boot", dest+"/uramdisk.image.gz")
	}

	// From microzed
	// cp "$SRC/u-boot.img" "$DEST"

	Copy(c.src+"/boot.bin", dest+"/boot.bin")
	Copy(c.src+"/uImage", dest+"/uImage")

	// Convert bit to bin. bit format is not compatible
	c.b.Run("python ../resources/fpga-bit-to-bin.py --flip  \"" + c.bitstream + "\" \"" + dest + "/fpga.bin\"")
}
