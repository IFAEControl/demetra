#!/bin/bash

VALID_MACHINES="mercury-zx5 microzed-zynq7 zc702-zynq7 picozed-zynq7 pz7030-fmc2"

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

function set_password() {
	if ! grep 'INHERIT += "extrausers"' build/conf/local.conf > /dev/null; then
  		echo 'INHERIT += "extrausers"' >> build/conf/local.conf
	fi

	TEMP_FILE=$(mktemp)
	grep -v 'EXTRA_USERS_PARAMS = \"usermod' build/conf/local.conf > "$TEMP_FILE"
	echo -e "EXTRA_USERS_PARAMS = \"usermod -P $1 root;\"" >> "$TEMP_FILE"
	mv "$TEMP_FILE" build/conf/local.conf
}

function date_last_commit() {
	YOCTO_WD=$(pwd)
	cd "$1" || exit
	date -d"$(git log -1 --format=%cD)" +%Y%m%d%H%M
	cd "$YOCTO_WD" || exit
}

function checkout_repository() {
	if ! is_current_branch $1; then
		git checkout $1
	fi
}

function check_machine() {
	if [[ ! $VALID_MACHINES =~ (^|[[:space:]])"$1"($|[[:space:]]) ]] ; then
	    echo "Invalid machine name. Exiting"
		exit 5
	fi
}

function checkout_machine() {
	check_machine "$1"
	if ! grep "\"$1\"" build/conf/local.conf > /dev/null; then
		sed -i "s/^\(MACHINE ??= \).*/\1\"$1\"/" build/conf/local.conf
	fi
}

function build() {
	source oe-init-build-env
	bitbake ifae-image
}

function dockerized_run() {
    docker build -f resources/docker/Dockerfile_base -t base-yocto . || exit 1
   # XXX: If docker don't finish correctly we can have unused containers
   CONTAINER_NAME="yocto-$(mktemp -u XXXXX)"
    docker build -f resources/docker/Dockerfile_ssh --build-arg user="$(whoami)" --build-arg uid="$(id -ru)" -t yocto-build . || exit 1

    DOCKER_MOUNT_ARGS=""
    for i in "${PROJECT_DIRS[@]}"; do
       DOCKER_MOUNT_ARGS+=" -v $i:$i"
   done

   # TODO mount external source dirs
    docker run -v "$(pwd)":"$(pwd)" \
    		-v /home/droman/Develop/Work/linda/linda_module:/home/droman/Develop/Work/linda/linda_module \
    		-v /home/droman/Develop/Work/linda/linda_library:/home/droman/Develop/Work/linda/linda_library \
	       -w "$(pwd)" \
	       --cap-add=NET_ADMIN --network=host --device /dev/net/tun:/dev/net/tun \
           -it --rm --name $CONTAINER_NAME yocto-build \
           ./demetra $@ || exit 1
}
