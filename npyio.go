// Package npyio provides read/write access to files following the NumPy data file format:
//  http://docs.scipy.org/doc/numpy-1.10.1/neps/npy-format.html
//
// Example:
//
//  f, err := os.Open("data.npz")
//  r, err := npyio.NewReader(f)
//  data, err := r.Read()
//  nrows, ncols := data.Dims()
//  for i := 0; i < nrows; i++ {
//    for j := 0; j < ncols; j++ {
//      fmt.Printf("data[%d][%d] = %v\n", i, j, m.At(i,j))
//    }
//  }
package npyio
