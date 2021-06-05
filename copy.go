package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

func CopyImage(b *Bash, dest_dir, src, device, machine, bitstream string, clean bool) {
	src = fmt.Sprint(src, "/", machine)

	if !Exists(dest_dir) {
		CreateDir(dest_dir)
	}

	if !Exists(src) {
		log.Print("Source image directory could not be found")
		runtime.Goexit()
	}

	if device != "" {
		b.Run("mount", device, dest_dir)
	}

	if clean {
		RemoveContents(dest_dir)
	}

	commonCopy(dest_dir, src, machine, bitstream)

	if device != "" {
		b.Run("udisksctl unmount -b", device)
		b.Run("udisksctl power-off -b", device)
	}
}

func CopyRemoteImage(b *Bash, src, machine, bitstream string, password, ssh_ip string, no_qspi bool) {
	//b.Run("../scripts/ssh-copy.sh", "build/tmp/deploy/images/", cfg.Machine, opt.Bitstream, opt.Password, opt.SshIP, strconv.FormatBool(opt.NoQSPI))
	src = fmt.Sprint(src, "/", machine)
	dest := MakeTmpDir()
	defer os.RemoveAll(dest)

	if !Exists(src) {
		log.Print("Source image directory could not be found")
		runtime.Goexit()
	}

	commonCopy(dest_dir, src, machine, bitstream)

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

func commonCopy(b *Bash, dest, src, machine, bitstream string) {
	//Explained in https://github.com/Xilinx/meta-xilinx/blob/master/README.booting.md#loading-via-sd

	switch machine {
	case "mercury-zx5":
		// Generate boot.bin for enclustra
		Copy(src + "/u-boot.elf", "resources/binaries")
		Copy(bitstream, "resources/binaries/fpga.bit")

		old_dir, _ := os.Getwd()
		err := os.Chdir("resources/binaries")
		LogAndExit(err)

		b.Run("mkbootimage boot.bif /tmp/boot.bin")

		err = os.Chdir(old_dir)
		LogAndExit(err)

		Copy("/tmp/boot.bin", src+"/boot.bin")
		Copy(src + "/" + machine + ".dtb",  dest + "/devicetree.dtb")
		Copy(src + "/core-image-minimal-" + machine + ".cpio.gz.u-boot", dest + "/uramdisk")
		Copy("resources/uEnv.txt", dest + "/uEnv.txt")

		// Check sizes
		b.Run("scripts/check_image_files_size.py", dest)
	case "zc702-zynq7":
		Copy(src + "/zynq-zc702.dtb", dest)
		Copy(src + "/boot.scr", dest)
		Copy(src + "/u-boot.img", dest)

		// Couple of hacks to be able to load the FPGA
		b.Run("cat " + src + "/uEnv.txt | tr -d '\t' | sed 's/bitstream_image=boot.bin/bitstream_image=fpga.bin/' > " + dest + "/uEnv.txt")
		b.Run("echo \"bootcmd=run loadfpga && run distro_bootcmd\" >> " + dest + "/uEnv.txt")
		// cp "$SRC/uEnv.txt" "$DEST" || exit 1
		Copy(src + "/core-image-minimal-"+machine + ".cpio.gz.u-boot", dest + "/uramdisk.image.gz")
	case "picozed-zynq7":
		Copy(src + "/core-image-minimal-"+machine + ".cpio.gz.u-boot", dest + "/uramdisk.image.gz")
	}
	
	// From microzed
	// cp "$SRC/u-boot.img" "$DEST"

	Copy(src + "/boot.bin", dest)
	Copy(src + "/uImage", dest)

	// Convert bit to bin. bit format is not compatible
	b.Run("python resources/fpga-bit-to-bin.py --flip  \"" + bitstream + "\" \"" + dest + "/fpga.bin\"")
}
