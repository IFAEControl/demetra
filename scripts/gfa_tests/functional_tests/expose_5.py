#!/usr/bin/python3
# -*- coding: utf-8 -*-
from gfaaccesslib.gfa import GFA
from gfaaccesslib.api_helpers import GFAExposureLock
import sys
import os

if len(sys.argv) > 1:
    IP = sys.argv[1]
else:
    IP = os.environ.get('GFA_IP', None) or IP_DEFAULT
PORT = 32000
APORT = 32001

try:
    gfa = GFA(IP, PORT, APORT)
    gfa.adccontroller.spi_write(0xf, 0x0)
    gfa.adccontroller.spi_write(0x2a, 0xcccc)
    gfa.adccontroller.adc_start_acq()
    g = gfa.clockmanager.stack

    g.clear()
    g.add_new_image_cmd()
    g.add_set_modes_cmd(True, True, True, True)
    g.add_wait_cmd(5)
    g.add_read_rows_cmd(600)
    g.add_none_cmd()

    gfa.clockmanager.remote_set_stack_contents()
    gfa.buffers.remote_set_data_provider(1, 5)
    acq_lock = GFAExposureLock()
    gfa.async.add_end_image_callback(acq_lock.async_callback_release)
    acq_lock.acquire()
    gfa.exposecontroller.remote_start_stack_exec()

    acq_lock.acquire()
    acq_lock.release()

    im_num = sorted(gfa.raws.list_images())[-1]
    im = gfa.raws.get_image(im_num)
    im.check_fake_im(5)
    gfa.raws.rem_image(im_num)

    gfa.buffers.remote_get_buffers_status()
    gfa.buffers.remote_clear_buffers()
    sys.exit(0)
except Exception as ex:
    print("There has been an exception: {0}".format(ex))
    sys.exit(1)
finally:
    gfa.close()
