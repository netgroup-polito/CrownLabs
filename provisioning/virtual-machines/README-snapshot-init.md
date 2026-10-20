# Prepare a VM before creating an image

Before creating an image from a CrownLabs virtual machine, you must reset
`cloud-init`. This prepares the VM to be reused as the starting point for new
virtual machines.

If this step is skipped, VMs created from the image may keep configuration from
the original VM or may not be configured correctly on their first startup.

## Before you start

Make sure that:

- All required software has been installed.
- All manual configuration has been completed.
- VM works as expected.
- Personal files, passwords, tokens and other sensitive information have been
  removed.

The image will contain the files currently stored in the VM. Only include data
that may safely be shared with users of the image.

## 1. Open a terminal inside the VM

Start the VM and connect to it from CrownLabs. Open a terminal inside the VM.

Wait until the VM has completed its startup, then run:

```bash
sudo cloud-init status --wait
```

This command waits for the current VM configuration to finish before the reset
is performed.

## 2. Reset cloud-init

Run:

```bash
sudo cloud-init clean --logs --machine-id
```

The command prepares the VM so that the next machine created from the image is
configured as a new machine.

If the command fails or `cloud-init` is not available, do not create the image.
Contact a CrownLabs administrator or the person responsible for the VM.

## 3. Shut down the VM

After the reset completes successfully, immediately shut down the VM:

```bash
sudo shutdown -h now
```

Do **not** restart the VM before creating the image. Restarting it runs
`cloud-init` again and cancels the preparation you have just completed.

If you restart the VM by mistake, repeat the procedure from the beginning.

Wait until CrownLabs shows the VM as powered off.

## 4. Return to the image creation modal

Close the connection to the VM and return to the **Create New Image** modal from
which you opened this guide through the **help!** link.

Once CrownLabs shows the VM as powered off, confirm that you have completed the
preparation steps and select **Create**. Keep the VM powered off while the image
is being created.
