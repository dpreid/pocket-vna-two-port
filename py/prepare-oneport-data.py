#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""

Demo.py 

demonstrate scikit-rf SOLT one-port cal

@author timothy.d.drysdale@gmail.com

Created 2022-02-20

"""


import skrf as rf
from skrf.calibration import OnePort
from skrf.media import DefinedGammaZ0
import matplotlib.pyplot as plt
import numpy as np

# measured files supplied from pocket-VNA measurement
meas2port = [\
        rf.Network('test/measured/oneport/short.s2p'),
        rf.Network('test/measured/oneport/open.s2p'),
        rf.Network('test/measured/oneport/load.s2p'),
        ]
# the data we want is S11

meas1port = []
for meas in meas2port:
    meas1port.append(rf.Network(frequency=meas.frequency, s=meas.s[:,0,0], name=meas.name))


line = DefinedGammaZ0(meas1port[0].frequency)

my_ideals = [\
        line.short(),
        line.open(),
        line.load(1e-99), #noreflection Gamma -> 0 (can't be zero, div by zero error)
        ]


## create a Calibration instance
cal = OnePort(\
        ideals = my_ideals,
        measured = meas1port,
        )


## run, and apply calibration to a DUT

# run calibration algorithm
cal.run()

# apply it to a dut
dut2port = rf.Network('test/supplied/oneport/DUTuncal.s2p')
dut1port = rf.Network(frequency=dut2port.frequency, s=dut2port.s[:,0,0], name="scikit cal")
dut_caled = cal.apply_cal(dut1port)

# save results for comparison against automated implementation of this approach
dut_caled.write_touchstone('test/expected/oneport/expected.s1p')

# check results against supplied data

expected2port = rf.Network('test/supplied/oneport/DUTcal.s2p')
expected1port = rf.Network(frequency=expected2port.frequency, s=expected2port.s[:,0,0], name="matlab cal")

plt.figure()
dut_caled.plot_s_db()
expected1port.plot_s_db()
plt.savefig("img/oneport/demo-db.png",dpi=300)
plt.show()
plt.close()

plt.figure()
dut_caled.plot_s_deg()
expected1port.plot_s_deg()
plt.savefig("img/oneport/demo-deg.png",dpi=300)
plt.show()
plt.close()

plt.figure()
scdb = np.squeeze(dut_caled.s_db)
mcdb = np.squeeze(expected1port.s_db)
plt.plot(dut_caled.f, scdb-mcdb)
plt.xlabel("Frequency (Hz)")
plt.ylabel("Error (dB)")
plt.savefig("img/oneport/demo-db-error.png",dpi=300)
plt.show()
plt.close()

plt.figure()
scdeg = np.squeeze(dut_caled.s_deg)
mcdeg = np.squeeze(expected1port.s_deg)
plt.plot(dut_caled.f, scdeg-mcdeg)
plt.ylim([-180,180])
plt.xlabel("Frequency (Hz)")
plt.ylabel("Error (deg)")
plt.savefig("img/oneport/demo-deg-error.png",dpi=300)
plt.show()
plt.close()


## prep for the JSON DEMO ... do it all again, but with JSON.

# get our arrays out of the network models
f = meas1port[0].f.tolist()
mssr = np.squeeze(meas1port[0].s_re).tolist()
mssi = np.squeeze(meas1port[0].s_im).tolist()

msor = np.squeeze(meas1port[1].s_re).tolist()
msoi = np.squeeze(meas1port[1].s_im).tolist()

mslr = np.squeeze(meas1port[2].s_re).tolist()
msli = np.squeeze(meas1port[2].s_im).tolist()

msdr = np.squeeze(dut1port.s_re).tolist()
msdi = np.squeeze(dut1port.s_im).tolist()


request = {
        "cmd":"oneport",
        "freq":f,
        "short":{
            "real":mssr,
            "imag":mssi
                },
         "open":{
            "real":msor,
            "imag":msoi
                },               
         "load":{
            "real":mslr,
            "imag":msli
                },                 
         "dut":{
            "real":msdr,
            "imag":msdi
                }  
        }

import json
with open('test/json/oneport/oneport.json', 'w') as f:
    json.dump(request, f)





