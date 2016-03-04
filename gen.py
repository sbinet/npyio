#!/usr/bin/env python2
from __future__ import print_function
import numpy as np

for dt in [
        "float32", "float64",
        "int8", "int16", "int32", "int64",
        "uint8", "uint16", "uint32", "uint64",
        ]:
    for order in ["f", "c"]:
        with open("testdata/data_%s_2x3_%sorder.npy" % (dt, order), "w") as f:
            print(">>> %s" % f.name)
            arr = np.arange(6, dtype=dt).reshape(2, 3, order=order)
            np.save(f, arr)
            pass
        
        with open("testdata/data_%s_6x1_%sorder.npy" % (dt, order), "w") as f:
            print(">>> %s" % f.name)
            arr = np.arange(6, dtype=dt).reshape(6,1, order=order)
            np.save(f, arr)
            pass

        with open("testdata/data_%s_1x1_%sorder.npy" % (dt,order), "w") as f:
            print(">>> %s" % f.name)
            arr = np.arange(1, dtype=dt).reshape(1,1, order=order)
            arr[0] = 42
            np.save(f, arr)
            pass

        with open("testdata/data_%s_scalar_%sorder.npy" % (dt,order), "w") as f:
            print(">>> %s" % f.name)
            np.save(f, getattr(np, dt)(42))
            pass

with open("testdata/data_float64_2x3x4_corder.npy", "w") as f:
    print(">>> %s" % f.name)
    arr = np.arange(2*3*4, dtype="float64").reshape(2,3,4, order="c")
    np.save(f, arr)
    pass

with open("testdata/nans_inf.npy", "w") as f:
    print(">>> %s" % f.name)
    arr = np.array([np.nan, -np.inf, 0, np.inf], dtype="float64", order="c")
    np.save(f, arr)
    pass
