#!/usr/bin/python3
# -*- coding: utf-8 -*-
from gfaaccesslib.gfa import GFA
from gfaaccesslib.api_helpers import GFAExposureStack, GFAStandardExposureBuilder, GFACCDGeom, GFAExposeMode
from gfaaccesslib.logger import log, formatter
import logging
import time
import sys
import os

__author__ = 'otger'

ch = logging.StreamHandler()
ch.setLevel(logging.DEBUG)
ch.setFormatter(formatter)
# log.addHandler(ch)

IP_EMBEDDED = '172.16.17.140'
IP_HOME = '192.168.0.164'
IP_NODHCP = '192.168.100.100'
IP_GFA_PROTO = '172.16.17.54'
IP_DEFAULT = IP_GFA_PROTO

if len(sys.argv) > 1:
    IP = sys.argv[1]
else:
    IP = os.environ.get('GFA_IP', None) or IP_DEFAULT
PORT = 32000
APORT = 32001

print("Connecting to GFA @{0}:{1}".format(IP, PORT))
log.info('Configured GFA to ip {0} - port {1}'.format(IP, PORT))

# There is no need to subscribe to async port
gfa = GFA(IP, PORT)

# We have to configure ccd geometry

# geometry values are stored at
geom = gfa.clockmanager.geom_conf

# it has default values when created
# print(geom.amplifier_cols)

# To change geometry values
# geom.amplifier_cols = 300

# if we are sure GFA has not been configured, we have to configure geometry.
gfa.clockmanager.remote_set_ccd_geom()

# if we need to recover geometry:
gfa.clockmanager.remote_get_ccd_geom()

# values recovered from GFA are automatically stored at
print(gfa.clockmanager.geom_conf)

# to set clocks timing default values:
gfa.clockmanager.remote_set_clock_timings()

# values are stored at:
print(gfa.clockmanager.time_conf)

# we can check all registers that must be configured are configured by
gfa.clockmanager.remote_get_info()
print(gfa.clockmanager.info.status.is_configured)

# Another important settings to be able to power up the system is to set the voltage values and the configuration of
# offsets and internal gains of the dac
# To set default values:
gfa.powercontroller.remote_set_dac_conf()

gfa.powercontroller.voltages.set_default_values()
gfa.powercontroller.remote_set_voltages()

# values are stored and can be changed at:
# print(gfa.powercontroller.voltages)

# to change a voltage value:
# gfa.powercontroller.voltages.DD.volts = 29.99
# gfa.powercontroller.remote_set_voltages()

# to recover current configured values
gfa.powercontroller.remote_get_configured_voltages()
print(gfa.powercontroller.voltages)

# we can know more information
# print(gfa.powercontroller.voltages.DD)

# At this point all voltages are configured, so all required settings are configured
# for both clockmanager and powercontroller. Automatically, GFA changes its status from
# idle to configured
gfa.exposecontroller.remote_get_status()
print(gfa.exposecontroller.status)

# At configured state power is not applied, all voltage enables are false:
gfa.powercontroller.remote_get_enables()
print(gfa.powercontroller.enables)

# There are extra settings like how much time between power up/down phases
# which are set by default

# print(gfa.powercontroller.powerup_timing_ms)

# but we can change:
gfa.powercontroller.powerup_timing_ms = 250
gfa.powercontroller.remote_set_phase_timing()
gfa.powercontroller.remote_get_phase_timing()

print(gfa.powercontroller.powerup_timing_ms)

# finally we can power up the system:
gfa.exposecontroller.remote_power_up()
itr = 0
while gfa.exposecontroller.status.ready_state is False:
    if itr > 10:
        gfa.powercontroller.remote_get_configured_channels()
        print(gfa.powercontroller.dac_channels)
        raise Exception("gfa should be in ready state")
    print(gfa.exposecontroller.status)
    time.sleep(0.4)
    gfa.exposecontroller.remote_get_status()
    itr += 1

print(gfa.exposecontroller.status)

# we can check if GFA is ready and check voltage enables status
if gfa.exposecontroller.status.ready_state:
    gfa.powercontroller.remote_get_enables()
    print(gfa.powercontroller.enables)

gfa.close()
