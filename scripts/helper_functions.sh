#!/bin/bash

VALID_MACHINES="mercury-zx5 microzed-zynq7 zc702-zynq7 picozed-zynq7"

function get_git_commit() {
   (
   cd "$1" || exit
   git rev-parse HEAD
   )
}

function get_current_branch() {
   git rev-parse --abbrev-ref HEAD
}

function is_current_branch() {
   if [ "$(get_current_branch)" != "$1" ]; then
        return 1
    else
        return 0
    fi
}

function clean_repository() {
	(
		git checkout -- .
		git clean -fd
	)
}

function apply_patch() {
	(
		OLD="$(pwd)"
		cd "$1" || exit 1
		git apply "${OLD}/../resources/$2" || exit 1
	)
}

function set_password() {
	if ! grep 'INHERIT += "extrausers"' build/conf/local.conf > /dev/null; then
  		echo 'INHERIT += "extrausers"' >> build/conf/local.conf
	fi

	TEMP_FILE=$(mktemp)
	grep -v 'EXTRA_USERS_PARAMS = \"usermod' build/conf/local.conf > "$TEMP_FILE"
	echo -e "EXTRA_USERS_PARAMS = \"usermod -P $PASSWORD root;\"" >> "$TEMP_FILE"
	mv "$TEMP_FILE" build/conf/local.conf
}

function date_last_commit() {
	YOCTO_WD=$(pwd)
	cd "$1" || exit
	date -d"$(git log -1 --format=%cD)" +%Y%m%d%H%M
	cd "$YOCTO_WD" || exit
}

function install_package() {
	if ! grep "$1" build/conf/local.conf > /dev/null; then
		echo "IMAGE_INSTALL_append = \" $1\"" >> build/conf/local.conf
	fi
}

function clone_do() {
   if [ -z "${RELEASE+x}" ]; then
      echo "RELEASE must be set"
      exit
   fi

	DIR=$(echo "$1" | rev | cut -d '/' -f 1 | rev)
	DIR=$(echo "$DIR" | sed 's/.git//g')

	if [ ! -d "$DIR" ]; then
		git -c http.sslVerify=false clone "$1" || exit $?
	fi
	cd  "$DIR" || exit 1

	[ "$DIR" != "meta-gfa" ] && [ "$DIR" != "meta-enclustra" ] && clean_repository
	git -c http.sslVerify=false  pull || exit $?

	if [ "$2" != "" ]; then
		git checkout "$2"
	else
		if ! is_current_branch $RELEASE; then
			git checkout $RELEASE
		fi
	fi

	cd .. || exit
}

function clone() {
	DIR=$(echo "$1" | rev | cut -d '/' -f 1 | rev)

	if [ "$(pwd | rev | cut -d '/' -f 1 | rev)" != "poky" ] && [ "$DIR" != "poky" ]; then
		cd poky || exit
		clone_do "$1" "$2"
		cd ..
	else 
		clone_do "$1" "$2"
	fi
		
}

function symlink() {
	#Make sure that the symbolic link will point to a valid directory
	rm "$2" 2> /dev/null
	ln -s "$1" "$2" || exit
}

function check_machine() {
	if [[ ! $VALID_MACHINES =~ (^|[[:space:]])"$1"($|[[:space:]]) ]] ; then
	    echo "Invalid machine name. Exiting"
		exit 5
	fi
}

function checkout_machine() {
	check_machine "$1"
	if ! grep "\"$1\"" poky/build/conf/local.conf > /dev/null; then
		sed -i "s/^\(MACHINE ??= \).*/\1\"$1\"/" poky/build/conf/local.conf
	fi
}

function check_layer() {
	TMP=$(mktemp)
	head -n -1 build/conf/bblayers.conf > "$TMP"
	if ! grep "$1" build/conf/bblayers.conf > /dev/null; then
		echo "$(pwd)/$1 \\"  >> "$TMP"
	fi
	echo "\""  >> "$TMP"
	mv "$TMP" build/conf/bblayers.conf
}

function setup_build_dir() {
	if [ ! -d "poky/build" ]; then
    	bash -c "cd poky; source ./oe-init-build-env > /dev/null" || exit 1
	fi

	if [ ! -d "poky/build" ]; then
	    echo "Error when creating poky build directory"
	    exit 1
	fi
}

function rebuild_local_conf() {
	rm poky/build/conf/local.conf || exit 1
    bash -c "cd poky; source ./oe-init-build-env > /dev/null" || exit 1

	if [ ! -d "poky/build" ]; then
	    echo "Error when creating poky build directory"
	    exit 1
	fi
}

function build() {
	source oe-init-build-env
	bitbake core-image-minimal
}