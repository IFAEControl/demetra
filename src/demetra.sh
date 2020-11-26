#!/bin/bash

source scripts/helper_functions.sh

# Default options
MACHINE=microzed-zynq7
SRC=poky/build/tmp/deploy/images
SSH=172.16.12.251
DEST=/tmp/sd/
PASSWORD=root
DEVICE=""
BITSTREAM=~/gfa_fw_sim/petalinux/gfa_uzed7010_sim_2014.4/subsystems/linux/hw-description/design_1_wrapper.bit
DEFAULT_IMAGE=core-image-minimal

source scripts/project_specific/dirs.sh

# In case it doesn't exist, create dummy files
touch scripts/project_specific/layers.sh
touch scripts/project_specific/repositories.sh

REMOTE_BUILD_IP="172.16.17.23"
REMOTE_BUILD_USER="droman"
REMOTE_BUILD_COMMAND="cd gfayocto; git clean -fd; git checkout -- .;git pull;./gfa.sh -db"
EXTRA_ARGS=""
RELEASE=dunfell
YOCTO_LOG_FILE="/tmp/yocto_log.csv" # Set the correct value in ~/.gfayocto_config.env

#source ~/.gfayocto_config.env &>/dev/null

function showHelp() {
cat << EOF
-h, --help          Show this message

YOCTO CONFIGURATION
-m, --machine       Machine name. It only can be mercury-zx5 or microzed-zynq7 (default)
-B, --bitstream     Bitstream location. It should be the full path to the bit file.
-p, --password      Password for the root user.
-D, --dest          Destination directory to copy the output image
-C, --clean         Recursively remove all files from destination before copy the image
-u, --device        If not null, the given device will be automatically mounter/unmouted to destination directory
                    Format can be in the form of /dev/sdX or UUID="x-y" (e.g: 'dev/sdz' or 'UUID="53AC-34FD"')
-H, --hdf           HDF file (will override configured bitstream). If needed it will be forwarded

ACTIONS
-b, --build         Build the machine
-c, --copy          Copy the image
-t, --pack          Pack the image
-S, --ssh-copy      Copy the image remotely (by default it will copy the content to the SD and QSPI)
-T, --test          Run gfa tests

REMOTE UPDATE OPTIONS
--no-qspi           Do not copy the new content to QSPI flash memory

MISC OPTIONS
-d, --docker        Use docker for executing the required action
                    Note that your user should be on the docker group or be root.
-P, --profile       Load the configuration from the given file (a profile)
-v, --verbose       Verbose output (i.e: print current configuration)
-E, --external      Use external source tree.

REMOTE OPTIONS
--args              Append given args when invoking demetra.sh remotely

ADVANCED OPTIONS
-r, --remote        Remote ssh build. Depends on docker
-R, --release       Specify a yocto release. By default sumo is used
-g, --git           Pull all gfa repositories (WARNING: It will discard any changes made to those repositories)
-l, --log           When packing, add new version to yocto log file with the given comment (requires --pack)
-s, --shell         Spawn a shell just before start compiling, useful for example for when needing to configure
                    the linux kernel inside a docker container.
EOF
exit 2
}

function extract_hdf() {
	# This temporary directory will be removed at the end of the script
	TMP_DIR=$(mktemp -d)
	(
	cd $TMP_DIR || exit
	cp "$1" . || exit
	unzip "$1" || exit
	cp ps7_init_gpl.* "$OLDPWD/poky/meta-gfa/recipes-bsp/u-boot/files/" || exit 1
	) || exit 1
	BITSTREAM="$(echo $TMP_DIR/system_top.bit)"
}

function run_tests() {
   (
       cd scripts/gfa_tests || exit
       echo -e "\n==== RUNING TESTS SCRIPT ===="
       ./run_tests.sh "$SSH"
   ) || exit 1
}

function archive() {
    REPO_NAME="$(basename "$1")"
    (
        cd "$1" || exit 1
        git ls-files | zip -@ "$2/${REPO_NAME}.zip" || exit 1
    ) &> /dev/null

    echo -n "${REPO_NAME}.zip"
}

function repo_commit() {
    (
    cd "$1" || exit 1
    if ! stat .git &> /dev/null; then
        echo "ng$(date +%m%d%H%M)"
    elif ! git diff-index --exit-code -s HEAD || [ $(git ls-files -o --exclude-standard | wc -l) -gt 0 ]; then
        echo "u$(date +%m%d%H%M)"
    else
       echo "c$(get_git_commit "$1")"
    fi
    ) || exit 1
}

function cp_repo_to_remote() {
    archive_name="$(archive "$1" "/tmp")"
    scp "/tmp/$archive_name" $REMOTE_BUILD_USER@$REMOTE_BUILD_IP:"$2/$archive_name" || exit
}

function uncompress_remote_repo() {
    REPO_NAME="$(basename "$1")"
    FILE_NAME="${REPO_NAME}.zip"
    echo "unzip -ou $2/$FILE_NAME -d $2/$REPO_NAME; exit" | ssh -tt $REMOTE_BUILD_USER@$REMOTE_BUILD_IP
}

function prepare_remote_repo() {
    cp_repo_to_remote "$1" "$2"
    uncompress_remote_repo "$1" "$2"
}

function build_remote() {
    TMP_DIR=$(mktemp -d)

    echo "mkdir -p $TMP_DIR; exit" | ssh -tt $REMOTE_BUILD_USER@$REMOTE_BUILD_IP

    if [[ ! -z "$HDF" ]]; then
        hdf_name="$(basename "$HDF")"
        scp "$HDF" $REMOTE_BUILD_USER@$REMOTE_BUILD_IP:$TMP_DIR/$hdf_name
        EXTRA_ARGS+=" -H $TMP_DIR/$hdf_name"
    fi

   NAME="$(basename "$profile")"
   cp $profile /tmp/remote_profile || exit 1

   for i in "${!PROJECT_DIRS[@]}"; do
       echo "${i}=$TMP_DIR/$(basename ${PROJECT_DIRS[$i]})" >> /tmp/remote_profile
   done

   scp /tmp/remote_profile $REMOTE_BUILD_USER@$REMOTE_BUILD_IP:$TMP_DIR/profile_$NAME
   
   for i in "${PROJECT_DIRS[@]}"; do
       prepare_remote_repo "$i" "$TMP_DIR"
   done


   echo "$REMOTE_BUILD_COMMAND -P /$TMP_DIR/profile_$NAME $EXTRA_ARGS; exit" | ssh -tt $REMOTE_BUILD_USER@$REMOTE_BUILD_IP   
   echo "rm -rf $TMP_DIR; exit" | ssh -tt $REMOTE_BUILD_USER@$REMOTE_BUILD_IP
}

function build_docker() {
    docker build -f resources/docker/Dockerfile_base -t base-yocto . || exit 1
   # XXX: If docker don't finish correctly we can have unused containers
   CONTAINER_NAME="yocto-$(mktemp -u XXXXX)"
    docker build -f resources/docker/Dockerfile_ssh --build-arg uid="$(id -ru)" -t yocto-build . || exit 1

    DOCKER_MOUNT_ARGS=""
    for i in "${PROJECT_DIRS[@]}"; do
       DOCKER_MOUNT_ARGS+=" -v $i:$i"
   done
    docker run -v "$(pwd)":"$(pwd)" $DOCKER_MOUNT_ARGS \
	       -w "$(pwd)" \
	       --cap-add=NET_ADMIN --device /dev/net/tun:/dev/net/tun \
	       --env-file "$profile" \
	       --env-file ~/.gfayocto_config.env \
               -it --rm --name $CONTAINER_NAME yocto-build \
               ./demetra.sh -b -m "$MACHINE" -p "$PASSWORD" -R "$RELEASE" $($shell && echo "-s") $($external && echo -e "\x2dE") || exit 1
}

function build_local() {
    ./scripts/prepare.sh "$MACHINE" "${PROJECT_DIRS[@]}" $external || exit 1

    (
    cd poky || exit 1
    
    # Change default root password
    set_password "$PASSWORD"

    source ./oe-init-build-env
  
    if $shell; then
       bash
    fi

    bitbake "$DEFAULT_IMAGE" || exit 1
    ) || exit 1
}

function clone_single_repo() {
    (
        BRANCH="$(get_current_branch)"
        mkdir -p "$1" || exit 1
        cd "$1" || exit 1
        stat .git &>/dev/null || git clone -b $BRANCH  "$2" "$1" || exit 1
    )
}

function clone_git_repos() {
#    clone_single_repo "$MODULE_DIR" "git@gitlab.pic.es:DESI-GFA/gfa_module.git" || exit 1
	exit 0
}

function update_git_repos() {
    for i in "${PROJECT_DIRS[@]}"; do
    (
        cd "$i" || exit 1
        git clean -fd . || exit 1
        git checkout -- . || exit 1
        git pull || exit 1
    )
    done
}

function do_copy() {
    check_machine "$MACHINE"
    scripts/copy.sh "$DEST" "$SRC" "$DEVICE" "$MACHINE" "$BITSTREAM" "$clean" ||
    sudo scripts/copy.sh "$DEST" "$SRC" "$DEVICE" "$MACHINE" "$BITSTREAM" "$clean" ||
    su -c "scripts/copy.sh $DEST $SRC $DEVICE $MACHINE $BITSTREAM $clean"
}

function do_sshcopy() {
    check_machine "$MACHINE"
    scripts/ssh-copy.sh "$SRC" "$MACHINE" "$BITSTREAM" "$PASSWORD" "$SSH" "$noqspi" || exit 1
}

function do_pack() {
    local hdf_name line_num tar_name tar_hash

    mkdir -p "$DEST"

    gfayocto=$(repo_commit .)
    {    
        echo "gfayocto $gfayocto";
    } > "$DEST/info_versions"

    wget -nv 'https://docs.google.com/spreadsheets/d/1TKlaFSGHjfkI2bBY5EEMCuuVj24wn13m9bSBYD4G4-s/export?format=csv&gid=0' -O /tmp/vivado.csv || exit 1

    hdf_name="$(basename "$HDF" .hdf)"
    line_num=$(grep ',' /tmp/vivado.csv | cut -d, -f1 | grep -v " " | grep -n "$hdf_name" | cut -d: -f1)
    tar_name="$(date +%Y.%m.%d-%H.%M)-yocto-multi-$line_num"
    tar -cvf "$tar_name.tar" -C "$DEST" .  || exit
    if [[ ! -z "$LOG_COMMENT" ]]; then
        if [[ -z "$HDF" ]]; then
           echo "When using \"log\" option \"hdf\" should also be set"
           exit 1
        fi
        tar_hash=$(sha1sum "$tar_name.tar" | cut -d' ' -f1)
        echo "$tar_name,$hdf_name,$RELEASE,$LOG_COMMENT,$tar_hash" >> "$YOCTO_LOG_FILE"
    fi
}

if [[ $# -eq 0 ]]; then
    showHelp
fi

# Finally check for configuration arguments, which will overwrite the user and default configuration
getopt --test > /dev/null
if [[ $? -ne 4 ]]; then
    echo "I’m sorry, $(getopt --test) failed in this environment."
    exit 1
fi

SHORT=hm:bdB:p:D:cCu:P:tSTvErR:,g,H:,l:,s
LONG=help,machine:,build,docker,bitsream:,password:,dest:,copy,clean,device:,profile:,pack,ssh-copy,test,verbose,external,remote,release:,git,hdf:,log:,shell,no-qspi,args:

# -temporarily store output to be able to check for errors
# -activate advanced mode getopt quoting e.g. via “--options”
# -pass arguments only via   -- "$@"   to separate them correctly
PARSED=$(getopt --options $SHORT --longoptions $LONG --name "$0" -- "$@")
if [[ $? -ne 0 ]]; then
    # e.g. $? == 1
    #  then getopt has complained about wrong arguments to stdout
    exit 2
fi
# use eval with "$PARSED" to properly handle the quoting
eval set -- "$PARSED"

docker=false
copy=false
build=false
clean=false
pack=false
sshcopy=false
tests=false
verbose=false
external=false
remote=false
git=false
shell=false
noqspi=false

while true; do
    case "$1" in
        -h|--help)
            showHelp
            shift
            ;;
        -m|--machine)
            MACHINE=$2
            shift 2
            ;;
        -B|--bitstream)
            BITSTREAM=$2
            shift 2
            ;;
        -p|--password)
            PASSWORD=$2
            shift 2
            ;;
        -D|--dest)
            DEST=$2
            shift 2
            ;;
        -C|--clean)
    	    clean=true
    	    shift
	        ;;
        -u|--device)
            DEVICE=$2
            shift 2
            ;;
        -d|--docker)
            docker=true
            shift
            ;;
        -P|--profile)
            profile=$2
            shift 2
            ;;
        -b|--build)
            build=true
            shift
            ;;
        -c|--copy)
            copy=true
            shift
            ;;
         -S|--ssh-copy)
            sshcopy=true
            shift
            ;;
        -t|--pack)
            pack=true
            # For packing previously we need to copy it
            # but we don't want to mount any device so we set it to ""
            copy=true
            DEVICE=""
            shift
            ;;
         -T|--test)
            tests=true
            shift
            ;;
         -v|--verbose)
            verbose=true
            shift
            ;;
         -E|--external)
            external=true
            shift
            ;;
         -r|--remote)
            remote=true
            shift
            ;;
         -g|--git)
            git=true
            shift
            ;;
         -R|--release)
            RELEASE=$2
            shift 2
            ;;
         -H|--hdf)
            HDF=$2
            shift 2
            ;;
         -l|--log)
            LOG_COMMENT="$2"
            shift 2
            ;;
         -s|--shell)
            shell=true
            shift
            ;;
         --no-qspi)
            noqspi=true
            shift
            ;;
         --args)
            EXTRA_ARGS="$2"
            shift 2
            ;;
         --)
            shift
            break
            ;;
         *)
            echo "Programming error"
            exit 3
            ;;
    esac
done

./scripts/download.sh "$RELEASE" || exit 1

# If a profile is given, then we will use that configuration
if [ "${profile+x}" ]; then
    source "$profile" 
else
    # Dummy profile
    profile="/dev/null"
fi

if $verbose; then
cat <<EOF 
Current values
================================================================================
MACHINE="$MACHINE" 
SRC="$SRC"
SSH="$SSH"
DEST="$DEST"
PASSWORD="$PASSWORD"
DEVICE="$DEVICE"
BITSTREAM="$BITSTREAM"
DEFAULT_IMAGE="$DEFAULT_IMAGE"
MODULE_DIR="$MODULE_DIR"
LIBRARY_DIR="$LIBRARY_DIR"
SERVER_DIR="$SERVER_DIR"
XADC_TEST_DIR="$XADC_TEST_DIR"
MCP_DIR="$MCP_DIR"
RELEASE="$RELEASE"
HDF="$HDF"
EXTRA_ARGS="$EXTRA_ARGS"
================================================================================
EOF
    echo "Press a key to continue"
    read -sn 1

fi

# Copy hdf
if [[ ! -z "$HDF" ]]; then
	extract_hdf "$HDF" || exit 1
fi

if $git; then
    git pull || exit
    clone_git_repos || exit
    update_git_repos || exit
fi

if $remote; then
    build_remote || exit 1
elif $docker && $build; then
	build_docker || exit 1
elif $build; then
    build_local || exit 1
fi

if $copy; then
    do_copy || exit
fi

if $sshcopy; then
    do_sshcopy || exit 
fi

if $tests; then
   run_tests || exit
fi

if $pack; then
   do_pack || exit
fi

if [[ ! -z "$HDF" ]]; then
	rm -r "$TMP_DIR"
fi
