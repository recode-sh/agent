#!/bin/bash
# Recode instance init
set -euo pipefail

log () {
  echo -e "${1}" >&2
}

# Remove "debconf: unable to initialize frontend: Dialog" warnings
echo 'debconf debconf/frontend select Noninteractive' | sudo tee debconf-set-selections > /dev/null

handleExit () {
  EXIT_CODE=$?
  exit "${EXIT_CODE}"
}

trap "handleExit" EXIT

# -- Set hostname

sudo hostnamectl set-hostname "${DEV_ENV_NAME_SLUG}"

# -- System dependencies

log "Installing system and Docker dependencies"

sudo apt-get --assume-yes --quiet --quiet install wget git vim apt-transport-https ca-certificates gnupg lsb-release

# -- Recode volume configuration

# log "Locating Recode volume"

# RECODE_VOL_MOUNTPOINT="/recode"
# RECODE_VOL_LABEL="recode-volume"
# DOCKER_DATA_DIR="${RECODE_VOL_MOUNTPOINT}/docker"

# # Find the volume not mounted (mountpoint == null) 
# # and with no partition (children == null)
# RECODE_VOL=$(lsblk --json | jq '.blockdevices[] | select(.mountpoint == null and .children == null) | .name' --raw-output)

# if [[ "${RECODE_VOL}" = "" ]]; then
#   echo "Recode volume not found"
#   exit 1
# elif [[ "${RECODE_VOL}" = *" "* ]]; then # eg: sda1 sda2
#   echo "Multiple volumes match the Recode one"
#   exit 1
# fi

# RECODE_VOL="/dev/${RECODE_VOL}"

# log "Recode volume found ${RECODE_VOL}"

# # If the output shows simply data, there is no file system on the device
# # Example:
# # [ec2-user ~]$ file --special-files /dev/xvdf
# # /dev/xvdf: data
# RECODE_VOL_IS_FORMATTED=$([[ $(sudo file --special-files "${RECODE_VOL}") = "${RECODE_VOL}: data" ]] && echo "false" || echo "true")

# if [[ "${RECODE_VOL_IS_FORMATTED}" = "false" ]]; then
#   log "Recode volume not formatted. Formatting now..."
#   sudo mkfs.ext4 "${RECODE_VOL}"
# fi

# log "Backuping /etc/fstab to /etc/fstab.orig"
# sudo cp /etc/fstab /etc/fstab.orig

# log "Labeling Recode volume to ${RECODE_VOL_LABEL}"
# sudo e2label "${RECODE_VOL}" "${RECODE_VOL_LABEL}"

# log "Creating ${RECODE_VOL_MOUNTPOINT} mountpoint"
# sudo mkdir --parents "${RECODE_VOL_MOUNTPOINT}"

# log "Adding Recode volume to fstab"
# echo "LABEL=${RECODE_VOL_LABEL}  ${RECODE_VOL_MOUNTPOINT}  ext4  defaults,nofail  0  2" | sudo tee --append /etc/fstab > /dev/null

# log "Mounting all devices"
# sudo mount --all

# if [[ "${RECODE_VOL_IS_FORMATTED}" = "true" ]]; then
#   log "Recode volume already formatted before mounting. Making sure the filesystem size match the attached volume..."
#   sudo resize2fs "${RECODE_VOL}"
# fi

# sudo mkdir --parents "${DOCKER_DATA_DIR}"
# sudo chown --recursive recode:recode "${RECODE_VOL_MOUNTPOINT}"

# -- Packages configuration

# Docker
log "Installing Docker"

if [[ ! -f "/usr/share/keyrings/docker-archive-keyring.gpg" ]]; then
  curl --fail --silent --show-error --location https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor --output /usr/share/keyrings/docker-archive-keyring.gpg
fi

if [[ ! -f "/etc/apt/sources.list.d/docker.list" ]]; then
	echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release --codename --short) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
fi

sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet remove docker docker-engine docker.io containerd runc
sudo apt-get --assume-yes --quiet --quiet install docker-ce docker-ce-cli containerd.io

# log "Configuring Docker"

# sudo mkdir /etc/systemd/system/docker.service.d
# sudo touch /etc/systemd/system/docker.service.d/override.conf

# sudo tee /etc/systemd/system/docker.service.d/override.conf > /dev/null << EOF
# [Service]
# ExecStart=
# ExecStart=/usr/bin/dockerd --host fd:// --containerd=/run/containerd/containerd.sock --data-root="${DOCKER_DATA_DIR}"
# EOF

# sudo systemctl daemon-reload
# sudo systemctl restart docker.service

# -- Run as "recode"

log "Configuring workspace for user \"recode\""

sudo --set-home --login --user recode -- env \
	GITHUB_USER_EMAIL="${GITHUB_USER_EMAIL}" \
	USER_FULL_NAME="${USER_FULL_NAME}" \
bash << 'EOF'

mkdir --parents .vscode-server

if [[ ! -f ".ssh/recode_github" ]]; then
	ssh-keygen -t ed25519 -C "${GITHUB_USER_EMAIL}" -f .ssh/recode_github -q -N ""
fi

chmod 644 .ssh/recode_github.pub
chmod 600 .ssh/recode_github

if ! grep --silent --fixed-strings "IdentityFile ~/.ssh/recode_github" .ssh/config; then
	rm --force .ssh/config
  echo "Host github.com" >> .ssh/config
	echo "  User git" >> .ssh/config
	echo "  Hostname github.com" >> .ssh/config
	echo "  PreferredAuthentications publickey" >> .ssh/config
	echo "  IdentityFile ~/.ssh/recode_github" >> .ssh/config
fi

chmod 600 .ssh/config

if ! grep --silent --fixed-strings "github.com" .ssh/known_hosts; then
  ssh-keyscan github.com >> .ssh/known_hosts
fi

GIT_GPG_KEY_COUNT="$(gpg --list-signatures --with-colons | grep 'sig' | grep "${GITHUB_USER_EMAIL}" | wc -l)"

if [[ $GIT_GPG_KEY_COUNT -eq 0 ]]; then
	gpg --quiet --batch --gen-key << EOF2
%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: ${USER_FULL_NAME}
Name-Email: ${GITHUB_USER_EMAIL}
Expire-Date: 0
EOF2
fi

GIT_GPG_KEY_ID="$(gpg --list-signatures --with-colons | grep 'sig' | grep "${GITHUB_USER_EMAIL}" | head --lines 1 | cut --delimiter ':' --fields 5)"

if [[ ! -f ".gnupg/recode_github_gpg_public.pgp" ]]; then
	GIT_GPG_PUBLIC_KEY="$(gpg --armor --export "${GIT_GPG_KEY_ID}")"

	echo "${GIT_GPG_PUBLIC_KEY}" >> .gnupg/recode_github_gpg_public.pgp
fi

chmod 644 .gnupg/recode_github_gpg_public.pgp

if [[ ! -f ".gnupg/recode_github_gpg_private.pgp" ]]; then
	GIT_GPG_PRIVATE_KEY="$(gpg --armor --export-secret-keys "${GIT_GPG_KEY_ID}")"

	echo "${GIT_GPG_PRIVATE_KEY}" >> .gnupg/recode_github_gpg_private.pgp
fi

chmod 600 .gnupg/recode_github_gpg_private.pgp

git config --global user.name "${USER_FULL_NAME}"
git config --global user.email "${GITHUB_USER_EMAIL}"

git config --global user.signingkey "${GIT_GPG_KEY_ID}"
git config --global commit.gpgsign true

EOF
