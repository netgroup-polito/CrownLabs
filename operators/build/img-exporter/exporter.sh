#!/bin/sh
IMG_DIR=/data
OUT_DIR=/img-tmp
IMG_NAME=disk.img
OUT_IMAGE=vm-snapshot.qcow2
PROG_NAME=$0

usage(){
	echo "Usage: $PROG_NAME [-options]"
	echo "  -d, --img-dir        Specify the working directory [DEFAULT=$IMG_DIR]"
	echo "  -o, --out-dir        Specify the output directory  [DEFAULT=$OUT_DIR]"
	echo "  -n, --img-name       Specify the name of the image [DEFAULT=$OUT_IMAGE]"
	exit 1
}

parse_args(){
	while [ "${1:-}" != "" ]; do
		case "$1" in
			"-d" | "--img-dir")
				shift
				IMG_DIR=$1
				;;
			"-o" | "--out-dir")
				shift
				OUT_DIR=$1
				;;
			"-n" | "--img-name")
				shift
				IMG_NAME=$1
				;;
			*)
				usage
				;;
		esac
		shift
	done
}

export_img(){
	echo "Converting the image..."

	# Check if output directory exists, if not create it
	# and try with the conversion of the image.
	mkdir -p "$OUT_DIR"
	qemu-img convert -c -f raw -O qcow2  "${IMG_DIR}/${IMG_NAME}" "${OUT_DIR}/${OUT_IMAGE}"

	echo "Creating Dockerfile..."
	# Create the Dockerfile.
	cat <<EOF > "${OUT_DIR}/Dockerfile"
FROM scratch
ADD ${OUT_IMAGE} /disk/
EOF
}

parse_args "$@"

if export_img;
then
	echo "${IMG_DIR}/${IMG_NAME} successully converted"
else
	echo "Conversion unsuccessfully completed"
	exit 1
fi