
variable "ANSIBLE_PLAYBOOK" {
  type    = string
  default = "${env("ANSIBLE_PLAYBOOK")}"
}

variable "DISK_SIZE" {
  type    = string
  default = "${env("DISK_SIZE")}"
}

variable "INSTALL_DESKTOP_ENVIRONMENT" {
  type    = string
  default = "${env("INSTALL_DESKTOP_ENVIRONMENT")}"
}

variable "ISO_CHECKSUM" {
  type    = string
  default = "${env("ISO_CHECKSUM")}"
}

variable "ISO_URL" {
  type    = string
  default = "${env("ISO_URL")}"
}

variable "MEMORY" {
  type    = string
  default = "${env("MEMORY")}"
}

source "qemu" "crownlabs" {
  accelerator                  = "kvm"
  boot_wait                    = "30s"
  cd_files                     = ["./meta-data", "./user-data"]
  cd_label                     = "cidata"
  disk_image                   = true
  disk_interface               = "virtio"
  disk_size                    = "${var.DISK_SIZE}"
  format                       = "raw"
  headless                     = true
  iso_checksum                 = "${var.ISO_CHECKSUM}"
  iso_url                      = "${var.ISO_URL}"
  memory                       = "${var.MEMORY}"
  net_device                   = "virtio-net"
  output_directory             = "."
  shutdown_command             = ""
  ssh_disable_agent_forwarding = true
  ssh_timeout                  = "30m"
  ssh_username                 = "ubuntu"
  ssh_password                 = "crownlabs"
  vm_name                      = "crownlabs.img"
}

build {
  sources = ["source.qemu.crownlabs"]

  provisioner "ansible" {
    extra_arguments = [
      "--become",
      "--extra-vars", 
      "ansible_sudo_pass=crownlabs ansible_ssh_pass=crownlabs crownlabs_mode=true install_desktop_environment=${var.INSTALL_DESKTOP_ENVIRONMENT}"
    ]
    user = "ubuntu"
    use_proxy = false
    playbook_file   = "./ansible/playbooks/${var.ANSIBLE_PLAYBOOK}"
  }

  provisioner "shell" {
    inline = ["sudo cloud-init clean"]
  }

  post-processor "shell-local" {
    inline = ["virt-sparsify crownlabs.img --convert qcow2 --compress crownlabs.qcow2 && rm crownlabs.img"]
  }
}
