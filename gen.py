from __future__ import print_function
import numpy as np

for dt in [
        "float32", "float64",
        "int8", "int16", "int32", "int64",
        "uint8", "uint16", "uint32", "uint64",
        ]:
    for order in ["f", "c"]:
        with open("testdata/data_%s_2x3_%sorder.npz" % (dt, order), "w") as f:
            print(">>> %s" % f.name)
            arr = np.arange(6).reshape(2,3, order=order)
            np.save(f, arr)


