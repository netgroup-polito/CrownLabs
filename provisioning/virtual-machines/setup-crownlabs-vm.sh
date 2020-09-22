#!/bin/bash

# This script provides multiple commands to create a new virtual machine using
# VirtualBox and install the operating system, automatize most of the required
# configuration (using ansible) and finally export the VM for CrownLabs.
# See the README file for more details.

# The following variables refer to the parameters required to push the resulting
# VM to the CrownLabs registry. They define a default for the different variables,
# which can be overridden exporting each variable before executing this script.
# In particular, it is possible to configure the username and password of the
# registry to avoid the interactive prompt.
CROWNLABS_REGISTRY=${CROWNLABS_REGISTRY:-"registry.crownlabs.polito.it"} # The URL of the CrownLabs registry
CROWNLABS_REGISTRY_FOLDER=${CROWNLABS_REGISTRY_FOLDER:-"netgroup"} # Must contain only lowercase letters, numbers, dashes
CROWNLABS_REGISTRY_IMAGE_VERSION=${CROWNLABS_REGISTRY_IMAGE_VERSION:-"$(date '+%Y%m%d')"} # The image tag
# Export these variables or uncomment and configure them directly in this script to avoid the interactive prompt
# CROWNLABS_REGISTRY_USERNAME=
# CROWNLABS_REGISTRY_PASSWORD=

# Configure the Ubuntu distribution selected for desktop installations
# Warning: changing the distribution may break the subsequent configuration
DESKTOP_UBUNTU_DISTRO=xubuntu

# Configure the credentials of the VM user
USERNAME=netlab
PASSWORD=netlab
FULL_USERNAME="$USERNAME (password $PASSWORD)"

# Configure the locale information
LOCALE="en_US"
COUNTRY="IT"
TIME_ZONE="UTC"


##########################################

# The command used to interact with VirtualBox in Linux
VBOXMANAGE_LINUX="VBoxManage"

# The command used to interact with VirtualBox in Linux
# Small adaptations may be necessary in case of a non-standard installation
VBOXMANAGE_WINDOWS="/mnt/c/Program Files/Oracle/VirtualBox/VBoxManage.exe"

# Am I running on plain Linux or on Linux for Windows?
[[ "$(< /proc/version)" == *@(Microsoft|WSL)* ]] \
    && VBOXMANAGE="${VBOXMANAGE_WINDOWS}" \
    || VBOXMANAGE="${VBOXMANAGE_LINUX}"

##########################################

echo "################################"
echo "###### SETUP CROWNLABS VM ######"
echo "################################"
echo

##########################################

EXIT_SUCCESS=0
EXIT_FAILURE=1

# Checks if a command is available or not
function check_available {
    CMDPATH="$1"
    CMDNAME=$(basename "$1")
    if command -v "${CMDPATH}" >/dev/null 2>&1
    then
        echo "* '${CMDNAME}' found!"
    else
        echo "'${CMDNAME}' required but not found. Abort"
        exit ${EXIT_FAILURE}
    fi
}

# Checks if the version of ansible is greater than the required one
function check_ansible_version {
    REQUIRED_VERSION="$1"
    ANSIBLE_VERSION=$(ansible-playbook --version | head --lines 1 | cut --delimiter ' ' --field 2)
    if printf '%s\n%s\n' "${REQUIRED_VERSION}" "${ANSIBLE_VERSION}" | sort --version-sort --check=quiet
    then
        echo "* 'ansible-playbook' Version: ${ANSIBLE_VERSION} - OK!"
    else
        echo "* 'ansible-playbook' Version: ${ANSIBLE_VERSION} - Required: ${REQUIRED_VERSION}. Abort!"
        exit ${EXIT_FAILURE}
    fi
}

function check_docker_privileges {
    docker ps >/dev/null 2>/dev/null || {
        echo "* 'docker': It appears you do not have enough privileges. Warning: do *NOT* run this script with sudo"
        exit ${EXIT_FAILURE}
    }
}

# Verify if all commands required are available. The check for the commands required to export the VM to CrownLabs
# are located in the corresponding section, to avoid introducing undesired dependencies in the other cases.
echo "Checking dependencies..."
check_available "${VBOXMANAGE}"
check_available "ansible-playbook"
check_ansible_version "2.8"
check_available "curl"
check_available "ssh"
check_available "sshpass"
echo

##########################################

# Print the usage message
usage() {
    echo "Usage: $0 <vm-name> [create|configure|configure-nic|export|delete|help]"
    echo "* create [desktop|server] <ubuntu-version> (--install-guest-additions): Create the VM and install the OS"
    echo "* configure <ansible-playbook.yml> (--vbox-only): Configures the VM's OS using ansible"
    echo "* configure-nic [nat|bridged]: Configures the NIC in nat or bridged mode"
    echo "* export [ova|crownlabs]: Exports the VM in OVA format, or pushes it to the CrownLabs registry"
    echo "* delete: Deletes the VM"
    echo "* help: Shows this help"
    exit ${EXIT_SUCCESS}
}

##########################################

# The name of the virtual machine to operate with
VMNAME=$1
[[ "" == "${VMNAME}" ]] && usage

VMNAMEREGEX='^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$'
if [[ ! "${VMNAME}" =~ ${VMNAMEREGEX} ]]
then
    echo "Error: invalid VM name. Valid characters: lowercase letters, numbers, dashes. Abort."
    exit ${EXIT_FAILURE}
fi

echo "Selected Virtual Machine: ${VMNAME}"
echo

##########################################

echo "Checking VirtualBox paths..."

BASEDIR=$(dirname "$0")
VBPATH=$("${VBOXMANAGE}" list systemproperties | sed -n 's/Default machine folder: *//p' | tr -d '\r')
VMPATH="${VBPATH}/${VMNAME}"
HDDPATH="${VMPATH}/${VMNAME}.vdi"

echo "* VirtualBox base path: ${VBPATH}"
echo "* Virtual Machine path: ${VMPATH}"
echo "* HDD path: ${HDDPATH}"
echo

##########################################

COMMAND=$2
case ${COMMAND} in

### Begin Create VM ###
"create")

UBUNTU_DISTRO=$3
UBUNTU_VERSION=$4
GA_FLAG=$5

# Check the correctness of the input parameters
[[ "${UBUNTU_DISTRO}" =~ ^(desktop|server)$ &&
   "${UBUNTU_VERSION}" =~ ^[0-9][0-9]\.[0-9][0-9](\.[0-9])?$ &&
   ( -z ${GA_FLAG} || ${GA_FLAG} != "--install-guest-additions") ]] || {
    echo "Usage: $0 <vm-name> create [desktop|server] <ubuntu-version> (--install-guest-additions)"
    exit ${EXIT_FAILURE};
}

VBOXVERSION=$(${VBOXMANAGE} --version | cut --delimiter '_' --field 1)

DOWNLOAD_PATH="${BASEDIR}/downloads"
mkdir --parents "${DOWNLOAD_PATH}" || \
    { echo "Failed to create '${DOWNLOAD_PATH}'. Abort"; exit ${EXIT_FAILURE}; }

if [[ "${UBUNTU_DISTRO}" == "desktop" ]]
then
    UBUNTU_DISTRO_NAME="${DESKTOP_UBUNTU_DISTRO}"
    UBUNTU_URL_FOLDER="${DESKTOP_UBUNTU_DISTRO}"
    UBUNTU_IMAGE_NAME="${DESKTOP_UBUNTU_DISTRO}-${UBUNTU_VERSION}-desktop-amd64.iso"
else
    # The legacy term refers to the version of the installer, since the new one appears not to support the preseed configuration
    UBUNTU_DISTRO_NAME="ubuntu-server"
    UBUNTU_URL_FOLDER="ubuntu-legacy-server"
    UBUNTU_IMAGE_NAME="ubuntu-${UBUNTU_VERSION}-legacy-server-amd64.iso"
fi

echo "Downloading the ${UBUNTU_DISTRO_NAME} (${UBUNTU_VERSION}) image..."
UBUNTU_IMAGE_URL=https://cdimages.ubuntu.com/${UBUNTU_URL_FOLDER}/releases/${UBUNTU_VERSION}/release/${UBUNTU_IMAGE_NAME}
UBUNTU_SHA256SUMS_URL=https://cdimages.ubuntu.com/${UBUNTU_URL_FOLDER}/releases/${UBUNTU_VERSION}/release/SHA256SUMS
INSTALL_ISO="${DOWNLOAD_PATH}/${UBUNTU_IMAGE_NAME}"
INSTALL_ISO_SHA256SUMS="${UBUNTU_DISTRO_NAME}-${UBUNTU_VERSION}.SHA256SUMS"

# Pre-check whether the URL is valid, since the actual download does not fail in this case (due to the --continue-at flag)
curl --head --silent --fail --output /dev/null "${UBUNTU_IMAGE_URL}" || {
        echo "Failed to download the Ubuntu image from '${UBUNTU_IMAGE_URL}'.";
        echo "Is the ubuntu version correct? Abort";
        exit ${EXIT_FAILURE};
}

curl --continue-at - --progress-bar --output "${INSTALL_ISO}" "${UBUNTU_IMAGE_URL}" || \
    { echo "Failed to download the Ubuntu image from '${UBUNTU_IMAGE_URL}'. Abort"; exit ${EXIT_FAILURE}; }

echo "Verifying the checksum of the ${UBUNTU_DISTRO_NAME} (${UBUNTU_VERSION}) image..."
curl --fail --silent --output "${DOWNLOAD_PATH}/${INSTALL_ISO_SHA256SUMS}" "${UBUNTU_SHA256SUMS_URL}" || \
    { echo "Failed to download the Ubuntu image checksum from '${UBUNTU_SHA256SUMS_URL}'. Abort"; exit ${EXIT_FAILURE}; }

if ( cd "${DOWNLOAD_PATH}"; sha256sum --strict --ignore-missing --status --check "${INSTALL_ISO_SHA256SUMS}"; )
then
    echo "Checksum verification correctly completed";
    rm --force "${DOWNLOAD_PATH}/${INSTALL_ISO_SHA256SUMS}"
else
    echo "Failed to verify the checksum. The downloaded Ubuntu image appears to be corrupted. Abort"
    rm --force "${DOWNLOAD_PATH}/${INSTALL_ISO_SHA256SUMS}"
    exit ${EXIT_FAILURE};
fi

# Install guest additions?
GA_INSTALL=$([[ "--install-guest-additions" == "$GA_FLAG" ]] && echo 1 || echo 0)

if [[ $GA_INSTALL -eq 1 ]]
then
    echo
    echo "Downloading the Guest Additions ISO..."
    GA_BASE_URL=https://download.virtualbox.org/virtualbox/
    GA_URL=${GA_BASE_URL}/${VBOXVERSION}/VBoxGuestAdditions_${VBOXVERSION}.iso
    GA_URL_SHA256SUMS=${GA_BASE_URL}/${VBOXVERSION}/SHA256SUMS
    GA_ISO="${DOWNLOAD_PATH}/VBoxGuestAdditions_${VBOXVERSION}.iso"
    GA_ISO_SHA256SUMS="VBoxGuestAdditions_${VBOXVERSION}.SHA256SUMS"

    # Pre-check whether the URL is valid, since the actual download does not fail in this case (due to the --continue-at flag)
    curl --head --silent --fail --output /dev/null "${GA_URL}" || \
        { echo "Failed to download the Guest Additions image from '${GA_URL}'. Abort"; exit ${EXIT_FAILURE}; }

    curl --continue-at - --progress-bar --output "${GA_ISO}" "${GA_URL}" || \
        { echo "Failed to download the Guest Additions image from '${GA_URL}'. Abort"; exit ${EXIT_FAILURE}; }

    echo "Verifying the checksum of the Guest Additions image..."
    curl --fail --silent --output "${DOWNLOAD_PATH}/${GA_ISO_SHA256SUMS}" "${GA_URL_SHA256SUMS}" || \
        { echo "Failed to download the Guest Additions image checksum from '${GA_URL_SHA256SUMS}'. Abort"; exit ${EXIT_FAILURE}; }

    if ( cd "${DOWNLOAD_PATH}"; sha256sum --strict --ignore-missing --status --check "${GA_ISO_SHA256SUMS}"; )
    then
        echo "Checksum verification correctly completed"
        rm --force "${DOWNLOAD_PATH}/${GA_ISO_SHA256SUMS}"
    else
        echo "Failed to verify the checksum. The downloaded Guest Additions image appears to be corrupted. Abort";
        rm --force "${DOWNLOAD_PATH}/${GA_ISO_SHA256SUMS}"
        exit ${EXIT_FAILURE};
    fi
fi

##########################################

echo
echo "Creating '${VMNAME}' virtual machine..."

# Abort if the VM already exists
if "${VBOXMANAGE}" list vms | grep --quiet "\"${VMNAME}\""
then
    echo "A VM with the same name already exists. Abort."
    exit ${EXIT_FAILURE};
fi

# VirtualBox Machine
VMOSTYPE=Ubuntu_64
"${VBOXMANAGE}" createvm --name "${VMNAME}" --ostype "${VMOSTYPE}" --register || \
    { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

# Check the VBoxManage version, since the clipboard flag changed since version 6.1
printf '%s\n%s\n' "${VBOXVERSION}" "6.1.0" | sort --check=quiet --version-sort \
  && CLIPBOARD_FLAG=clipboard \
  || CLIPBOARD_FLAG=clipboard-mode

# VirtualBox General Settings
"${VBOXMANAGE}" modifyvm "${VMNAME}" --cpus 2 --memory 4096 --vram 64 \
    --ioapic on --audio none --"${CLIPBOARD_FLAG}" bidirectional || \
        { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

# VM Description
"${VBOXMANAGE}" modifyvm "${VMNAME}" --description \
    "$(echo -e "Username: ${USERNAME}\nPassword: ${PASSWORD}")" || \
        { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

# VirtualBox HDD
"${VBOXMANAGE}" createhd --filename "${HDDPATH}" --size 15360 || \
    { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }
"${VBOXMANAGE}" storagectl "${VMNAME}" --name "SATA Controller" --portcount 1 \
    --add sata --controller IntelAHCI --hostiocache on --bootable on || \
        { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }
"${VBOXMANAGE}" storageattach "${VMNAME}" --storagectl "SATA Controller" \
    --type hdd --port 0 --device 0 --medium "${HDDPATH}" || \
        { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

##########################################

SCRIPT_TEMPLATE="${BASEDIR}/unattended-install-scripts/ubuntu_preseed.cfg"
POSTINST_SCRIPT="${BASEDIR}/unattended-install-scripts/ubuntu_postinstall.sh"

GA_INSTALL_FLAGS=("--no-install-additions")
if [[ ${GA_INSTALL} -eq 1 ]]
then
    GA_INSTALL_FLAGS=("--install-additions" "--additions-iso=${GA_ISO}")
fi

# Setup unattended OS installation
"${VBOXMANAGE}" unattended install "${VMNAME}" "${GA_INSTALL_FLAGS[@]}" \
    --iso="${INSTALL_ISO}" --hostname="${VMNAME}.local" \
    --full-user-name="${FULL_USERNAME}" \
    --user=${USERNAME} --password=${PASSWORD} \
    --script-template="${SCRIPT_TEMPLATE}" \
    --post-install-template="${POSTINST_SCRIPT}" \
    --package-selection-adjustment="minimal" \
    --locale="${LOCALE}" --country="${COUNTRY}" --time-zone=${TIME_ZONE} || \
        { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

# Start the Installation
"${VBOXMANAGE}" startvm "${VMNAME}" --type gui || \
    { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

exit ${EXIT_SUCCESS}
;;
### End Create VM ###


##########################################


### Begin Configure VM ###
"configure")

configure_cleanup() {
    # Remove the port forwarding rule
    if [[ "$VMNET" == "nat" ]]
    then
        "${VBOXMANAGE}" controlvm "${VMNAME}" natpf1 delete "SSH" || \
            { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }
    fi

    # Remove the inventory file
    rm --force "${INVENTORY_FILE}"
}

PLAYBOOK_PATH=$3
if [[ ! -f "${PLAYBOOK_PATH}" ]]
then
    echo "Usage: $0 <vm-name> configure <ansible-playbook.yml> (--vbox-only)"
    echo "Error: the Ansible playbook '${PLAYBOOK_PATH}' does not exist"
    exit ${EXIT_FAILURE};
fi

VBOX_ONLY_FLAG=$4
if [[ "--vbox-only" == "$VBOX_ONLY_FLAG" ]]
then
    CROWNLABS_MODE="False"
    ANSIBLE_PLAYBOOK_ARGS=("${@:5}")
else
    CROWNLABS_MODE="True"
    ANSIBLE_PLAYBOOK_ARGS=("${@:4}")
fi

# Abort if the VM does not exists
if ! "${VBOXMANAGE}" list vms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' does not exist. Abort."
    exit ${EXIT_FAILURE};
fi

# Abort if the VM is not running
if ! "${VBOXMANAGE}" list runningvms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' is currently not running, please start it. Abort."
    exit ${EXIT_FAILURE};
fi

# Get the mode associated with the network interface in VirtualBox
VMNETSTR=$("${VBOXMANAGE}" showvminfo "${VMNAME}" | sed -n 's/NIC 1: *//p' | tr -d '\r')
if echo "${VMNETSTR}" | grep --ignore-case --quiet nat
then
    VMNET=nat
elif echo "${VMNETSTR}" | grep --ignore-case --quiet bridged
then
    VMNET=bridged
else
    echo "Failed to get the VM network interface mode. Abort."
    exit ${EXIT_FAILURE};
fi

# Get the IP of the VM
VMIP=$("${VBOXMANAGE}" guestproperty get "${VMNAME}" "/VirtualBox/GuestInfo/Net/0/V4/IP" \
            | cut --delimiter=' ' --field 2)
if [[ -z "$VMIP" ]]
then
    echo "Failed to get the IP assigned to the '${VMNAME}' VM. Abort."
    exit ${EXIT_FAILURE};
fi

SSHIP=$VMIP
SSHPORT=22

# Ensure the port forwarding is removed even in case the execution is interrupted
trap configure_cleanup 0

# Add port forwording to allow SSH access
if [[ "$VMNET" == "nat" ]]
then
    SSHIP=127.0.0.1
    SSHPORT=2222
    "${VBOXMANAGE}" controlvm "${VMNAME}" natpf1 "SSH,tcp,$SSHIP,$SSHPORT,$VMIP,22" || \
        { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }
fi

# Create the inventory file
INVENTORY_FILE="${BASEDIR}/${VMNAME}-inventory.yml"
cat <<EOF > "${INVENTORY_FILE}"
---
all:
  hosts:
    "${VMNAME}":
      ansible_host: $SSHIP
      ansible_port: $SSHPORT
      ansible_user: $USERNAME
      ansible_ssh_pass: $PASSWORD
      ansible_become_pass: $PASSWORD
      ansible_ssh_extra_args: '-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'
      ansible_python_interpreter: auto
      crownlabs_mode: ${CROWNLABS_MODE}
EOF

echo "Configuring VM with Ansible playbook '${PLAYBOOK_PATH}' (crownlabs-mode: ${CROWNLABS_MODE})"
ansible-playbook --inventory "${INVENTORY_FILE}" "${PLAYBOOK_PATH}" "${ANSIBLE_PLAYBOOK_ARGS[@]}"

exit ${EXIT_SUCCESS}
;;
### End Configure VM ###


##########################################


### Begin Configure NIC ###
"configure-nic")

# Abort if the VM does not exists
if ! "${VBOXMANAGE}" list vms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' does not exist. Abort."
    exit ${EXIT_FAILURE};
fi

ISVMRUNNING=$("${VBOXMANAGE}" list runningvms | grep "\"${VMNAME}\"")

VMNIC=$3
case ${VMNIC} in
"nat")
    if [[ ${ISVMRUNNING} ]]
    then
        "${VBOXMANAGE}" controlvm "${VMNAME}" nic1 "${VMNIC}" || \
            { echo "Failed to configure the NIC mode. Abort"; exit ${EXIT_FAILURE}; }
    else
        "${VBOXMANAGE}" modifyvm "${VMNAME}" --nic1 "${VMNIC}" || \
            { echo "Failed to configure the NIC mode. Abort"; exit ${EXIT_FAILURE}; }
    fi

;;
"bridged")
    VMBRAD=$("${VBOXMANAGE}" list bridgedifs | sed -n 's/Name: *//p' | tr -d '\r' | head -n 1)
    if [[ -z "${VMBRAD}" ]]
    then
        echo "Failed to determine a possible bridge adapter. Abort."
        exit ${EXIT_FAILURE};
    fi

    if [[ ${ISVMRUNNING} ]]
    then
        "${VBOXMANAGE}" controlvm "${VMNAME}" nic1 "${VMNIC}" "${VMBRAD}" || \
            { echo "Failed to configure the NIC mode. Abort"; exit ${EXIT_FAILURE}; }
    else
        "${VBOXMANAGE}" modifyvm "${VMNAME}" --nic1 "${VMNIC}" --bridgeadapter1 "${VMBRAD}" || \
            { echo "Failed to configure the NIC mode. Abort"; exit ${EXIT_FAILURE}; }
    fi
;;

*)
echo "Usage: $0 <vm-name> configure-nic [nat|bridged]"
exit ${EXIT_FAILURE};
;;
esac

exit ${EXIT_SUCCESS}
;;
### End Configure NIC ###


##########################################


### Begin Export VM ###
"export")

# Abort if the VM does not exists
if ! "${VBOXMANAGE}" list vms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' does not exist. Abort."
    exit ${EXIT_FAILURE};
fi

# Abort if the VM is running
if "${VBOXMANAGE}" list runningvms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' is currently running, please stop it. Abort."
    exit ${EXIT_FAILURE};
fi

# Abort if the HDD does not exist
if ! [[ -f "${HDDPATH}" ]]
then
    echo "The HDD '${HDDPATH}' does not exist or has been moved from the expected location. Abort."
    exit ${EXIT_FAILURE};
fi

EXPORT_MODE=$3
case ${EXPORT_MODE} in

# Export the VM into the OVA format
"ova")

# Set boot order
"${VBOXMANAGE}" modifyvm "${VMNAME}" --boot1 disk --boot2 none --boot3 none --boot4 none || \
    { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

# Compact the HDD
echo "Compressing HDD..."
"${VBOXMANAGE}" modifymedium disk "${HDDPATH}" --compact || \
    { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

# Export the Virtual Machine into ova format
EXPORT_BASE_PATH="${BASEDIR}/export"
EXPORT_PATH="${EXPORT_BASE_PATH}/${VMNAME}-$(date "+%Y%m%d").ova"

echo
echo "Exporting '${VMNAME}' VM to '${EXPORT_PATH}'..."

mkdir --parents "${EXPORT_BASE_PATH}" || \
    { echo "Failed to create '${EXPORT_BASE_PATH}'. Abort"; exit ${EXIT_FAILURE}; }
"${VBOXMANAGE}" export "${VMNAME}" --output "${EXPORT_PATH}" || \
    { echo "VBoxManage command failed. Abort"; exit ${EXIT_FAILURE}; }

exit ${EXIT_SUCCESS}
;;

# Export the Virtual Machine to the CrownLabs registry
"crownlabs")

export_crownlabs_cleanup() {
    echo
    echo "Cleaning up..."

    # Remove the exported HDD
    echo "* Removing the exported HDD"
    [[ -z ${EXPHDDPATH} ]] || rm -f "${EXPHDDPATH}"

    # Remove the docker image
    echo "* Removing the docker image"
    [[ -z ${IMAGE_TAG} ]] || docker image rm "${IMAGE_TAG}" >/dev/null 2>&1

    # Logout from the repository
    echo "* Logging out from the crownlabs registry"
    docker logout "${CROWNLABS_REGISTRY}" >/dev/null 2>&1
}

# Trigger the cleanup function before exiting
trap export_crownlabs_cleanup 0

# Check for the additional dependencies required to export the VM to CrownLabs
echo "Checking additional dependencies..."
check_available "docker"
check_docker_privileges
check_available "virt-sparsify"

# Check the correctness of the registry folder name
CROWNLABS_REGISTRY_FOLDER_REGEX='^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$'
if [[ ! "${CROWNLABS_REGISTRY_FOLDER}" =~ ${CROWNLABS_REGISTRY_FOLDER_REGEX} ]]
then
    echo "Error: invalid registry folder. Valid characters: lowercase letters, numbers, dashes. Abort."
    exit ${EXIT_FAILURE}
fi

# Check for the readability of the executable containing the Linux kernel (required by virt-sparsify)
KERNEL_IMAGE=$(find /boot -maxdepth 1 -iname 'vmlinuz-*' | sort | tail -n 1)
if [[ ! -r "${KERNEL_IMAGE}" ]]
then
    echo "Unfortunately it seems you strumbled into this Ubuntu \"bug\" [https://bugs.launchpad.net/ubuntu/+source/linux/+bug/759725]"
    echo "Please run 'sudo dpkg-statoverride --add --update root root 0644 ${KERNEL_IMAGE}' and then rerun this script."
    exit ${EXIT_FAILURE}
fi

# Login to the docker registry
echo
echo "Logging in to the crownlabs registry"
USERNAME_ARG=$([[ -z "${CROWNLABS_REGISTRY_USERNAME}" ]] && echo "" || echo "--username ${CROWNLABS_REGISTRY_USERNAME}")
PASSWORD_ARG=$([[ -z "${CROWNLABS_REGISTRY_PASSWORD}" ]] && echo "" || echo "--password ${CROWNLABS_REGISTRY_PASSWORD}")

# shellcheck disable=SC2086
# USERNAME_ARG and PASSWORD_ARG need to be unquoted to allow for empty values
docker login "${CROWNLABS_REGISTRY}" ${USERNAME_ARG} ${PASSWORD_ARG} 2>/dev/null ||
    { echo "Login Failed. Abort"; exit ${EXIT_FAILURE}; }

# Create a temporary folder
echo
echo "Creating a temporary folder..."
BUILDDIR=$(mktemp -d) || \
    { echo "Failed to create a temporary folder. Abort"; exit ${EXIT_FAILURE}; }

DOCKERFILEPATH="${BUILDDIR}/Dockerfile"
EXPHDDPATH="${BUILDDIR}/${VMNAME}.qcow2"
IMAGE_TAG="${CROWNLABS_REGISTRY}/${CROWNLABS_REGISTRY_FOLDER}/${VMNAME}:${CROWNLABS_REGISTRY_IMAGE_VERSION}"

# Create the Dockerfile
cat <<EOF > "${DOCKERFILEPATH}"
FROM scratch
ADD $(basename "${EXPHDDPATH}") /disk/
EOF

# Export the VM's HDD to the qcow2 format
echo
echo "Compressing and exporting the HDD image to '${EXPHDDPATH}'..."
virt-sparsify --format vdi "${HDDPATH}" --convert qcow2 "${EXPHDDPATH}" --compress || \
    { echo "Failed to convert the VM's HDD. Abort"; exit ${EXIT_FAILURE}; }

# Build the Docker image
echo
echo "Building the Docker image..."
docker build --tag "${IMAGE_TAG}" "${BUILDDIR}" || \
    { echo "Failed to build the Docker image. Abort"; exit ${EXIT_FAILURE}; }

echo
echo "Pushing the Docker image to ${IMAGE_TAG}"
docker push "${IMAGE_TAG}" || \
    { echo "Failed to push the Docker image. Abort"; exit ${EXIT_FAILURE}; }

;;

*)
echo "Usage: $0 <vm-name> export [ova|crownlabs]"
exit ${EXIT_FAILURE};
;;
esac

exit ${EXIT_SUCCESS}
;;
### End Export VM ###


##########################################


### Begin Delete VM ###
"delete")

# Abort if the VM does not exists
if ! "${VBOXMANAGE}" list vms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' does not exist. Abort."
    exit ${EXIT_FAILURE};
fi

# Abort if the VM is running
if "${VBOXMANAGE}" list runningvms | grep --quiet "\"${VMNAME}\""
then
    echo "The VM '${VMNAME}' is currently running, please stop it. Abort."
    exit ${EXIT_FAILURE};
fi

echo "Deleting '${VMNAME}'"
"${VBOXMANAGE}" unregistervm --delete "${VMNAME}"

exit ${EXIT_SUCCESS}
;;
### End Delete VM ###


##########################################


*)
usage
;;

esac
