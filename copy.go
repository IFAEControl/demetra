package main

import (
	"fmt"
	"log"
	"os"
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

func CopyRemoteImage(b *Bash, src, machine, bitstream string, password, ssh_ip string, no_qspi bool) {
	//b.Run("../scripts/ssh-copy.sh", "build/tmp/deploy/images/", cfg.Machine, opt.Bitstream, opt.Password, opt.SshIP, strconv.FormatBool(opt.NoQSPI))
	src = fmt.Sprint(src, "/", machine)
	dest := MakeTmpDir()
	defer os.RemoveAll(dest)

	if !Exists(src) {
		log.Fatal("Source image directory could not be found")
	}

	b.Run("../scripts/common-copy.sh", dest, src, machine, bitstream)

	/*


	   OLD_DIR="$PWD"
	   cd "$DEST" || exit
	   tar -cSf "$DEST/yocto.tar" boot.bin uramdisk devicetree.dtb uImage fpga.bin uEnv.txt
	   cd "$OLD_DIR" || exit

	*/

	ssh := NewSshSession(ssh_ip, password)
	defer ssh.Close()

	ssh.Run("mount /dev/mmcblk0p1 /mnt")
	ssh.CopyFile(dest+"/boot.bin", "/tmp/boot.bin")
	ssh.CopyFile(dest+"/uImage", "/tmp/uImage")
	ssh.Run("umount /mnt")

	//bin/rm -f /mnt/* || exit

	/*

	   sshpass -p"$PASSWORD" scp -r "$DEST/yocto.tar" scripts/remote_update/ root@"$SSH":/tmp/
	   if ! $NO_QSPI; then
	      echo "sh /tmp/remote_update/mmc_copy.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH"
	   else
	      echo "sh /tmp/remote_update/mmc_copy.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH" || exit
	   fi
	   if ! $NO_QSPI; then
	       echo "sh /tmp/remote_update/qspi_copy.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH" || exit
	   fi
	   echo "sh /tmp/remote_update/reboot.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH" || exit


	   rm -r "$DEST"
	*/

}
