#!/bin/bash
#Script to convert qemu-images with preallocation
#Script functions
function script_help () {
  echo "
      Usage: $(basename $0) [options] original-file new-file

          -i   Ineractive mode

          -h   this help text

          original-file  File to convert

          new-file       File to create

      Example:
        $(basename $0) file.raw new-file.qcow2"

  exit ${1:-0}
}

function interactive_convert_file () {

  echo "Original file"
  read originalFile

  if [[ ! -f $originalFile ]]; then
    echo "File $originalFile not found"
    exit 1
  fi

  echo "File to convert to"
  read newFile

  if [[ -e $newFile ]]; then
    echo "File already exists!"
    exit 1
  fi


  qemu-img convert -f qcow2 -O qcow2 -o preallocation=metadata $originalFile $newFile

  exit ${1:-0}

}

function argument_convert_file () {

  if [[ ! -f $origFile ]]; then
    echo "File $origFile not found"
    exit 1
  fi

  if [[ -e $newFile ]]; then
    echo "File $newFile already exists!"
    exit 1
  fi

  qemu-img convert -f qcow2 -O qcow2 -o preallocation=metadata $origFile $newFile

  exit ${1:-0}

}

#Show help if no arguments or options are passed
[[ ! "$*" ]] && script_help 1
OPTIND=1


#Read command line options
while getopts "ih" opt; do
    case "$opt" in
      i) interactive_convert_file ;;
      h) script_help ;;
      \?) script_help 1 ;;
    esac
done
shift $(($OPTIND-1));

#Run argument function
origFile=$1
newFile=$2
argument_convert_file
