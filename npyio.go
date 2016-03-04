// Package npyio provides read/write access to files following the NumPy data file format:
//  http://docs.scipy.org/doc/numpy-1.10.1/neps/npy-format.html
//
// Example:
//
//  f, err := os.Open("data.npz")
//  r, err := npyio.NewReader(f)
//  var m mat64.Dense
//  err = r.Read(&m)
//	fmt.Printf("data = %v\n", mat64.Formatted(&m, mat64.Prefix("       ")))
//
// npyio can also read data directly into slices, arrays or scalars, provided
// there is a valid type conversion [numpy-data-type]->[go-type].
//
// Example:
//  var data []float64
//  err = r.Read(&data)
//
//  var data uint64
//  err = r.Read(&data)
package npyio
