#!/usr/bin/python3
# -*- coding: utf-8 -*-
from gfaaccesslib.gfa import GFA
from gfaaccesslib.api_helpers import GFAExposureLock
from gfaaccesslib.logger import log, formatter
import time
import sys
import os

__author__ = 'otger'

IP_GFA_PROTO = '172.16.17.54'
IP_DEFAULT = IP_GFA_PROTO

if len(sys.argv) > 1:
    IP = sys.argv[1]
else:
    IP = os.environ.get('GFA_IP', None) or IP_DEFAULT
PORT = 32000
APORT = 32001

gfa = GFA(IP, PORT, APORT)
try:
    gfa.adccontroller.spi_write(0xf, 0x0)
    gfa.adccontroller.spi_write(0x2a, 0xcccc)
    gfa.adccontroller.adc_start_acq()
    g = gfa.clockmanager.stack

    g.clear()
    g.add_new_exposure_cmd()
    g.add_set_modes_cmd(True, True, True, True)
    g.add_wait_cmd(500)
    g.add_dump_rows_cmd(400)
    g.add_read_rows_cmd(100)
    g.add_none_cmd()

    gfa.clockmanager.remote_set_stack_contents()

    gfa.buffers.remote_set_data_provider(1, 6)
    acq_lock = GFAExposureLock()
    gfa.async.add_end_exposure_callback(acq_lock.async_callback_release)
    acq_lock.acquire()
    gfa.exposecontroller.remote_expose()

    acq_lock.acquire()
    acq_lock.release()

    im_num = sorted(gfa.raws.list_images())[-1]
    gfa.raws.rem_image(im_num)

    gfa.buffers.remote_get_buffers_status()
    print(gfa.buffers.status)
    gfa.buffers.remote_clear_buffers()

except Exception as ex:
    log.exception("Exception")
finally:
    gfa.close()
