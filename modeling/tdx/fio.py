import os
import sys
import logging
import time
from optparse import OptionParser

from codecarbon import OfflineEmissionsTracker

from pycloudstack.vmguest import VMGuestFactory
from pycloudstack.vmparam import VM_TYPE_TD, VM_TYPE_EFI
from pycloudstack.cmdrunner import NativeCmdRunner

LOG = logging.getLogger(__name__)
CURR_DIR = os.path.realpath(os.path.dirname(__file__))
DEFAULT_IMAGE = os.path.join(CURR_DIR, "td-guest-ubuntu-22.04-test.qcow2")
DEFAULT_KERNEL = os.path.join(CURR_DIR, "vmlinuz-jammy")
DEFAULT_SSHKEY = os.path.join(CURR_DIR, "vm_ssh_test_key")

def parse_args():
    parser = OptionParser()
    parser.add_option("-i", "--image", dest="image", default=DEFAULT_IMAGE,
                    help="VM guest qcow2 image", metavar="FILE")
    parser.add_option("-k", "--kernel", dest="kernel", default=DEFAULT_KERNEL,
                    help="VM guest kernel image", metavar="FILE")
    parser.add_option("-s", "--sshkey", dest="sshkey", default=DEFAULT_SSHKEY,
                    help="VM SSH test key", metavar="FILE")
    parser.add_option("-q", "--quiet",
                    action="store_false", dest="verbose", default=True,
                    help="don't print status messages to stdout")

    (options, args) = parser.parse_args()
    if options.verbose:
        logging.basicConfig(level=logging.DEBUG)
    else:
        logging.basicConfig(level=logging.INFO)

    if not os.path.exists(options.image):
        LOG.error("Could not find the VM guest image %s" % options.image)
        sys.exit(1)
    if not os.path.exists(options.kernel):
        LOG.error("Could not find the VM guest kernel %s" % options.kernel)
        sys.exit(1)
    if not os.path.exists(options.sshkey):
        LOG.error("Could not find the VM SSH key %s" % options.sshkey)
        sys.exit(1)

    return (options.image, options.kernel, options.sshkey)

def run_fio_workload_on_baremetal(name, duration, output="carbon.csv", fio_enable=True):
    tracker =  OfflineEmissionsTracker(
        project_name=name,
        country_iso_code="USA",
        output_dir=CURR_DIR,
        output_file=output,
    )
    LOG.info(f"run_fio_workload_on_baremetal {name}")
    tracker.start()
    if fio_enable:
        command_list  = [
            'fio --name=randwrite --ioengine=libaio --iodepth=1 --rw=randwrite \
                --bs=64k --direct=0 --size=1024M --numjobs=4 --runtime=%d \
                --time_based --group_reporting' % duration,
        ]
        for cmd in command_list:
            runner = NativeCmdRunner(cmd.split())
            runner.run()
    else:
        time.sleep(duration)
    emissions: float = tracker.stop()
    LOG.info(f"Emissions: {emissions} kg")


def run_fio_workload_in_vm(name, vmtype, image, kernel, sshkey, duration, \
    output="carbon.csv", fio_enable=True):
    tracker =  OfflineEmissionsTracker(
        project_name=name,
        country_iso_code="USA",
        output_dir=CURR_DIR,
        output_file=output,
    )

    vm_factory = VMGuestFactory(image, kernel)

    tracker.start()
    inst = vm_factory.new_vm(vmtype, auto_start=True)
    if not inst.wait_for_ssh_ready():
        LOG.error("Fail to start the VM")
    LOG.info("Successful start the VM")

    if fio_enable:
        command_list  = [
            'fio --name=randwrite --ioengine=libaio --iodepth=1 --rw=randwrite \
                --bs=64k --direct=0 --size=1024M --numjobs=4 --runtime=%d \
                --time_based --group_reporting' % duration,
        ]
        for cmd in command_list:
            runner = inst.ssh_run(cmd.split(), sshkey)
    else:
        time.sleep(duration)
    emissions: float = tracker.stop()
    LOG.info(f"Emissions: {emissions} kg")


if __name__=="__main__":
    image, kernel, sshkey = parse_args()
    for duration in [60, 120, 180]:
        run_fio_workload_on_baremetal("bare-empty-%d" % duration, duration, \
            fio_enable=False)
        run_fio_workload_on_baremetal("bare-fio-%d" % duration, duration)
    for duration in [60, 120, 180]:
        run_fio_workload_in_vm("tdvm-empty-%d" % duration, VM_TYPE_TD, image, kernel,
            sshkey, duration, fio_enable=False)
        run_fio_workload_in_vm("tdvm-fio-%d" % duration, VM_TYPE_TD, image, kernel,
            sshkey, duration)
    for duration in [60, 120, 180]:
        run_fio_workload_in_vm("legacy-vm-empty-%d" % duration, VM_TYPE_TD, image, kernel,
            sshkey, duration, fio_enable=False)
        run_fio_workload_in_vm("legacy-vm-fio-%d" % duration, VM_TYPE_EFI, image, kernel,
            sshkey, duration)